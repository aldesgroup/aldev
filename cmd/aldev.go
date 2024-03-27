/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
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
	useLocalDeps bool
	options      string
)

func init() {
	aldevCmd.PersistentFlags().StringVarP(&cfgFileName, "file", "f", ".aldev.yaml", "aldev config file")
	aldevCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "activates debug logging")
	aldevCmd.PersistentFlags().BoolVarP(&useLocalDeps, "use-local-deps", "u", false,
		"to use the local dependencies declared in the config file")
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
	// making sure we recover any big crashing error
	defer utils.Recover("building & deploying the app")

	// reading the Aldev config again, in case it has changed
	cfg := utils.ReadConfig(cfgFileName)

	// computing the custom options
	if useLocalDeps && len(cfg.Web.LocalDeps) > 0 {
		options = " --use-local "
		options = options + strings.Join(cfg.Web.LocalDeps, options)
	}
	if options != "" {
		options = " --" + options
	}

	// making sure we clean up at the end
	defer func() {
		time.Sleep(100 * time.Millisecond)
		utils.Run("We'll clean up the context now", "tilt down%s", options)
	}()

	// making sure the config map is here and up-to-date
	utils.EnsureConfigmap(cfg)

	// making sure some needed files are here: base local deployment
	baseDir := utils.EnsureDir(cfg.Deploying.Dir, "base")
	utils.EnsureFileFromTemplate(cfg, path.Join(baseDir, "kustomization.yaml"), templates.KustomizationBase)
	utils.EnsureFileFromTemplate(cfg, path.Join(baseDir, cfg.AppName+"-api-.yaml"), templates.API)
	utils.EnsureFileFromTemplate(cfg, path.Join(baseDir, cfg.AppName+"-api-lb.yaml"), templates.LB)
	utils.EnsureFileFromTemplate(cfg, path.Join(baseDir, cfg.AppName+"-web.yaml"), templates.Web)

	// docker files
	dockerDir := utils.EnsureDir(cfg.Deploying.Dir, "docker")
	utils.EnsureFileFromTemplate(cfg, path.Join(dockerDir, cfg.AppName+"-local-api-docker"), templates.DockerLocalAPI)
	utils.EnsureFileFromTemplate(cfg, path.Join(dockerDir, cfg.AppName+"-local-web-docker"), templates.DockerLocalWEB)
	utils.EnsureFileFromTemplate(cfg, path.Join(dockerDir, cfg.AppName+"-remote-api-docker"), templates.DockerRemoteAPI)
	utils.EnsureFileFromTemplate(cfg, path.Join(dockerDir, cfg.AppName+"-remote-web-docker"), templates.DockerRemoteWeb)

	// adding overlays
	overlaysDir := utils.EnsureDir(cfg.Deploying.Dir, "overlays")
	addOverlay(cfg, overlaysDir, "dev", nil)
	addOverlay(cfg, overlaysDir, "local", [][]string{
		{"patch-no-web-container.yaml", templates.NoWebContainerPatch},
	})
	addOverlay(cfg, overlaysDir, "sandbox", nil)
	addOverlay(cfg, overlaysDir, "staging", nil)
	addOverlay(cfg, overlaysDir, "production", nil)
	// utils.EnsureFileFromTemplate(cfg, path.Join(localDepsDir, "patch-no-web-container.yaml"), templates.NoWebContainerPatch)

	// deployment with Gitlab
	utils.EnsureFileFromTemplate(cfg, ".gitlab-ci.yml", templates.GitlabCI)

	// last but not least, the Tiltfile
	utils.EnsureFileFromTemplate(cfg, "Tiltfile", templates.Tiltfile)

	// panic("stopping here for now")

	// making sure the namespace is fresh
	kustomization := "dev"
	if useLocalDeps {
		kustomization = "local"
	}
	if string(utils.RunAndGet("We want to check what's in our namespace",
		utils.Fatal, "kubectl get all --namespace %s-%s", cfg.AppName, kustomization)) != "" {
		utils.Run("The namespace needs some cleanup first", "tilt down%s", options)
	}

	// Running a command that never finish with the cancelable context
	mode := ""
	if verbose {
		mode = " --verbose --debug"
	}
	utils.RunWithCtx("Now we start Tilt to handle all the k8s deployments",
		ctx, "tilt up%s --stream%s", mode, options)

	// Wait for the context to be canceled or the program to exit
	<-ctx.Done()
}

// ----------------------------------------------------------------------------
// Command utils
// ----------------------------------------------------------------------------

// adding an overlay with its name; each patch should be at least: [0]: the filename, [1]: the template;
// [2], [3], etc, are string format parameters to fill the "%s" placeholders in the template.
func addOverlay(cfg *utils.AldevConfig, overlaysDir, overlayName string, patches [][]string) {
	overlay := utils.EnsureDir(overlaysDir, overlayName)

	// handling the patches at first
	kustomizationPatches := ""
	if patches != nil {
		kustomizationPatches = "\n" + "patches:"
		for _, patch := range patches {
			// adding the patch to the kustomization file
			if len(patch) < 2 {
				utils.Fatal("Patches should be provided as at least 1 filename, and 1 template")
			}
			filename := patch[0]
			template := patch[1]
			kustomizationPatches += "\n" + "  - path: " + filename

			// adding the file, from a template, with potential extra params
			templateParams := []any{}
			for i := 2; i < len(patch); i++ {
				templateParams = append(templateParams, patch[i])
			}
			utils.EnsureFileFromTemplate(cfg, path.Join(overlay, filename), template, templateParams...)
		}
	}

	// writing out the kustomization file, with its namespace resource
	utils.EnsureFileFromTemplate(cfg, path.Join(overlay, "kustomization.yaml"),
		templates.KustomizationOverlay+kustomizationPatches, overlayName, overlayName)
	utils.EnsureFileFromTemplate(cfg, path.Join(overlay, fmt.Sprintf("namespace-%s.yaml", overlayName)),
		templates.NewNamespace, overlayName)

	// utils.EnsureFileFromTemplate(cfg, path.Join(overlay, "patch-no-web-container.yaml"), templates.NoWebContainerPatch)
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
