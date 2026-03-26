// ----------------------------------------------------------------------------
// Generating code - mainly config files - for an app to work locally &
// when its deployed in various environment
// ----------------------------------------------------------------------------

package utils

import (
	"log/slog"
	"path"
	"strings"
	"sync"

	"github.com/aldesgroup/aldev/templates"
	core "github.com/aldesgroup/corego"
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
	// we're not generating if it's already there, or we're regenerating
	if !core.DirExists(Config().Deploying.Dir) || regen {

		// what we need for local deployment
		localDir := core.EnsureDir(Config().Deploying.Dir, "local")
		EnsureFileFromTemplate(path.Join(localDir, "nginx.conf"), templates.LocalNGINX)
		EnsureFileFromTemplate(path.Join(localDir, "compose.yaml"), templates.LocalCOMPOSE)
		containerFilePath := path.Join(Config().Deploying.Dir, "Containerfile")
		moduleContent := core.ReadFile(path.Join(Config().API.SrcDir, "go.mod"), true) // reads all the go.mod file
		moduleGoVersion := core.Before(core.After(string(moduleContent), "go "), "\n") // keeps only the go version, i.e. 1.24.3
		moduleGoVersion = moduleGoVersion[:strings.LastIndex(moduleGoVersion, ".")]    // keeps only the major & middle numbers, i.e. 1.24
		EnsureFileFromTemplate(containerFilePath, templates.ContainerFILE, moduleGoVersion)

		if verbose {
			Debug("The containerfile used for remote deployment can be used with the following commands (Example): ")
			Debug("podman build -f %s -t %s-test-api . && podman run --rm -it -p %d:55555 %s-test-api",
				containerFilePath, Config().AppNameShort, Config().API.Port, Config().AppNameShort)
			Debug("Test it with: curl http://localhost:%d/rest/translation/fr\\?Namespace\\=Common", Config().API.Port)
		}

		// what we need for remote deployment
		generator := remoteDeploymentGeneratorRegistry.generators[Config().Deploying.RemotePlatform]
		core.PanicMsgIf(generator == nil, "No generator found for remote platform '%s'", Config().Deploying.RemotePlatform)
		generator.generateDeployConfig(core.EnsureDir(Config().Deploying.Dir, "remote"))
	} else {
		slog.Debug("Configuration generation is not required")
	}
}
