// ----------------------------------------------------------------------------
// Generating code - mainly config files - for an app to work locally &
// when its deployed in various environment
// ----------------------------------------------------------------------------

package utils

import (
	"path"
	"strings"

	"github.com/aldesgroup/aldev/templates"
	core "github.com/aldesgroup/corego"
)

func GenerateDeployFiles(ctx CancelableContext) {
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
		Debug("Test it with: curl http://localhost:42069/rest/translation/fr\\?Namespace\\=Common")
	}
}
