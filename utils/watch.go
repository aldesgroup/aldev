package utils

import "github.com/fsnotify/fsnotify"

func WatcherFor(filepaths ...string) *fsnotify.Watcher {
	// new watcher
	watcher, errNew := fsnotify.NewWatcher()
	FatalIfErr(errNew)

	// watching the given files
	for _, filepath := range filepaths {
		Info("Watching path: %s", filepath)
		FatalIfErr(watcher.Add(filepath))
	}

	return watcher
}
