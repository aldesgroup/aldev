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

func GenerateDeployConfigs(ctx CancelableContext) {

	// making sure the config map is here and up-to-date
	EnsureConfigmap()

	// making sure some needed files are here: base local deployment
	baseDir := EnsureDir(nil, Config().Deploying.Dir, "base")
	EnsureFileFromTemplate(path.Join(baseDir, Config().AppName+"-api-.yaml"), templates.API)
	EnsureFileFromTemplate(path.Join(baseDir, Config().AppName+"-api-lb.yaml"), templates.LB)
	if IsDevWebApp() {
		EnsureFileFromTemplate(path.Join(baseDir, "kustomization.yaml"), templates.KustomizationBaseComplete)
		EnsureFileFromTemplate(path.Join(baseDir, Config().AppName+"-web.yaml"), templates.Web)
	} else {
		EnsureFileFromTemplate(path.Join(baseDir, "kustomization.yaml"), templates.KustomizationBaseAPI)
	}

	// docker files
	dockerDir := EnsureDir(nil, Config().Deploying.Dir, "docker")
	EnsureFileFromTemplate(path.Join(dockerDir, Config().AppName+"-local-api-docker"), templates.DockerLocalAPI)
	EnsureFileFromTemplate(path.Join(dockerDir, Config().AppName+"-remote-api-docker"), templates.DockerRemoteAPI)
	if IsDevWebApp() {
		EnsureFileFromTemplate(path.Join(dockerDir, Config().AppName+"-local-web-docker"), templates.DockerLocalWEB)
		EnsureFileFromTemplate(path.Join(dockerDir, Config().AppName+"-remote-web-docker"), templates.DockerRemoteWeb)
	}

	// adding overlays
	overlaysDir := EnsureDir(nil, Config().Deploying.Dir, "overlays")
	addOverlay(overlaysDir, "dev", nil)
	if IsDevWebApp() {
		addOverlay(overlaysDir, "local", [][]string{{"patch-no-web-container.yaml", templates.NoWebContainerPatch}})
	} else {
		addOverlay(overlaysDir, "local", nil)
	}
	addOverlay(overlaysDir, "sandbox", nil)
	addOverlay(overlaysDir, "staging", nil)
	addOverlay(overlaysDir, "production", nil)

	// deployment with Gitlab
	// EnsureFileFromTemplate(".gitlab-ci.yml", templates.GitlabCI)

	// last but not least, the Tiltfile
	tiltfileTemplate := templates.TiltfileAPI
	if IsDevWebApp() {
		tiltfileTemplate += templates.TiltfileWebPart
	}
	if IsDevNative() {
		tiltfileTemplate += templates.TiltfileNativePart
	}
	EnsureFileFromTemplate("Tiltfile", tiltfileTemplate)

	// list of env vars for the web app
	if IsDevWebApp() {
		EnsureFileFromTemplate(path.Join(Config().Web.SrcDir, ".env-list"), templates.WebEnvList)
	}
}

// ----------------------------------------------------------------------------
// Utils
// ----------------------------------------------------------------------------

// adding an overlay with its name; each patch should be at least: [0]: the filename, [1]: the template;
// [2], [3], etc, are string format parameters to fill the "%s" placeholders in the template.
func addOverlay(overlaysDir, overlayName string, patches [][]string) {
	overlay := EnsureDir(nil, overlaysDir, overlayName)

	// handling the patches at first
	kustomizationPatches := ""
	if patches != nil {
		kustomizationPatches = "\n" + "patches:"
		for _, patch := range patches {
			// adding the patch to the kustomization file
			if len(patch) < 2 {
				Fatal(nil, "Patches should be provided as at least 1 filename, and 1 template")
			}
			filename := patch[0]
			template := patch[1]
			kustomizationPatches += "\n" + "  - path: " + filename

			// adding the file, from a template, with potential extra params
			templateParams := []any{}
			for i := 2; i < len(patch); i++ {
				templateParams = append(templateParams, patch[i])
			}
			EnsureFileFromTemplate(path.Join(overlay, filename), template, templateParams...)
		}
	}

	// writing out the kustomization file, with its namespace resource
	EnsureFileFromTemplate(path.Join(overlay, "kustomization.yaml"),
		templates.KustomizationOverlay+kustomizationPatches, overlayName, overlayName)
	EnsureFileFromTemplate(path.Join(overlay, fmt.Sprintf("namespace-%s.yaml", overlayName)),
		templates.NewNamespace, overlayName)
}
