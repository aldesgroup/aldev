package codeswap

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aldesgroup/aldev/cmd"
	"github.com/aldesgroup/aldev/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/cobra"
)

// ----------------------------------------------------------------------------
// Command declaration
// ----------------------------------------------------------------------------

// aldevSwapCmd represents a subcommand
var aldevSwapCmd = &cobra.Command{
	Use:   "codeswap",
	Short: "Swaps bit of code - like import paths - in targeted files",
	Long: "This performs swaps of code in order, for example, to work locally more efficiently," +
		" by allowing to replace paths to vendor libraries by paths to local git projects",
	Run: aldevSwapRun,
}

var (
	folders        map[string]bool
	done           map[string]bool
	sets           []*swapSet
	watchedFolders []string
)

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevSwapCmd)
}

// ----------------------------------------------------------------------------
// Main logic
// ----------------------------------------------------------------------------

var finished = false
var finishedMx = new(sync.Mutex)

func setFinished() {
	finishedMx.Lock()
	finished = true
	utils.Debug("We're finished!")
	finishedMx.Unlock()
}

func isFinished() bool {
	finishedMx.Lock()
	defer finishedMx.Unlock()
	utils.Debug("Are we finished ? %t", finished)
	return finished
}

func aldevSwapRun(command *cobra.Command, args []string) {
	// Reading this command's arguments, and reading the aldev YAML config file
	cmd.ReadCommonArgsAndConfig()

	// the main cancelable context, that should stop everything
	aldevCtx := utils.InitAldevContext(10, setFinished)

	// which files are going to be impacted?
	sets, watchedFolders = getWatchedFilesAndFolders(aldevCtx)

	// performing the initial swaps
	doAllTheSwaps(aldevCtx, false, true)

	// adding a watcher to detect some file changes, for additional needed swaps
	watcher := utils.WatcherFor(watchedFolders...)

	// making sure we'll roll the changes back at the end
	defer func() {
		// let's stop the watching right away
		utils.FatalIfErr(aldevCtx, watcher.Close())

		// TODO like in Aldev, wait for this to be done instead of sleeping
		time.Sleep(10 * time.Millisecond)

		// performing the swaps, in reverse
		doAllTheSwaps(aldevCtx, true, true)
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

						utils.Debug("/!\\ File modified: %s (event = %s)", event.Name, event.String())

						// caching to prevent stuttering
						cache.SetDefault(event.String(), true)

						// which files are going to be impacted NOW?
						sets, watchedFolders = getWatchedFilesAndFolders(aldevCtx)

						// adding a watcher to detect some file changes
						utils.FatalIfErr(aldevCtx, watcher.Close()) // closing the old one
						watcher = utils.WatcherFor(watchedFolders...)

						// let's wait a bit here than the FS has finished doing it's stuff
						time.Sleep(200 * time.Millisecond)

						// performing the swaps on the newly computed sets
						doAllTheSwaps(aldevCtx, false, false)
					}
				}

			case errWatcher := <-watcher.Errors:
				utils.FatalIfErr(aldevCtx, errWatcher)
			}
		}
	}()

	// not quitting while the context is still going
	<-aldevCtx.Done()
}

func doAllTheSwaps(ctx utils.CancelableContext, rollback bool, startOrFinish bool) {
	// we're not allowing forward swaps if we're finished, only rollbacks
	if isFinished() && !rollback {
		return
	}

	start := time.Now()
	for _, set := range sets {
		set.doSwaps(ctx, rollback)
	}

	if startOrFinish {
		if rollback {
			utils.Info("Swap Mode OFF")
		} else {
			utils.Info("Swap Mode ON")
		}
	}

	// bit of logging
	utils.Info("All the code swapping done in %s", time.Since(start))

	// waiting a bit here in order to prevent the watcher to detect the changes done here
	time.Sleep(50 * time.Millisecond)

	// syncing the Go.sum file with the swaps done
	goCodeCtx := utils.InitAldevContext(100, nil).WithExecDir(utils.GetGoSrcDir())
	utils.Run("Making sure the Go.sum file is synced", goCodeCtx, false, "go mod tidy")
}

// a set associates a swap config, and the files that should be modified according to it
type swapSet struct {
	swapConf *utils.CodeSwapsConfig
	files    []string
}

// builds all the sets for all the swap configs configured
func getWatchedFilesAndFolders(ctx utils.CancelableContext) (sets []*swapSet, watchedFolders []string) {
	folders = map[string]bool{}
	done = map[string]bool{}

	for _, swapConf := range utils.Config().CodeSwaps {
		utils.Debug("--- Swap set: from '%s' for %s file(s)", swapConf.From, strings.Join(swapConf.For, ", "))
		sets = append(sets, (&swapSet{swapConf: swapConf}).buildFrom(ctx, swapConf.From))
	}

	for folder := range folders {
		watchedFolders = append(watchedFolders, folder)
	}

	sort.Strings(watchedFolders)

	return
}

// gathering all the files corresponding to the same swap config
func (thisSet *swapSet) buildFrom(ctx utils.CancelableContext, dir string) *swapSet {
	entries, errDir := os.ReadDir(dir)
	utils.FatalIfErr(ctx, errDir)

	for _, entry := range entries {
		filename := path.Join(dir, entry.Name())
		if entry.IsDir() {
			if entry.Name() != "node_modules" && entry.Name() != ".git" && entry.Name() != "dist" && entry.Name() != "vendor" {
				initialNbFolders := len(folders)
				thisSet.buildFrom(ctx, filename)
				if len(folders) > initialNbFolders {
					folders[dir] = true
				}
			}
		} else {
			for _, targetPath := range thisSet.swapConf.For {
				matched, _ := filepath.Match(targetPath, entry.Name())
				if matched && !done[filename] {
					thisSet.files = append(thisSet.files, filename)
					folders[dir] = true
					done[filename] = true
					utils.Debug("Will be watching file: %s", filename)
				}
			}
		}
	}

	return thisSet
}

const inlineComment = " /* " + utils.TagHOTSWAPPED + " do not commit! */"
const eofComment = " // " + utils.TagHOTSWAPPED + " do not commit!"

// writing all the swaps for the files of the given set
func (thisSet *swapSet) doSwaps(ctx utils.CancelableContext, rollback bool) {
	// performing the swaps for all the files of this set
	for _, filename := range thisSet.files {
		utils.Debug("Checking for swaps to do in file: %s", filename)
		// reading the current file
		contentBytes, errRead := os.ReadFile(filename)
		utils.FatalIfErr(ctx, errRead)
		contentString := string(contentBytes)

		// the text obtained at the end
		modifiedText := contentString

		// performing all the needed swaps
		for _, swap := range thisSet.swapConf.Do {
			comment := inlineComment
			if swap.EOFCom {
				comment = eofComment
			}
			if !rollback { // swapping
				modifiedText = strings.ReplaceAll(modifiedText, swap.Replace, swap.With+comment)
			} else { // swapping back
				modifiedText = strings.ReplaceAll(modifiedText, swap.With+comment, swap.Replace)
			}
		}

		// writing out the result, if there's any change
		if modifiedText != contentString {
			direction := "forward"
			if rollback {
				direction = "reverse"
			}
			utils.Info("File %s is being %s-swapped", filename, direction)
			utils.WriteStringToFile(ctx, filename, "%s", modifiedText)
		}
	}
}
