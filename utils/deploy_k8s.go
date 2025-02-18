// ----------------------------------------------------------------------------
// The code here is about local deployment with Containers / Kubernetes / Tilt
// ----------------------------------------------------------------------------
package utils

import (
	"os"
)

var (
	tiltOptions string
)

func DeployToLocalCluster(ctx CancelableContext) {
	// computing the custom options
	if swapCode {
		tiltOptions = " --use-local"
	}
	if IsDevAPI() && !(IsDevWebApp() || IsDevNative()) {
		tiltOptions += " --api-only"
	}
	if verbose {
		tiltOptions += " --verbose"
	}
	if tiltOptions != "" {
		tiltOptions = " --" + tiltOptions
	}

	// making sure the namespace is fresh
	kustomization := "dev"
	if swapCode {
		kustomization = "local"
	}
	if string(RunAndGet("We want to check what's in our namespace", ".", false,
		"kubectl get all --namespace %s-%s", Config().AppName, kustomization)) != "" {
		Run("The namespace needs some cleanup first", ctx, false, "tilt down%s", tiltOptions)
	}

	// making sure we clean up at the end
	defer func() {
		Run("We'll clean up the context now", NewBaseContext().WithStdErrWriter(os.Stdout).WithStdOutWriter(os.Stdout), true, "tilt down%s", tiltOptions)
	}()

	// Running a command that never finishes, with the cancelable context
	mode := ""
	if verbose {
		mode = " --verbose --debug"
	}
	Run("Now we start Tilt to handle all the k8s & docker deployments",
		ctx.WithErrLogFn(ErrorAndCancel), true, "tilt up%s --stream%s", mode, tiltOptions)

	// Wait for the context to be canceled or the program to exit
	<-ctx.Done()
}
