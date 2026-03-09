// ----------------------------------------------------------------------------
// The code here is about local deployment with Containers
// ----------------------------------------------------------------------------
package utils

import (
	"os"
	"path"
	"time"

	core "github.com/aldesgroup/corego"
	"github.com/fsnotify/fsnotify"
	"github.com/patrickmn/go-cache"
)

type exclusionType int

const exclusionTypeEXCLUDExONE = 1
const exclusionTypeEXCLUDExALL = 2

var excludedPaths map[string]exclusionType

// this function allows to us to continuously develop our Go source, weither it's for an API, or a library
// this means : rebuilding it every time it's changed, and also running the needed codegen
func RunGoSrcDev(ctx CancelableContext) {
	// making sure the local env is ready for running the Go app
	ensureLocalEnvReady()

	// the paths we don't want to we watched
	excludedPaths = map[string]exclusionType{
		GetGoSrcDir(): exclusionTypeEXCLUDExONE, // not including the API folder itself, because of the conf file and go.sum
		"_include":    exclusionTypeEXCLUDExALL, // obviously not trigering codegen / rebuild on codegen'd files, otherwise: infinite loop
		"class":       exclusionTypeEXCLUDExALL, // obviously not trigering codegen / rebuild on other codegen'd files, otherwise: infinite loop
		".git":        exclusionTypeEXCLUDExALL, // not looking into a .git folder
		"bin":         exclusionTypeEXCLUDExALL, // also obviously not trigering on the binaries
	}

	// the root paths to watch for changes
	rootPaths := append(GetGoAdditionalWatchedPaths(), GetGoSrcDir())

	// which files are going to be impacted?
	watchedFolders := getWatchedFolders(rootPaths...)

	// performing the initial build & run
	go devUp()

	// adding a watcher to detect some file changes, for additional needed swaps
	watcher := WatcherFor(watchedFolders...)

	// making sure we'll roll the changes back at the end
	defer func() {
		// let's stop the watching right away
		core.PanicIfErr(watcher.Close())

		// TODO like in Aldev, wait for this to be done instead of sleeping
		time.Sleep(10 * time.Millisecond)

		// removing the currently running stuff
		devDown()
	}()

	// watching all the files here and rebooting the watching if something is changed
	// so as to handle new files, or new imports in existing files for instance
	go func() {
		// capturing ghost changes on files
		cache := cache.New(3*time.Second, time.Minute)

		// listening for events
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					if _, alreadyCaptured := cache.Get(event.String()); !alreadyCaptured {
						// let's wait a bit first
						time.Sleep(100 * time.Millisecond)

						Debug("/!\\ File modified: %s (event = %s)", event.Name, event.String())

						// caching to prevent stuttering
						cache.SetDefault(event.String(), true)

						// which files are going to be impacted NOW?
						watchedFolders = getWatchedFolders(rootPaths...)

						// adding a watcher to detect some file changes
						core.PanicIfErr(watcher.Close()) // closing the old one
						watcher = WatcherFor(watchedFolders...)

						// let's wait a bit here than the FS has finished doing it's stuff
						time.Sleep(200 * time.Millisecond)

						// restarting the API
						devDown()
						go devUp()
					}
				}

			case errWatcher := <-watcher.Errors:
				core.PanicIfErr(errWatcher)
			}
		}
	}()

	// not quitting while the context is still going
	<-ctx.Done()
}

// getting all the folders and files to watch for the API
func getWatchedFolders(givenPaths ...string) (watchedFolders []string) {
	// iterating over all the given folders
	for _, givenPath := range givenPaths {
		// add this folder to watch it,
		if core.DirExists(givenPath) {
			// but only if it's not completely excluded
			if exclType := excludedPaths[path.Base(givenPath)]; exclType != exclusionTypeEXCLUDExALL {
				// adding the current path, only if it's not excluded
				if exclType != exclusionTypeEXCLUDExONE {
					watchedFolders = append(watchedFolders, givenPath)
				}

				// adding subfolders
				for _, entry := range core.EnsureReadDir(givenPath) {
					watchedFolders = append(watchedFolders, getWatchedFolders(path.Join(givenPath, entry.Name()))...)
				}
			}
		}
	}

	return
}

func devUp() {
	// building the API and code-generating the missing stuff
	codeGenCtx := NewBaseContext().WithStdErrWriter(os.Stdout).WithStdOutWriter(os.Stdout).WithAllowFailure(true)
	if Run("Building & code-generating", codeGenCtx, false, "aldev codegen %s", core.IfThenElse(verbose, "-v", "")) && IsDevAPI() {

		// locally deploying the API with 3 instances
		QuickRun("Starting the API", "podman-compose -f %s/local/compose.yaml up --scale %s_api=%d",
			Config().Deploying.Dir, Config().AppNameShort, Config().API.LocalInstances)
	}
}

func devDown() {
	// // Nuking everything launched with Podman... That may be a little bit too much
	// // We'll prolly have to smooth that out sometimes later
	if IsDevAPI() {
		QuickRun("Stopping the API", "podman rm --force --filter name=local_%s_", Config().AppNameShort)
	}

	// // Also, making sure Podman's internal network is removed to be able to start from fresh later on
	// QuickRun("Stopping the API (2/2)", "%s", "podman network rm local_default")
}

func ensureLocalEnvReady() {
	if IsDevAPI() {
		localEnvCtx := NewBaseContext().WithStdErrWriter(os.Stdout).WithStdOutWriter(os.Stdout).WithAllowFailure(true)
		if !Run("Checking the 'shared-net' network existence", localEnvCtx, false, "%s", "podman network exists shared-net") {
			Run("Creating the 'shared-net' network", localEnvCtx, false, "%s", "podman network create shared-net")
		}
	}
}
