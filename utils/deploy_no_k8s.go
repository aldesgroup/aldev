// ----------------------------------------------------------------------------
// The code here is about local deployment without Containers / Kubernetes / Tilt
// ----------------------------------------------------------------------------
package utils

import "sync"

func DeployWithNoContainer(ctx CancelableContext) {

	wg := new(sync.WaitGroup)

	// proceed to download external resources
	goRoutine(wg, downloadAllTranslationsFromGoogle, ctx)

	// proceed to download external resources
	goRoutine(wg, fetchVendoredLibraries, ctx)

	// waiting here for all the tasks to be finished
	wg.Wait()
}
