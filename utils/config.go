// ----------------------------------------------------------------------------
// The code here is about the configuration of Aldev
// ----------------------------------------------------------------------------
package utils

import (
	"os"

	"gopkg.in/yaml.v3"
)

type aldevConfig struct {
	AppName string // the name of the app - beware: the key has to be "appname" in the YAML file
	API     struct {
		Dir    string // where the API's Goald-based code should be found
		Config string // the path to the config file for the API
		Port   int    // the port used to expose the whole load-balanced API service
	}
	Web struct {
		Dir      string   // where the Web app's GoaldR-based code should be found
		UseLocal []string // the names of the modules to use locally (npm link); e.g. @aldes/emeraldr, or goaldr
	}
	Deploying struct { // Section for the local deployment of the app
		Dir string // where all the deploying config should be
	}
}

func ReadConfig(cfgFileName string) *aldevConfig {
	Debug("Reading Aldev config")

	cfg := &aldevConfig{}

	// Reading the config file into bytes
	yamlBytes, errRead := os.ReadFile(cfgFileName)
	FatalIfErr(errRead)

	// Unmarshalling the YAML file
	FatalIfErr(yaml.Unmarshal(yamlBytes, cfg))

	return cfg
}
