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
	registerRemoteDeploymentGenerator(&fullAzureDeploymentGenerator{})
}

// fullAzureDeploymentGenerator is responsible for generating the deployment configuration
// for the Azure platform for the API backend as well as the web frontend.
type fullAzureDeploymentGenerator struct{}

// ----------------------------------------------------------------------------
// Implementing the iRemoteDeploymentGenerator interface
// ----------------------------------------------------------------------------

func (thisGen *fullAzureDeploymentGenerator) getPlatform() string {
	return "fullAzure"
}

func (thisGen *fullAzureDeploymentGenerator) getPlatformHost() string {
	return "azure-api.net"
}

const infraDir = "a-infra"
const apimDir = "b-apim"
const webDir = "c-web"

func (thisGen *fullAzureDeploymentGenerator) generateDeployConfig(remoteDir string, resolvedAPIConfigs map[string]map[string]any) {
	// generating the Terraform files for the infrastructure deployment
	thisGen.generateInfraDeployment(remoteDir)

	// generating the Terraform files for the APIM deployment
	thisGen.generateAPIMDeployment(remoteDir)

	// generating the Gitlab CI/CD config file if needed
	thisGen.generateCiCdConfig(remoteDir)

	// generating the web deployment config files if needed
	thisGen.generateWebDeployment(remoteDir)

	// generating additional instructions if needed
	thisGen.generateAdditionalInstructions(resolvedAPIConfigs)
}

func (thisGen *fullAzureDeploymentGenerator) GetServers() (servers map[core.EnvType]string) {
	servers = map[core.EnvType]string{}
	for envName, envParams := range Config().Deploying.Platform.Config.Environments {
		apimName := getEnvParamValue(envName, envParams, "apim_name")
		envURL := fmt.Sprintf("https://%s.azure-api.net/%s-%s", apimName, Config().AppNameLower, envName)
		servers[getEnvType(envName, true)] = envURL
	}

	return
}

// ----------------------------------------------------------------------------
// Generating the infrastructure deployment config files for Azure
// ----------------------------------------------------------------------------

func (thisGen *fullAzureDeploymentGenerator) generateInfraDeployment(remoteDir string) {
	// Starting with the GLOBAL stuff
	global := Config().Deploying.Platform.Config.Global
	globalDirName := "0-glo"
	globalDir := core.EnsureDir(path.Join(remoteDir, infraDir, globalDirName))

	// Are we using Gitlab for CI/CD?
	isGitlabCICD := Config().Deploying.CICD != nil && Config().Deploying.CICD.Type == "gitlab"

	// the backend
	EnsureFileFromTemplate(path.Join(globalDir, "backend.tf"), replaceIn("global", global, azure.TerraformAzureBACKEND, "resource_ns"), "global")

	// the main file
	if isGitlabCICD {
		gitlabClientID := Config().Deploying.CICD.Config["gitlab_cid"]
		global["gitlab_cid"] = gitlabClientID
		global["GITLAB_CID_SNAKE"] = strings.ToUpper(strings.ReplaceAll(gitlabClientID, "-", "_"))
		global["SOME_SPACE"] = strings.Repeat(" ", len(global["GITLAB_CID_SNAKE"])-3)
		global["APP_NAME_SPACE0"] = strings.Repeat(" ", len(Config().AppNameLower))                            // the spacing due to the difference appname
		global["APP_NAME_SPACE1"] = strings.Repeat(" ", len(Config().AppNameKebab)-len(Config().AppNameLower)) // the spacing due to the difference (app-name - appname)
		global["APP_NAME_SPACE2"] = strings.Repeat(" ", len(Config().AppNameKebab))                            // the spacing due to the app-name
		global["gitlab_url"] = Config().Deploying.CICD.Config["gitlab_url"]
		global["git_repo"] = getGitRepo()
		EnsureFileFromTemplate(path.Join(globalDir, "main.tf"),
			replaceIn("global", global, azure.TerraformAzureINFRAxGLOBALnGITLAB, "resource_ns", "domain_name", "oauth2_scope_guid", "gitlab_cid", "gitlab_url", "git_repo"))
	} else {
		EnsureFileFromTemplate(path.Join(globalDir, "main.tf"), replaceIn("global", global, azure.TerraformAzureINFRAxGLOBAL, "resource_ns", "domain_name", "oauth2_scope_guid"))
	}

	// NOW, dealing with the ENVIRONMENTS

	// this is the main file generically describing the infrastructure, for all the environment types
	// so it's meant to be used by each environment with a custom config
	EnsureFileFromTemplate(path.Join(remoteDir, infraDir, "main.tf"), replaceIn("global", global, azure.TerraformAzureINFRAxCOMMON, "resource_ns"))

	// now dealing for each environment
	for envName, envParams := range Config().Deploying.Platform.Config.Environments {
		// creating the folder for the current environment
		envType := getEnvType(envName, true)
		envDirName := fmt.Sprintf("%d-%s", envType, envName)
		envDir := core.EnsureDir(path.Join(remoteDir, infraDir, envDirName))

		// creating the Terraform backend file for the current env
		EnsureFileFromTemplate(path.Join(envDir, "backend.tf"), replaceIn(envName, envParams, azure.TerraformAzureBACKEND, "resource_ns"), envName)

		// passing the port
		envParams["port"] = strconv.Itoa(getRemotePort(envName))

		// creating / customizing the Terraform main file for the current env
		replacements := []string{
			"location", "identity_sub_name", "environment_sub_name",
			"acr_sub_name", "acr_name", "acr_rg",
			"domain_name", "port",
			"apim_sub_name", "apim_name", "apim_rg"}
		EnsureFileFromTemplate(path.Join(envDir, "main.tf"), replaceIn(envName, envParams, azure.TerraformAzureINFRAxENV, replacements...), envName)

		// taking the opportunity to store environment-specific global variables
		global["env-"+envType.String()] = envName                                                      // eg. env-STAGING -> qua
		global["sub-"+envType.String()] = getEnvParamValue(envName, envParams, "environment_sub_name") // eg. sub-STAGING -> subscr-qua

		if Config().Web != nil {
			global["fd-sub-"+envType.String()] = getEnvParamValue(envName, envParams, "frontdoor_sub_name") // eg. fd-sub-STAGING
			global["fd-rg-"+envType.String()] = getEnvParamValue(envName, envParams, "frontdoor_rg")        // eg. fd-rg-STAGING
			global["fd-name-"+envType.String()] = getEnvParamValue(envName, envParams, "frontdoor_name")    // eg. fd-name-STAGING
		}
	}
}

// ----------------------------------------------------------------------------
// Generating the Gitlab CI/CD config file
// ----------------------------------------------------------------------------

func (thisGen *fullAzureDeploymentGenerator) generateCiCdConfig(remoteDir string) {
	// the Gitlab CI/CD conf
	if Config().Deploying.CICD != nil && Config().Deploying.CICD.Type == "gitlab" {
		replacements := []string{
			"identity_sub_name", "acr_sub_name",
			"acr_name", "resource_ns",
			"env-SANDBOX", "sub-SANDBOX",
			"env-STAGING", "sub-STAGING",
			"env-PRODUCTION", "sub-PRODUCTION",
			"GITLAB_CID_SNAKE", "SOME_SPACE",
			"APP_NAME_SPACE0", "APP_NAME_SPACE1", "APP_NAME_SPACE2",
		}

		withWeb := Config().Web != nil
		if withWeb {
			replacements = append(replacements, "fd-sub-SANDBOX", "fd-rg-SANDBOX", "fd-name-SANDBOX")
			replacements = append(replacements, "fd-sub-STAGING", "fd-rg-STAGING", "fd-name-STAGING")
			replacements = append(replacements, "fd-sub-PRODUCTION", "fd-rg-PRODUCTION", "fd-name-PRODUCTION")
		}

		EnsureFileFromTemplate(".gitlab-ci.yml", replaceIn("", Config().Deploying.Platform.Config.Global, azure.GitlabAzureCIxCDxCONF(withWeb), replacements...))
	} else {
		core.PanicMsg("Empty or unhandled CI/CD method: '%s'", Config().Deploying.CICD)
	}
}

// ----------------------------------------------------------------------------
// Generating the APIM deployment config files for Azure
// ----------------------------------------------------------------------------

func (thisGen *fullAzureDeploymentGenerator) generateAPIMDeployment(remoteDir string) {
	// No GLOBAL stuff for the APIM - except we need the global config
	global := Config().Deploying.Platform.Config.Global

	// NOW, dealing with the ENVIRONMENTS

	// this is the main file generically describing the infrastructure, for all the environment types
	// so it's meant to be used by each environment with a custom config
	core.EnsureDir(path.Join(remoteDir, apimDir))
	EnsureFileFromTemplate(path.Join(remoteDir, apimDir, "main.tf"), replaceIn("global", global, azure.TerraformAzureAPIMxCOMMON, "resource_ns"))

	// now dealing for each environment
	for envName, envParams := range Config().Deploying.Platform.Config.Environments {
		// creating the folder for the current environment
		envType := getEnvType(envName, true)
		envDirName := fmt.Sprintf("%d-%s", envType, envName)
		envDir := core.EnsureDir(path.Join(remoteDir, apimDir, envDirName))

		// creating the Terraform backend file for the current env
		EnsureFileFromTemplate(path.Join(envDir, "backend.tf"), replaceIn(envName, envParams, azure.TerraformAzureBACKEND, "resource_ns"), envName+".apim")

		// creating / customizing the Terraform main file for the current env
		replacements := []string{"environment_sub_name", "apim_sub_name", "apim_name", "apim_rg", "domain_name"}
		EnsureFileFromTemplate(path.Join(envDir, "main.tf"), replaceIn(envName, envParams, azure.TerraformAzureAPIMxENV, replacements...), envName)
	}
}

// ----------------------------------------------------------------------------
// Generating the web deployment config files for Azure
// ----------------------------------------------------------------------------

func (thisGen *fullAzureDeploymentGenerator) generateWebDeployment(remoteDir string) {
	// nothing to do here if no web configuration is provided
	if Config().Web == nil {
		return
	}

	// this is the main file generically describing the infrastructure, for all the environment types
	// so it's meant to be used by each environment with a custom config
	core.EnsureDir(path.Join(remoteDir, webDir))
	EnsureFileFromTemplate(path.Join(remoteDir, webDir, "main.tf"), replaceIn("", nil, azure.TerraformAzureWEBxCOMMON, "resource_ns"))

	// now dealing for each environment
	for envName, envParams := range Config().Deploying.Platform.Config.Environments {
		// creating the folder for the current environment
		envType := getEnvType(envName, true)
		envDirName := fmt.Sprintf("%d-%s", envType, envName)
		envDir := core.EnsureDir(path.Join(remoteDir, webDir, envDirName))

		// creating the Terraform backend file for the current env
		EnsureFileFromTemplate(path.Join(envDir, "backend.tf"), replaceIn(envName, envParams, azure.TerraformAzureBACKEND, "resource_ns"), envName+".web")

		// building the hostname, which is different for production
		domainName := getEnvParamValue(envName, envParams, "domain_name")
		webName := getEnvParamValue(envName, envParams, "web_name")
		if envType == core.EnvTypePRODUCTION {
			envParams["host_name"] = fmt.Sprintf("%s.%s", webName, domainName)
		} else {
			envParams["host_name"] = fmt.Sprintf("%s-%s.%s", webName, envName, domainName)
		}

		// creating / customizing the Terraform main file for the current env
		replacements := []string{"environment_sub_name", "frontdoor_sub_name", "frontdoor_name", "frontdoor_rg", "host_name"}
		EnsureFileFromTemplate(path.Join(envDir, "main.tf"), replaceIn(envName, envParams, azure.TerraformAzureWEBxENV, replacements...), envName)
	}
}

// ----------------------------------------------------------------------------
// Generating additional instructions for deployment and other stuff
// ----------------------------------------------------------------------------

func (thisGen *fullAzureDeploymentGenerator) generateAdditionalInstructions(resolvedAPIConfigs map[string]map[string]any) {

	// building the local secrets instructions bit
	localSecrets := strings.Join(
		core.MapFn(
			extractSecrets(resolvedAPIConfigs["local"], 0),
			func(s string) string { return s + "=the secret value for this key" },
		), "\n")

	// building the remote secrets instructions bit
	remoteSecrets := ""
	for _, envType := range []core.EnvType{core.EnvTypeSANDBOX, core.EnvTypeSTAGING, core.EnvTypePRODUCTION} {
		envName := envNameForType[envType]
		envConfig := resolvedAPIConfigs[envName]
		if envName == "local" {
			continue
		}
		remoteSecrets += fmt.Sprintf("#### in the '%s' environment:\n\n", envName)
		remoteSecrets += "```sh\n"
		for i, secret := range extractSecrets(envConfig, 0) {
			if i > 0 {
				remoteSecrets += "\n"
			}
			remoteSecrets += fmt.Sprintf("az keyvault secret set --vault-name %s --name %s --value \"this-secret-value\"\n",
				"kv-{resource_ns}-"+envName+"-{{.AppNameLower}}", secret)
		}
		remoteSecrets += "```\n\n"
	}

	// combining the local and remote secrets instructions into the final instructions file
	instructions := fmt.Sprintf(azure.AdditionalINSTRUCTIONS, localSecrets, remoteSecrets)
	replacements := []string{"identity_sub_name", "resource_ns"}
	instructions = replaceIn("", nil, instructions, replacements...)
	EnsureFileFromTemplate("READTOO.md", instructions)
}

// ----------------------------------------------------------------------------
// Utils
// ----------------------------------------------------------------------------

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

// Replaces the keys in the given template with the values from the given env config, and returns the resulting template as a string
func replaceIn(envName string, envParams DeployEnvConfig, template string, replacementKeys ...string) string {
	replacementPairs := []string{}
	for _, replacementKey := range replacementKeys {
		replacementPairs = append(replacementPairs, "{"+replacementKey+"}", getEnvParamValue(envName, envParams, replacementKey))
	}
	return strings.NewReplacer(replacementPairs...).Replace(template)
}

// Gets the Git repo name from the Git config file, to be used in the Terraform backend configuration when deploying with Gitlab CI/CD
func getGitRepo() string {
	core.PanicMsgIf(!core.DirExists(".git"), "Git has not been initialised in this project yet!")
	core.PanicMsgIf(!core.FileExists(".git/config"), "Git config is absent, which should not happen!")
	repoLine, _ := core.FindLineInFileStartingWith(".git/config", "url = ", true, true)
	return core.Before(core.After(repoLine, ":"), ".git")
}
