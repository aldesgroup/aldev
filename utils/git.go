// ----------------------------------------------------------------------------
// Handling Git hooks & more
// ----------------------------------------------------------------------------

package utils

import (
	"github.com/aldesgroup/aldev/templates"
)

const (
	TagHOTSWAPPED = "___HOTSWAPPED___"
)

// Installing Git hooks to unify our practices
func InstallGitHooks(ctx CancelableContext) {
	if _, exists := DirExists(".git"); exists {
		pcFile := ".git/hooks/pre-commit"
		EnsureFileFromTemplate(pcFile, templates.GitHookPRECOMMIT, TagHOTSWAPPED)
		if !IsWindows() {
			Run("Activating the pre-commit hook", ctx, false, "chmod +x %s", pcFile)
		}
		cmFile := ".git/hooks/commit-msg"
		EnsureFileFromTemplate(cmFile, templates.GitHookCOMMITMSG)
		if !IsWindows() {
			Run("Activating the commit-msg hook", ctx, false, "chmod +x %s", cmFile)
		}
	}
}

type versionObject struct {
	Commit string
	Value  string
}
