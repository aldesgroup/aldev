// ----------------------------------------------------------------------------
// Generating code - mainly config files - for an app to work locally &
// when its deployed in various environment
// ----------------------------------------------------------------------------

package utils

import (
	"fmt"
	"path"
	"strings"

	"github.com/aldesgroup/aldev/templates"
	core "github.com/aldesgroup/corego"
)

func init() {
	registerRemoteDeploymentGenerator(&azureDeploymentGenerator{})
}

type azureDeploymentGenerator struct{}

func (thisGen *azureDeploymentGenerator) getPlatform() string {
	return "azure"
}

func (thisGen *azureDeploymentGenerator) generateDeployConfig(remoteDir string) {
	// Starting with the GLOBAL stuff
	global := Config().Deploying.Global
	globalDirName := fmt.Sprintf("0-%s", global.Name)
	globalDir := core.EnsureDir(path.Join(remoteDir, globalDirName))

	// the backend
	EnsureFileFromTemplate(path.Join(globalDir, "backend.tf"), replaceIn(global, templates.TerraformAzureBACKEND, "resource_ns"), "global")

	// the main file
	EnsureFileFromTemplate(path.Join(globalDir, "main.tf"), replaceIn(global, templates.TerraformAzureGLOBALxMAIN, "resource_ns", "domain_name", "oauth2_scope_guid"))

	// NOW, dealing with the ENVIRONMENTS

	// this is the main file generically describing the infrastructure, for all the environment types
	// so it's meant to be used by each environment with a custom config
	EnsureFileFromTemplate(path.Join(remoteDir, "main.tf"), replaceIn(global, templates.TerraformAzureCOMMONxMAIN, "resource_ns"))

	// now dealing for each environment
	for i, env := range Config().Deploying.Environments {
		// creating the folder for the current environment
		envDirName := fmt.Sprintf("%d-%s", i+1, env.Name)
		envDir := core.EnsureDir(path.Join(remoteDir, envDirName))

		// creating the Terraform backend file for the current env
		EnsureFileFromTemplate(path.Join(envDir, "backend.tf"), replaceIn(env, templates.TerraformAzureBACKEND, "resource_ns"), env.Name)

		// creating / customizing the Terrafor main file for the current env
		replacements := []string{"location", "identity_sub_name", "management_sub_name", "environment_sub_name", "acr_name", "acr_rg", "domain_name"}
		EnsureFileFromTemplate(path.Join(envDir, "main.tf"), replaceIn(env, templates.TerraformAzureENVxMAIN, replacements...), env.Name)
	}
}

func getEnvParamValue(paramName string, env *DeployEnvConfig) string {
	// trying the env value
	if envValue := env.Params[paramName]; envValue != "" {
		return envValue
	}

	// else, fallback on the global value - if there's one
	globalValue := Config().Deploying.Global.Params[paramName]
	core.PanicMsgIf(globalValue == "", "No param '%s' found in the configured env \"%s\"'s params or globalparams", paramName, env.Name)
	return globalValue
}

func replaceIn(env *DeployEnvConfig, template string, replacementKeys ...string) string {
	replamentPairs := []string{}
	for _, replacementKey := range replacementKeys {
		replamentPairs = append(replamentPairs, "{"+replacementKey+"}", getEnvParamValue(replacementKey, env))
	}
	return strings.NewReplacer(replamentPairs...).Replace(template)
}
