/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"path"
	"time"

	"github.com/aldesgroup/aldev/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/cobra"
)

// ----------------------------------------------------------------------------
// Root command execution
// ----------------------------------------------------------------------------

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the AldevCmd.
func Execute() {
	err := aldevCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

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

var (
	// flags
	cfgFileName       string
	verbose           bool
	useLocalDeps      bool
	disableGeneration bool
)

func init() {
	aldevCmd.PersistentFlags().StringVarP(&cfgFileName, "file", "f", ".aldev.yaml", "aldev config file")
	aldevCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "activates debug logging")
	aldevCmd.Flags().BoolVarP(&useLocalDeps, "use-local-deps", "l", false,
		"to use the local dependencies declared in the config file")
	aldevCmd.PersistentFlags().BoolVarP(&disableGeneration, "disable-generation", "d", false, "disable the generation of all the config files, but not code generation")
}

// ----------------------------------------------------------------------------
// Public accesses
// ----------------------------------------------------------------------------

func GetAldevCmd() *cobra.Command {
	return aldevCmd
}

func GetConfigFilename() string {
	return cfgFileName
}

// ----------------------------------------------------------------------------
// Main logic watching for some files, and (re)starting the building &
// deploying function
// ----------------------------------------------------------------------------

func aldevRun(command *cobra.Command, args []string) {
	// it's only here that we have this variable valued
	if verbose {
		utils.SetVerbose()
	}

	// also valueing here, since the source of truth must lie in the utils package
	if useLocalDeps {
		utils.SetUseLocalDeps()
	}

	// reading the Aldev config one first time
	cfg := utils.ReadConfig(cfgFileName)

	// the main cancelable context, that should stop everything
	aldevCtx := utils.InitAldevContext(2000, nil)

	// --- one-time stuff

	// one time thing: install the pre-commit hook
	go utils.InstallGitHooks(aldevCtx, cfg)

	// one time thing: using Aldev swap when locally developping the dependencies alongside
	if useLocalDeps {
		go utils.Run("Allowing HMR to work even with dependencies", aldevCtx, true, "aldev swap")
	}

	// --- main loop stuff

	// for which file changes are we going to restart the main loop?
	watched := []string{cfgFileName} // Aldev's config
	if cfg.API != nil {
		watched = append(watched, path.Join(cfg.GetSrcDir(), cfg.GetConfigPath())) // the API or lib's config
	}

	// adding a watcher to detect some file changes
	watcher := utils.WatcherFor(watched...)
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

						// cancelling the current loop context and restarting it
						aldevCtx.RestartLoop()

						// Waiting a bit for the previous execution to stop gracefully
						time.Sleep(1000 * time.Millisecond) // todo wait for the clean up to be finished

						// Restarting the main building / local deployment function
						utils.Step("Restarting the main function")

						// go rebuilding & deploying the app again
						go asyncPrepareAndRun(aldevCtx.GetLoopCtx())

						// Waiting a bit for the new execution to start sufficiently
						time.Sleep(1000 * time.Millisecond)
					}
				}

			case errWatcher := <-watcher.Errors:
				utils.FatalIfErr(aldevCtx, errWatcher)
			}
		}
	}()

	// building & deploying the app
	go asyncPrepareAndRun(aldevCtx.GetLoopCtx())

	// not quitting while the context is still going
	<-aldevCtx.Done()

	time.Sleep(10 * time.Millisecond)
}

// ----------------------------------------------------------------------------
// Building & deploying the app once the API & Aldev's configs are settled
// ----------------------------------------------------------------------------

// building & deploying the app
func asyncPrepareAndRun(ctx utils.CancelableContext) {
	// making sure we recover any big crashing error
	defer utils.Recover(ctx, "building & deploying the app")

	// reading the Aldev config again, in case it has changed
	cfg := utils.ReadConfig(cfgFileName)

	// proceed to download the needed external resources
	utils.DownloadExternalResources(ctx, cfg)

	// in library mode, there no need for k8s, deployments, env vars, etc.
	if cfg.IsLibrary() {
		utils.QuickRun("Installing / refreshing the dev environment", cfg.Lib.Install)
		utils.Run("Developing the lib", ctx, true, cfg.Lib.Develop)

		// Wait for the context to be canceled or the program to exit
		<-ctx.Done()

	} else {

		// Generating config files for deploying the app locally, CI / CD, etc.
		if !disableGeneration {
			utils.GenerateConfigs(cfg)
		}

		// Ready for launch
		utils.Launch(ctx, cfg)
	}
}
