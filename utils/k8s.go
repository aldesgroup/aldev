// ----------------------------------------------------------------------------
// The code here is about using kubernetes
// ----------------------------------------------------------------------------
package utils

import (
	"os"
	"path"
	"strings"
	"time"
)

var (
	tiltOptions string
)

// Making sure we have a ConfigMap to pass to K8S before deploying to a local cluster
func EnsureConfigmap() {
	Debug("Making sure the configmap is up-to-date")
	configFilepath := path.Join(Config().API.SrcDir, Config().API.Config)

	// some controls first
	if Config().Deploying.Dir == "" {
		Fatal(nil, "Empty 'deploying.dir' config!")
	}
	configFile, errStat := os.Stat(configFilepath)
	FatalIfErr(nil, errStat)

	// (re)init the file
	baseDir := EnsureDir(nil, Config().Deploying.Dir, "base")
	configMapFilename := path.Join(baseDir, Config().AppName+"-cm.yaml")
	WriteStringToFile(nil, configMapFilename, "# generated from api/config.yaml by Aldev")

	// creating the config map
	cmd := "kubectl create configmap %s-configmap" // creating a configmap object here
	cmd += " -o yaml"                              // not forgetting the namespace here, and we want a YAML output...
	cmd += " --dry-run=client --from-file=%s"      // ... so we dry-run this, from the config file found in the API sources
	fileContentBytes := RunAndGet("We need to build a configmap from our API's config", ".", false,
		cmd, Config().AppName, configFilepath)

	// tweaking it
	fileContent := string(fileContentBytes)
	fileContent = strings.Replace(fileContent, "creationTimestamp: null", "creationTimestamp: \"%s\"", 1)

	// outputting it
	WriteStringToFile(nil, configMapFilename, fileContent, configFile.ModTime().Format("2006-01-02T15:04:05Z"))
}

func DeployToLocalCluster(ctx CancelableContext) {
	if !IsLinux() {
		Fatal(ctx, "This mode is not supported in non-Linux environments (yet)")
	}

	// computing the custom options
	if swapCode {
		tiltOptions = " --use-local"
	}
	if IsDevAPI() {
		tiltOptions += " --api-only"
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
