// ----------------------------------------------------------------------------
// Generating code - mainly config files - for an app to work locally &
// when its deployed in various environment
// ----------------------------------------------------------------------------

package utils

import (
	"path"

	"github.com/aldesgroup/aldev/templates"
	core "github.com/aldesgroup/corego"
)

func GenerateDeployFiles(ctx CancelableContext) {
	// what we need for local deployment
	localDir := core.EnsureDir(Config().Deploying.Dir, "local")
	EnsureFileFromTemplate(path.Join(localDir, "nginx.conf"), templates.LocalNGINX)
	EnsureFileFromTemplate(path.Join(localDir, "compose.yaml"), templates.LocalCOMPOSE)
}
