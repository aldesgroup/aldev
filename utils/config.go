// ----------------------------------------------------------------------------
// The code here is about the configuration of Aldev
// ----------------------------------------------------------------------------
package utils

import (
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

type AldevConfig struct {
	AppName string // the name of the app - beware: the key has to be "appname" in the YAML file
	API     struct {
		SrcDir string // where the API's Goald-based code should be found
		Config string // the path to the config file for the API
		Port   int    // the port used to expose the whole load-balanced API service
		I18n   struct {
			Link    string // the link to download the translations for the current app
			KeySize int    // the max size of the key in the translation UID route.part.key
			File    string // the path of the file where to write the downloaded translations
		}
		Build struct {
			BinDir         string // folder where to put the compiled binary, relatively to the API directory
			ResolvedBinDir string // the bin directory, relatively to the project's rool
			Library        bool   // if true, then we're "aldeveloping" a library here
		}
	}
	Web struct {
		SrcDir    string   // where the Web app's GoaldR-based code should be found
		LocalDeps []string // the names of the modules to use locally (npm link); e.g. @aldes/emeraldr, or goaldr
		Port      int      // the port used to expose the app's frontend
	}
	Deploying struct { // Section for the local deployment of the app
		Dir string // where all the deploying config should be
	}
}

func ReadConfig(cfgFileName string) *AldevConfig {
	Debug("Reading Aldev config")

	cfg := &AldevConfig{}

	// Reading the config file into bytes
	yamlBytes, errRead := os.ReadFile(cfgFileName)
	FatalIfErr(errRead)

	// Unmarshalling the YAML file
	FatalIfErr(yaml.Unmarshal(yamlBytes, cfg))

	// Some useful computation
	cfg.API.Build.ResolvedBinDir = path.Join(cfg.API.SrcDir, cfg.API.Build.BinDir)

	return cfg
}
