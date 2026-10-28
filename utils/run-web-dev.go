// ----------------------------------------------------------------------------
// The code here is about local deployment with Containers
// ----------------------------------------------------------------------------
package utils

import (
	"os"
)

// this function allows to us to continuously develop our Web app
func RunWebDev(ctx CancelableContext) {
	go func() {
		webAppCtx := NewBaseContext().WithStdErrWriter(os.Stdout).WithStdOutWriter(os.Stdout).
			WithAllowFailure(true).WithExecDir(getWebSrcDir())
		if IsDevLibrary() {
			Run("Building the library", webAppCtx, true, "npm run dev")
		} else {
			Run("Starting the web app", webAppCtx, true, "npm run dev -- --port %d", Config().Web.Build.Port)
		}
	}()
}
