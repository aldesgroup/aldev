package utils

import "time"

var (
	useLocalDeps bool
	tiltOptions  string
)

func SetUseLocalDeps() {
	useLocalDeps = true
}

func Launch(ctx CancelableContext, cfg *AldevConfig) {

	// computing the custom options
	if useLocalDeps {
		tiltOptions = " --use-local"
	}
	if tiltOptions != "" {
		tiltOptions = " --" + tiltOptions
	}

	// making sure the namespace is fresh
	kustomization := "dev"
	if useLocalDeps {
		kustomization = "local"
	}
	if string(RunAndGet("We want to check what's in our namespace", ".", false,
		"kubectl get all --namespace %s-%s", cfg.AppName, kustomization)) != "" {
		Run("The namespace needs some cleanup first", ctx, false, "tilt down%s", tiltOptions)
	}

	// making sure we clean up at the end
	defer func() {
		time.Sleep(100 * time.Millisecond)
		Run("We'll clean up the context now", ctx, false, "tilt down%s", tiltOptions)
	}()

	// Running a command that never finishes, with the cancelable context
	mode := ""
	if verbose {
		mode = " --verbose --debug"
	}
	Run("Now we start Tilt to handle all the k8s & docker deployments",
		ctx, true, "tilt up%s --stream%s", mode, tiltOptions)

	// Wait for the context to be canceled or the program to exit
	<-ctx.Done()
}
