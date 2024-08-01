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
		pcFile := ".git/hooks/pre-commit"
		EnsureFileFromTemplate(cfg, pcFile, templates.GitHookPRECOMMIT, TagHOTSWAPPED)
		Run("Activating the pre-commit hook", ctx, false, "chmod +x %s", pcFile)
		cmFile := ".git/hooks/commit-msg"
		EnsureFileFromTemplate(cfg, cmFile, templates.GitHookCOMMITMSG)
		Run("Activating the commit-msg hook", ctx, false, "chmod +x %s", cmFile)
	}
}

type versionObject struct {
	Commit string
	Value  string
}
