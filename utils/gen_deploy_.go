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
	"sync"

	"github.com/aldesgroup/aldev/templates"
	core "github.com/aldesgroup/corego"
	"gopkg.in/yaml.v3"
)

// ----------------------------------------------------------------------------
// Driving regeneration
// ----------------------------------------------------------------------------

var regen bool

func SetRegen(isRegen bool) {
	regen = isRegen
}

func IsRegen() bool {
	return regen
}

// ----------------------------------------------------------------------------
// Dealing with several types of remote deployments
// ----------------------------------------------------------------------------

type remoteDeploymentGenerator interface {
	getPlatform() string
	generateDeployConfig(remoteDir string)
}

var remoteDeploymentGeneratorRegistry = &struct {
	generators map[string]remoteDeploymentGenerator // all the business objects! mapped by the name
	mx         sync.Mutex
}{
	generators: map[string]remoteDeploymentGenerator{},
}

func registerRemoteDeploymentGenerator(generator remoteDeploymentGenerator) {
	remoteDeploymentGeneratorRegistry.mx.Lock()
	defer remoteDeploymentGeneratorRegistry.mx.Unlock()
	remoteDeploymentGeneratorRegistry.generators[generator.getPlatform()] = generator
}

// ----------------------------------------------------------------------------
// Creating all the files to be able to deploy locally and remotely
// ----------------------------------------------------------------------------

func GenerateDeployFiles(ctx CancelableContext) {
	if Config().Deploying == nil {
		Info("No 'deploying' section in the config")
		return
	}
	if Config().Deploying.Platform == nil {
		Info("No deployment platform configured")
		return
	}
	if Config().Deploying.Platform.Type == "" {
		Info("Deployment platform type should not be empty")
		return
	}
	if Config().Deploying.Platform.Config == nil {
		Info("'config' tag is absent for deployment platform '%s'", Config().Deploying.Platform.Type)
		return
	}

	// do we need to generate stuff?
	generationNeeded := regen
	if regen {
		Info("Regeneration has been required!")
	}

	// we might want to regen after a failed generation, that's started but never finished
	generationOngoing := !core.FileExists(".confgen") || string(core.ReadFile(".confgen", false)) == "ongoing"
	if generationOngoing {
		Info("Regeneration will happen because of a previous failed generation")
	}
	generationNeeded = generationNeeded || generationOngoing

	// we should regen if the config has changed - it shouldn't happen very often anyway
	aldevConfigHasChanged := !core.FileExists(".confgen") || core.EnsureModTime(".confgen").Before(core.EnsureModTime(".aldev.yaml"))
	if aldevConfigHasChanged {
		Info("Regeneration will happen because '.aldev.yaml' has changed since the last generation")
	}
	generationNeeded = generationNeeded || aldevConfigHasChanged

	// we're not generating if it's already there, or we're regenerating
	if generationNeeded {
		// LFG
		core.WriteStringToFile(".confgen", "ongoing")

		// checking the configured remote environments for the API
		// and the described deployed environments are the same
		configuredEnvs := core.GetSortedKeys(Config().API.Runtimes.Remote)
		deployedEnvs := core.GetSortedKeys(Config().Deploying.Platform.Config.Environments)
		if !core.SlicesEquals(configuredEnvs, deployedEnvs, true) {
			core.PanicMsg("The configured API runtimes (%s) and the deployed environments (%s) don't match!",
				strings.Join(configuredEnvs, ", "), strings.Join(deployedEnvs, ", "))
		}

		// --------------------------------------------------------------------
		// API runtimes configuration
		// --------------------------------------------------------------------

		// we at least need to generate the local API config file
		ensureLocalEnvType()
		generateAPIConfFile("local", Config().API.Runtimes.Local, regen)

		// if we've configured remote environments, then we'll need config files for them
		for envName, envConf := range Config().API.Runtimes.Remote {
			generateAPIConfFile(envName, envConf, regen)
		}

		// --------------------------------------------------------------------
		// API deployment configuration
		// --------------------------------------------------------------------

		// what we need for local deployment
		localDir := core.EnsureDir(Config().Deploying.Dir, "local")
		EnsureFileFromTemplate(path.Join(localDir, "nginx.conf"), templates.LocalNGINX)
		EnsureFileFromTemplate(path.Join(localDir, "compose.yaml"), templates.LocalCOMPOSE)

		// now the base for remote deployment, i.e. the container file
		containerFilePath := path.Join(Config().Deploying.Dir, "Containerfile")
		moduleContent := core.ReadFile(path.Join(Config().API.SrcDir, "go.mod"), true) // reads all the go.mod file
		moduleGoVersion := core.Before(core.After(string(moduleContent), "go "), "\n") // keeps only the go version, i.e. 1.24.3
		moduleGoVersion = moduleGoVersion[:strings.LastIndex(moduleGoVersion, ".")]    // keeps only the major & middle numbers, i.e. 1.24
		EnsureFileFromTemplate(containerFilePath, templates.ContainerFILE, moduleGoVersion)

		if verbose {
			port := getEnvPortString(Config().API.Runtimes.Local)
			Debug("The containerfile used for remote deployment can be used with the following commands (Example): ")
			Debug("podman build --build-arg ENV=local -f %[1]s -t %[2]s-test-api . && podman run --rm -it -p %[3]s:%[3]s %[2]s-test-api",
				containerFilePath, Config().AppNameShort, port)
			Debug("Test it with: curl http://localhost:%d/rest/translation/fr\\?Namespace\\=Common", port)
		}

		// --------------------------------------------------------------------
		// Platform-dependant stuff right here
		// --------------------------------------------------------------------

		// what we need for remote deployment - which depend on the targeted platform
		generator := remoteDeploymentGeneratorRegistry.generators[Config().Deploying.Platform.Type]
		core.PanicMsgIf(generator == nil, "No deploy config generator found for remote platform '%s'", Config().Deploying.Platform.Type)
		generator.generateDeployConfig(core.EnsureDir(Config().Deploying.Dir, "remote"))

		// --------------------------------------------------------------------
		// Finishing
		// --------------------------------------------------------------------
		core.WriteStringToFile(".confgen", "done")
	} else {
		Debug("Configuration generation is not required")
	}

	core.WriteStringToFile(".confgen", "true")

}

// ----------------------------------------------------------------------------
// Generating API config files
// ----------------------------------------------------------------------------

func generateAPIConfFile(envName string, envConfig *APIRuntimeConfig, regen bool) {
	// checking the existing config file
	confFileName := fmt.Sprintf("%s/conf-%s.yaml", Config().API.SrcDir, envName)
	if core.FileExists(confFileName) && !regen {
		Debug("%s already exists!", confFileName)
		return
	}

	// building the base config for the current environment
	base := mergeYAMLNodes(&Config().API.Runtimes.Common.Base, &envConfig.Base)
	base = prependValue("port", getEnvPortString(envConfig), base)
	base = prependValue("appdesc", Config().AppDesc, base)
	base = prependValue("appname", Config().AppName, base)

	// making a new config by merging the env config on top of the common config
	newConf := &struct {
		Base   yaml.Node
		Custom yaml.Node
	}{
		Base:   *base,
		Custom: *mergeYAMLNodes(&Config().API.Runtimes.Common.Custom, &envConfig.Custom),
	}

	// marshalling
	newConfBytes, errMarsh := yaml.Marshal(newConf)
	core.PanicMsgIfErr(errMarsh, "Could not marshal merged config for env '%s'", envName)

	// writing out the file
	core.WriteBytesToFile(confFileName, newConfBytes)
	Debug("Just wrote file : %s", confFileName)
}

// ----------------------------------------------------------------------------
// Utils
// ----------------------------------------------------------------------------

// returns the env type (as a string) from the given environment with the given name
func getEnvTypeString(envName string, envConfig *APIRuntimeConfig, failIfEmpty bool) string {
	baseConfig := envConfig.Base.Content
	for i, node := range baseConfig {
		if node.Value == "envtype" {
			return baseConfig[i+1].Value
		}
	}

	core.PanicMsgIf(failIfEmpty, "Runtime env '%s' does not define its 'envtype'!", envName)
	return ""
}

// returns the env type from the remote environment with the given name
func getEnvType(envName string, failIfEmpty bool) core.EnvType {
	return core.EnvTypeValFrom(getEnvTypeString(envName, Config().API.Runtimes.Remote[envName], failIfEmpty))
}

// this sets the "LOCAL" env type on the local runtime config, if not present already
func ensureLocalEnvType() {
	if localEnv := Config().API.Runtimes.Local; getEnvTypeString("local", localEnv, false) == "" {
		// init if needed
		if localEnv.Base.Kind == 0 {
			localEnv.Base.Kind = yaml.MappingNode
			localEnv.Base.Tag = "!!map"
		}

		// adding the env type
		localEnv.Base.Content = append(localEnv.Base.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "envtype"},
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: core.EnvTypeLOCAL.String()},
		)
	}
}

func getEnvPortString(envConfig *APIRuntimeConfig) string {
	if envConfig.Port > 0 {
		return strconv.Itoa(envConfig.Port)
	}
	if Config().API.Runtimes.Common != nil {
		return strconv.Itoa(Config().API.Runtimes.Common.Port)
	}
	return ""
}

func prependValue(yamlName, value string, toNode *yaml.Node) *yaml.Node {
	toNode.Content = append(
		[]*yaml.Node{
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: yamlName},
			{Kind: yaml.ScalarNode, Value: value},
		},
		toNode.Content...)

	return toNode
}

// mergeYAMLNodes merges node B on top of node A.
// - Mappings: keys from B override or extend A
// - Sequences/Scalars: B fully replaces A
// Returns a new merged node (A and B are not mutated).
func mergeYAMLNodes(a, b *yaml.Node) *yaml.Node {
	// Unwrap documents
	if a.Kind == yaml.DocumentNode {
		a = a.Content[0]
	}
	if b.Kind == yaml.DocumentNode {
		b = b.Content[0]
	}

	// If B is zero-value or null, A wins
	if b.Kind == 0 || (b.Kind == yaml.ScalarNode && b.Tag == "!!null") {
		return deepCopyNode(a)
	}

	// Only deep-merge mappings; everything else (sequences, scalars) => B wins
	if a.Kind != yaml.MappingNode || b.Kind != yaml.MappingNode {
		return deepCopyNode(b)
	}

	// Build a key->valueNode index from A
	aIndex := make(map[string]int) // key => index of value node in a.Content
	for i := 0; i < len(a.Content)-1; i += 2 {
		aIndex[a.Content[i].Value] = i + 1
	}

	// Start with a deep copy of A
	merged := deepCopyNode(a)

	// Apply keys from B
	for i := 0; i < len(b.Content)-1; i += 2 {
		keyNode := b.Content[i]
		valNode := b.Content[i+1]
		key := keyNode.Value

		if aValIdx, exists := aIndex[key]; exists {
			// Key exists in A: recurse to merge the values
			merged.Content[aValIdx] = mergeYAMLNodes(merged.Content[aValIdx], valNode)
		} else {
			// Key is new: append a deep copy of the key+value pair
			merged.Content = append(merged.Content, keyNode, deepCopyNode(valNode))
		}
	}

	return merged
}

// deepCopyNode recursively clones a yaml.Node.
func deepCopyNode(n *yaml.Node) *yaml.Node {
	if n == nil {
		return nil
	}
	clone := *n // shallow copy of the struct
	if len(n.Content) > 0 {
		clone.Content = make([]*yaml.Node, len(n.Content))
		for i, child := range n.Content {
			clone.Content[i] = deepCopyNode(child)
		}
	}
	return &clone
}
