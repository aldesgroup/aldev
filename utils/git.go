// ----------------------------------------------------------------------------
// Handling Git hooks & more
// ----------------------------------------------------------------------------

package utils

import "github.com/aldesgroup/aldev/templates"

const (
	tagHOTSWAPPED = "___HOTSWAPPED___"
)

// Installing Git hooks to unify our practices
func InstallGitHooks(ctx CancelableContext, cfg *AldevConfig) {
	file := ".git/hooks/pre-commit"
	EnsureFileFromTemplate(cfg, file, templates.GitHookPRECOMMIT, tagHOTSWAPPED)
	Run("Activating the pre-commit hook", ctx, false, "chmod +x %s", file)
}
