// ----------------------------------------------------------------------------
// Handling Git hooks & more
// ----------------------------------------------------------------------------

package utils

import "github.com/aldesgroup/aldev/templates"

const (
	TagHOTSWAPPED = "___HOTSWAPPED___"
)

// Installing Git hooks to unify our practices
func InstallGitHooks(ctx CancelableContext, cfg *AldevConfig) {
	if _, exists := DirExists(".git"); exists {
		file := ".git/hooks/pre-commit"
		EnsureFileFromTemplate(cfg, file, templates.GitHookPRECOMMIT, TagHOTSWAPPED)
		Run("Activating the pre-commit hook", ctx, false, "chmod +x %s", file)
	}
}

type versionObject struct {
	Commit string
	Value  string
}
