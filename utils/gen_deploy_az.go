// ----------------------------------------------------------------------------
// Generating code - mainly config files - for an app to work locally &
// when its deployed in various environment
// ----------------------------------------------------------------------------

package utils

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/aldesgroup/aldev/templates/azure"
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
	global := Config().Deploying.Platform.Config.Global
	globalDirName := "0-glo"
	globalDir := core.EnsureDir(path.Join(remoteDir, globalDirName))

	// Are we using Gitlab for CI/CD?
	isGitlabCICD := Config().Deploying.CICD != nil && Config().Deploying.CICD.Type == "gitlab"

	// the backend
	EnsureFileFromTemplate(path.Join(globalDir, "backend.tf"), replaceIn("global", global, azure.TerraformAzureBACKEND, "resource_ns"), "global")

	// the main file
	if isGitlabCICD {
		global["gitlab_cid"] = Config().Deploying.CICD.Config["gitlab_cid"]
		global["gitlab_url"] = Config().Deploying.CICD.Config["gitlab_url"]
		global["git_repo"] = getGitRepo()
		EnsureFileFromTemplate(path.Join(globalDir, "main.tf"),
			replaceIn("global", global, azure.TerraformAzureGLOBALxMAINnGITLAB, "resource_ns", "domain_name", "oauth2_scope_guid", "gitlab_cid", "gitlab_url", "git_repo"))
	} else {
		EnsureFileFromTemplate(path.Join(globalDir, "main.tf"), replaceIn("global", global, azure.TerraformAzureGLOBALxMAIN, "resource_ns", "domain_name", "oauth2_scope_guid"))
	}

	// NOW, dealing with the ENVIRONMENTS

	// this is the main file generically describing the infrastructure, for all the environment types
	// so it's meant to be used by each environment with a custom config
	EnsureFileFromTemplate(path.Join(remoteDir, "main.tf"), replaceIn("global", global, azure.TerraformAzureCOMMONxMAIN, "resource_ns"))

	// now dealing for each environment
	for envName, envParams := range Config().Deploying.Platform.Config.Environments {
		// creating the folder for the current environment
		envType := getEnvType(envName, true)
		envDirName := fmt.Sprintf("%d-%s", envType, envName)
		envDir := core.EnsureDir(path.Join(remoteDir, envDirName))

		// creating the Terraform backend file for the current env
		EnsureFileFromTemplate(path.Join(envDir, "backend.tf"), replaceIn(envName, envParams, azure.TerraformAzureBACKEND, "resource_ns"), envName)

		// passing the port
		envParams["port"] = strconv.Itoa(getRemotePort(envName))

		// creating / customizing the Terrafor main file for the current env
		replacements := []string{"location", "identity_sub_name", "management_sub_name", "environment_sub_name", "acr_name", "acr_rg", "domain_name", "port"}
		EnsureFileFromTemplate(path.Join(envDir, "main.tf"), replaceIn(envName, envParams, azure.TerraformAzureENVxMAIN, replacements...), envName)

		// keeping track of the environment we've just dealt with
		global["env-"+envType.String()] = envName                                                      // eg. env-STAGING -> qua
		global["sub-"+envType.String()] = getEnvParamValue(envName, envParams, "environment_sub_name") // eg. sub-STAGING -> subscr-qua

	}

	// the CI/CD conf
	if Config().Deploying.CICD != nil && Config().Deploying.CICD.Type == "gitlab" {
		replacements := []string{
			"management_sub_name", "acr_name", "resource_ns",
			"env-SANDBOX", "sub-SANDBOX",
			"env-STAGING", "sub-STAGING",
			"env-PRODUCTION", "sub-PRODUCTION",
		}
		EnsureFileFromTemplate(".gitlab-ci.yml", replaceIn("", global, azure.GitlabAzureCIxCDxCONF, replacements...))
	} else {
		core.PanicMsg("Empty or unhandled CI/CD method: '%s'", Config().Deploying.CICD)
	}
}

func getEnvParamValue(envName string, envParams DeployEnvConfig, paramName string) string {
	// trying the env value
	if envValue := envParams[paramName]; envValue != "" {
		return envValue
	}

	// else, fallback on the global value - if there's one
	globalValue := Config().Deploying.Platform.Config.Global[paramName]
	core.PanicMsgIf(globalValue == "", "No param '%s' found in the configured env \"%s\"'s params or globalparams", paramName, envName)
	return globalValue
}

func getRemotePort(envName string) int {
	if Config().API.Runtimes == nil || Config().API.Runtimes.Remote == nil || Config().API.Runtimes.Remote[envName] == nil {
		return 0
	}

	if env := Config().API.Runtimes.Remote[envName]; env.Port > 0 {
		return env.Port
	}

	if Config().API.Runtimes.Common != nil {
		return Config().API.Runtimes.Common.Port
	}

	return 0
}

func replaceIn(envName string, envParams DeployEnvConfig, template string, replacementKeys ...string) string {
	replacementPairs := []string{}
	for _, replacementKey := range replacementKeys {
		replacementPairs = append(replacementPairs, "{"+replacementKey+"}", getEnvParamValue(envName, envParams, replacementKey))
	}
	return strings.NewReplacer(replacementPairs...).Replace(template)
}

func getGitRepo() string {
	core.PanicMsgIf(!core.DirExists(".git"), "Git has not been initialised in this project yet!")
	core.PanicMsgIf(!core.FileExists(".git/config"), "Git config is absent, which should not happen!")
	repoLine, _ := core.FindLineInFileStartingWith(".git/config", "url = ", true, true)
	return core.Before(core.After(repoLine, ":"), ".git")
}
