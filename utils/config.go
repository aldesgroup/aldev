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
	AppName string    // the name of the app - beware: the key has to be "appname" in the YAML file
	Lib     *struct { // if this section's non-empty, then this project is not an app but a library, and section "API", "Web", "Deploying" are discarded
		SrcDir  string // where the library source code can be found
		Install string // command that should be run to install stuff, like needed dependencies, etc.
		Develop string // command that should be run to allow for continuously developping & building the library
	}
	API *struct {
		SrcDir string // where the API's Goald-based code should be found
		Config string // the path to the config file for the API, from the API's folder
		Port   int    // the port used to expose the whole load-balanced API service
		I18n   *struct {
			Link    string // the link to download the translations for the current app
			KeySize int    // the max size of the key in the translation UID route.part.key
			File    string // the path of the file where to write the downloaded translations
		}
		DataDir        string // where to find bootstraping data to run the app
		BinDir         string // the directory where to find the API's compiled binary, as seen from the API source folder (srcdir)
		resolvedBinDir string // the bin directory as seen from the project's root
	}
	Web *struct {
		SrcDir  string      // where the Web app's GoaldR-based code should be found
		Port    int         // the port used to expose the app's frontend
		EnvVars []*struct { // environment variables to pass to the web app
			Name  string // the variable name; must start with "WEB_"
			Desc  string // a description for the
			Value string // the value we're using for the local dev environment
		}
	}
	Vendors   []*VendorConfig // external projects to vendor into our project
	Deploying *struct {       // Section for the local deployment of the app
		Dir string // where all the deploying config should be
	}
	LocalSwaps []*LocalSwapsConfig // Automatically, temporarily swapping bits of code
}

type LocalSwapsConfig struct {
	From string   // the path from which to look for swaps; "." for the current project, "../../dependency" to swap in another lib
	For  []string // the file paths for which to apply the same swaps; can be provided as a glob, e.g. "./src/**/*.ts?",
	Do   []*struct {
		Replace string // the substring to look for and replace
		With    string // the replacement
		EOFCom  bool   // should the added comment (HOTSWAPPED) be at the end of the line (// ...) rather than inline (/* ... */) ?
	}
}

type VendorConfig struct {
	Repo    string // the repo of the external project
	Version string // the maximal version to fetch; can also be the "latest"
	From    string // the path inside the repo to copy; if void, then all the repo is copied
	To      string // the place where to paste the copied cod
}

func (thisCfg *AldevConfig) IsLibrary() bool {
	return thisCfg.Lib != nil
}

func (thisCfg *AldevConfig) GetResolvedBinDir() string {
	if thisCfg.API.resolvedBinDir == "" {
		thisCfg.API.resolvedBinDir = path.Join(thisCfg.API.SrcDir, thisCfg.API.BinDir)
	}

	return thisCfg.API.resolvedBinDir
}

func ReadConfig(cfgFileName string) *AldevConfig {
	Debug("Reading Aldev config")

	cfg := &AldevConfig{}

	// Reading the config file into bytes
	yamlBytes, errRead := os.ReadFile(cfgFileName)
	FatalIfErr(nil, errRead)

	// Unmarshalling the YAML file
	FatalIfErr(nil, yaml.Unmarshal(yamlBytes, cfg))

	return cfg
}
