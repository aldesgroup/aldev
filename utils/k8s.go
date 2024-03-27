// ----------------------------------------------------------------------------
// The code here is about using kubernetes
// ----------------------------------------------------------------------------
package utils

import (
	"os"
	"path"
	"strings"
)

func EnsureConfigmap(cfg *AldevConfig) {
	Debug("Making sure the configmap is up-to-date")

	// some controls first
	if cfg.Deploying.Dir == "" {
		Fatal("Empty 'deploying.dir' config!")
	}
	configFile, errStat := os.Stat(cfg.API.Config)
	FatalIfErr(errStat)

	// (re)init the file
	baseDir := EnsureDir(cfg.Deploying.Dir, "base")
	configMapFilename := path.Join(baseDir, cfg.AppName+"-cm.yaml")
	WriteToFile(configMapFilename, "# generated from app-api/config.yaml by Aldev")

	// creating the config map
	cmd := "kubectl create configmap %s-configmap" // creating a configmap object here
	cmd += " -o yaml"                              // not forgetting the namespace here, and we want a YAML output...
	cmd += " --dry-run=client --from-file=%s"      // ... so we dry-run this, from the config file found in the API sources
	fileContentBytes := RunAndGet("We need to build a configmap from our API's config",
		Fatal, cmd, cfg.AppName, cfg.API.Config)

	// tweaking it
	fileContent := string(fileContentBytes)
	fileContent = strings.Replace(fileContent, "creationTimestamp: null", "creationTimestamp: \"%s\"", 1)

	// outputting it
	WriteToFile(configMapFilename, fileContent, configFile.ModTime().Format("2006-01-02T15:04:05Z"))
}
