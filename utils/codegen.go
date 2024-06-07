// ----------------------------------------------------------------------------
// Generating code - mainly config files - for an app to work locally &
// when its deployed in various environment
// ----------------------------------------------------------------------------

package utils

import (
	"fmt"
	"path"

	"github.com/aldesgroup/aldev/templates"
)

func GenerateConfigs(cfg *AldevConfig) {
	// making sure the config map is here and up-to-date
	EnsureConfigmap(cfg)

	// making sure some needed files are here: base local deployment
	baseDir := EnsureDir(cfg.Deploying.Dir, "base")
	EnsureFileFromTemplate(cfg, path.Join(baseDir, "kustomization.yaml"), templates.KustomizationBase)
	EnsureFileFromTemplate(cfg, path.Join(baseDir, cfg.AppName+"-api-.yaml"), templates.API)
	EnsureFileFromTemplate(cfg, path.Join(baseDir, cfg.AppName+"-api-lb.yaml"), templates.LB)
	EnsureFileFromTemplate(cfg, path.Join(baseDir, cfg.AppName+"-web.yaml"), templates.Web)

	// docker files
	dockerDir := EnsureDir(cfg.Deploying.Dir, "docker")
	EnsureFileFromTemplate(cfg, path.Join(dockerDir, cfg.AppName+"-local-api-docker"), templates.DockerLocalAPI)
	EnsureFileFromTemplate(cfg, path.Join(dockerDir, cfg.AppName+"-local-web-docker"), templates.DockerLocalWEB)
	EnsureFileFromTemplate(cfg, path.Join(dockerDir, cfg.AppName+"-remote-api-docker"), templates.DockerRemoteAPI)
	EnsureFileFromTemplate(cfg, path.Join(dockerDir, cfg.AppName+"-remote-web-docker"), templates.DockerRemoteWeb)

	// adding overlays
	overlaysDir := EnsureDir(cfg.Deploying.Dir, "overlays")
	addOverlay(cfg, overlaysDir, "dev", nil)
	addOverlay(cfg, overlaysDir, "local", [][]string{
		{"patch-no-web-container.yaml", templates.NoWebContainerPatch},
	})
	addOverlay(cfg, overlaysDir, "sandbox", nil)
	addOverlay(cfg, overlaysDir, "staging", nil)
	addOverlay(cfg, overlaysDir, "production", nil)

	// deployment with Gitlab
	EnsureFileFromTemplate(cfg, ".gitlab-ci.yml", templates.GitlabCI)

	// last but not least, the Tiltfile
	EnsureFileFromTemplate(cfg, "Tiltfile", templates.Tiltfile)

	// list of env vars for the web app
	EnsureFileFromTemplate(cfg, path.Join(cfg.Web.SrcDir, ".env-list"), templates.WebEnvList)
}

// ----------------------------------------------------------------------------
// Utils
// ----------------------------------------------------------------------------

// adding an overlay with its name; each patch should be at least: [0]: the filename, [1]: the template;
// [2], [3], etc, are string format parameters to fill the "%s" placeholders in the template.
func addOverlay(cfg *AldevConfig, overlaysDir, overlayName string, patches [][]string) {
	overlay := EnsureDir(overlaysDir, overlayName)

	// handling the patches at first
	kustomizationPatches := ""
	if patches != nil {
		kustomizationPatches = "\n" + "patches:"
		for _, patch := range patches {
			// adding the patch to the kustomization file
			if len(patch) < 2 {
				Fatal("Patches should be provided as at least 1 filename, and 1 template")
			}
			filename := patch[0]
			template := patch[1]
			kustomizationPatches += "\n" + "  - path: " + filename

			// adding the file, from a template, with potential extra params
			templateParams := []any{}
			for i := 2; i < len(patch); i++ {
				templateParams = append(templateParams, patch[i])
			}
			EnsureFileFromTemplate(cfg, path.Join(overlay, filename), template, templateParams...)
		}
	}

	// writing out the kustomization file, with its namespace resource
	EnsureFileFromTemplate(cfg, path.Join(overlay, "kustomization.yaml"),
		templates.KustomizationOverlay+kustomizationPatches, overlayName, overlayName)
	EnsureFileFromTemplate(cfg, path.Join(overlay, fmt.Sprintf("namespace-%s.yaml", overlayName)),
		templates.NewNamespace, overlayName)
}
