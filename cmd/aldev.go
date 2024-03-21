/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"html/template"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aldesgroup/aldev/templates"
	"github.com/aldesgroup/aldev/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ----------------------------------------------------------------------------
// Command declaration
// ----------------------------------------------------------------------------

// aldevCmd represents the base command when called without any subcommands
var aldevCmd = &cobra.Command{
	Use:   "aldev",
	Short: "Quick dev with Goald, GoaldR & Kubernetes",
	Long: "Run aldev to start or continue developing a Goald / GoaldR application " +
		"with automatic deployment in a local k8s cluster and live reloading.",
	Run: aldevRun,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the aldevCmd.
func Execute() {
	err := aldevCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	cfgFileName string
	verbose     bool
)

func init() {
	aldevCmd.PersistentFlags().StringVarP(&cfgFileName, "file", "f", ".aldev.yaml", "aldev config file")
	aldevCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "activates debug logging")
}

// ----------------------------------------------------------------------------
// Configuration
// ----------------------------------------------------------------------------

type aldevConfig struct {
	AppName string // the name of the app - beware: the key has to be "appname" in the YAML file
	API     struct {
		Config string // the path to the config file for the API
	}
	Deploying struct { // Section for the local deployment of the app
		Dir string // where all the deploying config should be
	}
}

func readConfig() *aldevConfig {
	utils.Debug("Reading aldev config")

	cfg := &aldevConfig{}

	// Reading the config file into bytes
	yamlBytes, errRead := os.ReadFile(cfgFileName)
	utils.FatalIfErr(errRead)

	// Unmarshalling the YAML file
	utils.FatalIfErr(yaml.Unmarshal(yamlBytes, cfg))

	return cfg
}

// ----------------------------------------------------------------------------
// Main logic watching for some files, and (re)starting the building &
// deploying function
// ----------------------------------------------------------------------------

func aldevRun(cmd *cobra.Command, args []string) {
	// it's only here that we have this variable valued
	if verbose {
		utils.SetVerbose()
	}

	// reading the aldev config one first time
	cfg := readConfig()

	// the main cancelable context, that should stop everything
	mainCtx := initMainContext()

	// a context for the loop below, and a cancellation function to get out of it
	loopCtx, cancelLoopFn := context.WithCancel(context.Background())

	// allowing to stop the loop from this main function
	mainCtx.SetCancelLoopFn(cancelLoopFn)

	// adding a watcher to detect some file changes
	watcher := watcherFor(cfg.API.Config)
	defer watcher.Close()

	// loop to react to these file changes
	go func() {
		// capturing ghost changes on files
		cache := cache.New(3*time.Second, time.Minute)

		// listening for events
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					if _, alreadyCaptured := cache.Get(event.String()); !alreadyCaptured {
						utils.Step("/!\\ File modified: %s (event = %s)", event.Name, event.String())

						// caching to prevent stuttering
						cache.SetDefault(event.String(), true)

						// Cancel the previous execution of the function
						cancelLoopFn()

						// Recreate the context
						loopCtx, cancelLoopFn = context.WithCancel(context.Background())

						// allowing to stop the loop from this main function
						mainCtx.SetCancelLoopFn(cancelLoopFn)

						// Waiting a bit for the previous execution to stop gracefully
						time.Sleep(1000 * time.Millisecond) // todo wait for the clean up to be finished

						// Restarting the main building / local deployment function
						utils.Step("Restarting the main function")

						// go rebuilding & deploying the app again
						go asyncBuildAndDeploy(loopCtx)

						// Waiting a bit for the new execution to start sufficiently
						time.Sleep(1000 * time.Millisecond)
					}
				}

			case errWatcher := <-watcher.Errors:
				utils.FatalIfErr(errWatcher)
			}
		}
	}()

	// building & deploying the app
	go asyncBuildAndDeploy(loopCtx)

	// not quitting while the context is still going
	<-mainCtx.Done()
}

// ----------------------------------------------------------------------------
// Building & deploying the app once the API & aldev's configs are settled
// ----------------------------------------------------------------------------

// building & deploying the app
func asyncBuildAndDeploy(ctx context.Context) {
	// making sure we clean up at the end
	defer cleanUp()

	// making sure we recover any big crashing error
	defer utils.Recover("building & deploying the app")

	// reading the config again, in case it has changed
	cfg := readConfig()

	// making sure the config map is here and up-to-date
	ensureConfigmap(cfg)

	// making sure the right namespace exists
	ensureNamespace(cfg)

	// making sure some needed files are here
	ensureFileFromTemplate(cfg, cfg.Deploying.Dir+"/"+cfg.AppName+"-app.yaml", templates.K8sLocal)
	ensureFileFromTemplate(cfg, cfg.Deploying.Dir+"/"+cfg.AppName+"-docker-local-api", templates.DockerLocalAPI)
	ensureFileFromTemplate(cfg, "Tiltfile", templates.Tiltfile)

	// Running a command that never finish with the cancelable context
	utils.RunWithCtx(ctx, "tilt up --stream")

	// Wait for the context to be canceled or the program to exit
	<-ctx.Done()
}

// ----------------------------------------------------------------------------
// Command utils
// ----------------------------------------------------------------------------

func cleanUp() {
	time.Sleep(100 * time.Millisecond)
	utils.Info("We'll clean up the context now")
	utils.Run("tilt down")
}

type mainContext struct {
	ctx          context.Context
	mainCancelFn func()
	cancelLoopFn func()
}

func (thisCtx *mainContext) SetCancelLoopFn(cancelLoopFn func()) {
	thisCtx.cancelLoopFn = cancelLoopFn
}

func (thisCtx *mainContext) cancelAll() {
	thisCtx.cancelLoopFn()
	utils.Info("Waiting for some cleanup...")
	time.Sleep(3500 * time.Millisecond) // TODO waiting
	thisCtx.mainCancelFn()
}

func (thisCtx *mainContext) Done() <-chan struct{} {
	return thisCtx.ctx.Done()
}

func initMainContext() *mainContext {
	mainContext := &mainContext{}

	// Initialize a cancelable context
	mainContext.ctx, mainContext.mainCancelFn = context.WithCancel(context.Background())

	// Initialize channel to receive signals
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	// Start a goroutine to handle signals
	go func() {
		sig := <-signalCh
		utils.Warn("Received signal: %v", sig)

		// Cancel the context on signal received
		mainContext.cancelAll()
	}()

	return mainContext
}

func watcherFor(filepaths ...string) *fsnotify.Watcher {
	// new watcher
	watcher, errNew := fsnotify.NewWatcher()
	utils.FatalIfErr(errNew)

	// watching the given files
	for _, filepath := range filepaths {
		utils.Info("Watching file: %s", filepath)
		utils.FatalIfErr(watcher.Add(filepath))
	}

	return watcher
}

func ensureConfigmap(cfg *aldevConfig) {
	utils.Debug("Making sure the configmap is up-to-date")

	// some controls first
	if cfg.Deploying.Dir == "" {
		utils.Fatal("Empty 'deploying.dir' config!")
	}
	configFile, errStat := os.Stat(cfg.API.Config)
	utils.FatalIfErr(errStat)

	// (re)init the file
	configMapFilename := cfg.Deploying.Dir + "/" + cfg.AppName + "-cm.yaml"
	utils.EnsureDir(cfg.Deploying.Dir)
	utils.WriteToFile(configMapFilename, "# generated from app-api/config.yaml by aldev")

	// creating the config map
	cmd := "kubectl create configmap %s-configmap" // creating a configmap object here
	cmd += " -n %s-local -o yaml"                  // not forgetting the namespace here, and we want a YAML output...
	cmd += " --dry-run=client --from-file=%s"      // ... so we dry-run this, from the config file found in the API sources
	fileContentBytes := utils.RunAndGet(utils.Fatal, cmd, cfg.AppName, cfg.AppName, cfg.API.Config)

	// tweaking it
	fileContent := string(fileContentBytes)
	fileContent = strings.Replace(fileContent, "creationTimestamp: null", "creationTimestamp: \"%s\"", 1)

	// outputting it
	utils.WriteToFile(configMapFilename, fileContent, configFile.ModTime().Format("2006-01-02T15:04:05Z"))
}

func ensureNamespace(cfg *aldevConfig) {
	utils.Debug("Making sure the correct namespace exists")

	if string(utils.RunAndGet(utils.Noop, "kubectl get namespace %s-local", cfg.AppName)) == "" {
		utils.Run("kubectl create namespace %s-local", cfg.AppName)
	}
}

func ensureFileFromTemplate(cfg *aldevConfig, filepath, tpl string) {
	utils.Debug("Making sure this file exists: %s", filepath)

	// Create a new template
	tmpl, errTpl := template.New(filepath).Parse(tpl)
	utils.FatalIfErr(errTpl)

	// Create a new file to write the result
	outputFile, errCreate := os.Create(filepath)
	utils.FatalIfErr(errCreate)
	defer outputFile.Close()

	// Execute the template with the data
	utils.FatalIfErr(tmpl.Execute(outputFile, cfg))
}
