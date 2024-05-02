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
		Config string // the path to the config file for the API, from the API's folder
		Port   int    // the port used to expose the whole load-balanced API service
		I18n   struct {
			Link    string // the link to download the translations for the current app
			KeySize int    // the max size of the key in the translation UID route.part.key
			File    string // the path of the file where to write the downloaded translations
		}
		Build struct {
			BinDir         string // folder where to put the compiled binary, relatively to the API directory
			ResolvedBinDir string // the bin directory, relatively to the project's rool
			// Library        bool   // if true, then we're "aldeveloping" a library here
		}
		DataDir string // where to find bootstraping data to run the app
	}
	Web struct {
		SrcDir string // where the Web app's GoaldR-based code should be found
		Port   int    // the port used to expose the app's frontend
	}
	Deploying struct { // Section for the local deployment of the app
		Dir string // where all the deploying config should be
	}
	LocalSwaps []*LocalSwapsConfig // Automatically, temporarily swapping bits of code
}

type LocalSwapsConfig struct {
	From string   // the path from which to look for swaps; "." for the current project, "../../dependency" to swap in another lib
	For  []string // the file paths for which to apply the same swaps; can be provided as a glob, e.g. "./src/**/*.ts?",
	Do   []struct {
		Replace string // the substring to look for and replace
		With    string // the replacement
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
