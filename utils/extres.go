// ----------------------------------------------------------------------------
// The code here is about downloading external resources
// ----------------------------------------------------------------------------
package utils

import (
	"sync"

	core "github.com/aldesgroup/corego"
)

// Downloading external resources, like translations, vendors, etc
func DownloadExternalResources(ctx CancelableContext, withTranslations bool) {
	// making sure the cache folder exists if we need it
	if len(Config().Vendors) > 0 {
		core.EnsureDir(GetCacheDir())
	}

	// syncing
	wg := new(sync.WaitGroup)

	// proceed to download external resources
	if withTranslations {
		goRoutine(wg, downloadAllTranslationsFromGoogle, ctx)
	}

	// proceed to download external resources
	goRoutine(wg, fetchVendoredLibraries, ctx)

	// waiting here for all the tasks to be finished
	wg.Wait()
}
