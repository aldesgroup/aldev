// ----------------------------------------------------------------------------
// The code here is about downloading external resources
// ----------------------------------------------------------------------------
package utils

import "sync"

// Downloading external resources, like translations, vendors, etc
func DownloadExternalResources(ctx CancelableContext, cfg *AldevConfig) {
	// making sure the cache folder exists if we need it
	if len(cfg.Vendors) > 0 {
		if cfg.CacheDir == "" {
			Fatal("The cache directory cannot be empty; should be by default: ../tmp")
		}

		EnsureDir(cfg.CacheDir)
	}

	// syncing
	wg := new(sync.WaitGroup)

	// proceed to download external resources
	goRoutine(wg, downloadTranslationsFromGoogle, ctx, cfg)

	// proceed to download external resources
	goRoutine(wg, fetchVendoredLibraries, ctx, cfg)

	// waiting here for all the tasks to be finished
	wg.Wait()
}
