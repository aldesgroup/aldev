package utils

import "github.com/fsnotify/fsnotify"

func WatcherFor(filepaths ...string) *fsnotify.Watcher {
	// new watcher
	watcher, errNew := fsnotify.NewWatcher()
	FatalIfErr(nil, errNew)

	// watching the given files
	for _, filepath := range filepaths {
		Debug("Watching path: %s", filepath)
		FatalIfErr(nil, watcher.Add(filepath))
	}

	return watcher
}
