// ----------------------------------------------------------------------------
// The code here is about using kubernetes
// ----------------------------------------------------------------------------
package utils

import (
	"os"
	"strings"
)

func EnsureNamespace(cfg *aldevConfig) {
	Debug("Making sure the correct namespace exists")

	if string(RunAndGet(Noop, "kubectl get namespace %s-local", cfg.AppName)) == "" {
		Run("kubectl create namespace %s-local", cfg.AppName)
	}
}

func EnsureConfigmap(cfg *aldevConfig) {
	Debug("Making sure the configmap is up-to-date")

	// some controls first
	if cfg.Deploying.Dir == "" {
		Fatal("Empty 'deploying.dir' config!")
	}
	configFile, errStat := os.Stat(cfg.API.Config)
	FatalIfErr(errStat)

	// (re)init the file
	configMapFilename := cfg.Deploying.Dir + "/" + cfg.AppName + "-cm.yaml"
	EnsureDir(cfg.Deploying.Dir)
	WriteToFile(configMapFilename, "# generated from app-api/config.yaml by Aldev")

	// creating the config map
	cmd := "kubectl create configmap %s-configmap" // creating a configmap object here
	cmd += " -n %s-local -o yaml"                  // not forgetting the namespace here, and we want a YAML output...
	cmd += " --dry-run=client --from-file=%s"      // ... so we dry-run this, from the config file found in the API sources
	fileContentBytes := RunAndGet(Fatal, cmd, cfg.AppName, cfg.AppName, cfg.API.Config)

	// tweaking it
	fileContent := string(fileContentBytes)
	fileContent = strings.Replace(fileContent, "creationTimestamp: null", "creationTimestamp: \"%s\"", 1)

	// outputting it
	WriteToFile(configMapFilename, fileContent, configFile.ModTime().Format("2006-01-02T15:04:05Z"))
}
