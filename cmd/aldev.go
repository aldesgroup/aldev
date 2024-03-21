/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/aldesgroup/aldev/templates"
	"github.com/aldesgroup/aldev/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/cobra"
)

// ----------------------------------------------------------------------------
// Command declaration
// ----------------------------------------------------------------------------

// aldevCmd represents the base command when called without any subcommands
var aldevCmd = &cobra.Command{
	Use:   "aldev",
	Short: "Quick dev with Goald, GoaldR & Kubernetes",
	Long: "Run Aldev to start or continue developing a Goald / GoaldR application " +
		"with automatic deployment in a local k8s cluster and live reloading.",
	Run: aldevRun,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the AldevCmd.
func Execute() {
	err := aldevCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var (
	cfgFileName  string
	verbose      bool
	webLocalDeps []string
)

func init() {
	aldevCmd.PersistentFlags().StringVarP(&cfgFileName, "file", "f", ".aldev.yaml", "aldev config file")
	aldevCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "activates debug logging")
	aldevCmd.LocalFlags().StringSliceVarP(&webLocalDeps, "use-local", "l", []string{},
		"the web dependencies to use locally rather than their released versions")
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

	// reading the Aldev config one first time
	cfg := utils.ReadConfig(cfgFileName)

	// the main cancelable context, that should stop everything
	mainCtx := utils.InitMainContext()

	// a context for the loop below, and a cancellation function to get out of it
	loopCtx, cancelLoopFn := context.WithCancel(context.Background())

	// allowing to stop the loop from this main function
	mainCtx.SetCancelLoopFn(cancelLoopFn)

	// adding a watcher to detect some file changes
	watcher := watcherFor(cfg.API.Config, cfgFileName)
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
// Building & deploying the app once the API & Aldev's configs are settled
// ----------------------------------------------------------------------------

// building & deploying the app
func asyncBuildAndDeploy(ctx context.Context) {
	// making sure we clean up at the end
	defer func() {
		time.Sleep(100 * time.Millisecond)
		utils.Info("We'll clean up the context now")
		utils.Run("tilt down")
	}()

	// making sure we recover any big crashing error
	defer utils.Recover("building & deploying the app")

	// reading the Aldev config again, in case it has changed
	cfg := utils.ReadConfig(cfgFileName)

	// making sure the config map is here and up-to-date
	utils.EnsureConfigmap(cfg)

	// making sure the right namespace exists
	utils.EnsureNamespace(cfg)

	// making sure some needed files are here
	utils.EnsureFileFromTemplate(cfg, cfg.Deploying.Dir+"/"+cfg.AppName+"-app.yaml", templates.AppLocal)
	utils.EnsureFileFromTemplate(cfg, cfg.Deploying.Dir+"/"+cfg.AppName+"-docker-local-api", templates.DockerLocalAPI)
	utils.EnsureFileFromTemplate(cfg, "Tiltfile", templates.Tiltfile)

	// making sure the namespace is fresh
	if string(utils.RunAndGet(utils.Fatal, "kubectl get all --namespace %s-local", cfg.AppName)) != "" {
		utils.Info("The namespace needs some cleanup first")
		utils.Run("tilt down")
	}

	// Running a command that never finish with the cancelable context
	options := ""
	if len(cfg.Web.UseLocal) > 0 {
		options = " --use-local "
		options = options + strings.Join(cfg.Web.UseLocal, options)
	}
	if options != "" {
		options = " --" + options
	}
	utils.RunWithCtx(ctx, "tilt up --stream%s", options)

	// Wait for the context to be canceled or the program to exit
	<-ctx.Done()
}

// ----------------------------------------------------------------------------
// Command utils
// ----------------------------------------------------------------------------

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
