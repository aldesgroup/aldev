package utils

import (
	core "github.com/aldesgroup/corego"
	"github.com/fsnotify/fsnotify"
)

func WatcherFor(filepaths ...string) *fsnotify.Watcher {
	// new watcher
	watcher, errNew := fsnotify.NewWatcher()
	core.PanicIfErr(errNew)

	// watching the given files
	for _, filepath := range filepaths {
		Debug("Watching path: %s", filepath)
		core.PanicIfErr(watcher.Add(filepath))
	}

	return watcher
}
