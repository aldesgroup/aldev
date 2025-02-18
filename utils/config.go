// ----------------------------------------------------------------------------
// The code here is about the configuration of Aldev
// ----------------------------------------------------------------------------
package utils

import (
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

var (
	// the instance bearing all the configuration
	config *AldevConfig
)

func Config() *AldevConfig {
	if config == nil {
		Fatal(nil, "Aldev configuration has never been read!")
	}

	return config
}

type AldevConfig struct {
	AppName   string    // the name of the app - beware: the key has to be "appname" in the YAML file
	Languages string    // the languages available for this app, seperated by a comma - for example: en,fr,it,de,zh,es
	Lib       *struct { // must be filled if this project is a library
		SrcDir         string // where the library source code can be found
		Config         string // the path to the config file for the API, from the API's folder
		Install        string // command that should be run to install stuff, like needed dependencies, etc.
		Develop        string // command that should be run to allow for continuously developping & building the library
		BinDir         string // the directory where to find the library's compiled binary, as seen from the library source folder (srcdir)
		resolvedBinDir string // the bin directory as seen from the project's root
	}
	API *struct { // must be filled if there's an API
		SrcDir         string // where the API's Goald-based code should be found
		Config         string // the path to the config file for the API, from the API's folder
		Port           int    // the port used to expose the whole load-balanced API service
		I18n           *I18nConfig
		DataDir        string // where to find bootstraping data to run the app
		BinDir         string // the directory where to find the API's compiled binary, as seen from the API source folder (srcdir)
		resolvedBinDir string // the bin directory as seen from the project's root
	}
	// APIOnly bool      // if true, then no web app is handled
	Web *struct { // must be filled if there's a web app
		SrcDir  string      // where the Web app's GoaldR-based code should be found
		Port    int         // the port used to expose the app's frontend
		EnvVars []*struct { // environment variables to pass to the web app
			Name  string // the variable name; must start with "WEB_"
			Desc  string // a description for the
			Value string // the value we're using for the local dev environment
		}
	}
	Native *struct { // must be filled if there's a native app
		SrcDir  string      // where the Native app's GoaldN-based code should be found
		I18n    *I18nConfig //
		DataDir string      // where to find bootstraping data to run the app
	}
	Vendors   []*VendorConfig // external projects to vendor into our project
	Deploying *struct {       // Section for the local deployment of the app
		Dir string // where all the deploying config should be
	}
	CodeSwaps []*CodeSwapsConfig // Automatically, temporarily swapping bits of code
	Symlinks  []*SymlinkConfig   // Create symlinks, to help code-swapping for instance
}

type I18nConfig struct {
	Links   []string // the link to download the translations from; each new file can override previous translations
	KeySize int      // the max size of the key in the translation UID namespace.key
	File    string   // the path of the file where to write the downloaded translations
}

type VendorConfig struct {
	Repo    string // the repo of the external project
	Version string // the maximal version to fetch; can also be the "latest"
	From    string // the path inside the repo to copy; if void, then all the repo is copied
	To      string // the place where to paste the copied cod
}

type CodeSwapsConfig struct {
	From string   // the path from which to look for swaps; "." for the current project, "../../dependency" to swap in another lib
	For  []string // the file paths for which to apply the same swaps; can be provided as a glob, e.g. "./src/**/*.ts?",
	Do   []*struct {
		Replace string // the substring to look for and replace
		With    string // the replacement
		EOFCom  bool   // should the added comment (HOTSWAPPED) be at the end of the line (// ...) rather than inline (/* ... */) ?
	}
}

type SymlinkConfig struct {
	Link string // what to link
	As   string // how to link it
}

// returns the name of the folder where to find the Go source code
func GetGoSrcDir() string {
	if IsDevLibrary() {
		return Config().Lib.SrcDir
	}

	return Config().API.SrcDir
}

func GetBinDir() string {
	if IsDevLibrary() {
		return Config().Lib.BinDir
	}

	return Config().API.BinDir
}

func GetConfigPath() string {
	if IsDevLibrary() {
		return Config().Lib.Config
	}

	return Config().API.Config
}

func ReadConfig(cfgFileName string) {
	Debug("Reading Aldev config")

	config = &AldevConfig{}

	// Reading the config file into bytes
	yamlBytes, errRead := os.ReadFile(cfgFileName)
	FatalIfErr(nil, errRead)

	// Unmarshalling the YAML file
	FatalIfErr(nil, yaml.Unmarshal(yamlBytes, config))

	// Adding the languages to the env vars for the web
	if config.Web != nil {
		config.Web.EnvVars = append(config.Web.EnvVars, &struct {
			Name  string
			Desc  string
			Value string
		}{
			Name:  "WEB_LANGUAGES",
			Desc:  "the languages that should be available in the web app",
			Value: config.Languages,
		})
	}
}

// computed property on an Aldev config object
func (cfg *AldevConfig) ResolvedBinDir() string {
	if IsDevLibrary() {
		if cfg.Lib.resolvedBinDir == "" {
			cfg.Lib.resolvedBinDir = path.Join(cfg.Lib.SrcDir, cfg.Lib.BinDir)
		}

		return cfg.Lib.resolvedBinDir
	}

	if cfg.API.resolvedBinDir == "" {
		cfg.API.resolvedBinDir = path.Join(cfg.API.SrcDir, cfg.API.BinDir)
	}

	return cfg.API.resolvedBinDir
}
