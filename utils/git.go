// ----------------------------------------------------------------------------
// Handling Git hooks & more
// ----------------------------------------------------------------------------

package utils

import (
	"github.com/aldesgroup/aldev/templates"
	core "github.com/aldesgroup/corego"
)

const (
	TagHOTSWAPPED = "___HOTSWAPPED___"
)

// Installing Git hooks to unify our practices
func InstallGitHooks(ctx CancelableContext) {
	if core.DirExists(".git") {
		pcFile := ".git/hooks/pre-commit"
		EnsureFileFromTemplate(pcFile, templates.GitHookPRECOMMIT, TagHOTSWAPPED)
		if !core.IsWindows() {
			Run("Activating the pre-commit hook", ctx, false, "chmod +x %s", pcFile)
		}
		cmFile := ".git/hooks/commit-msg"
		EnsureFileFromTemplate(cmFile, templates.GitHookCOMMITMSG)
		if !core.IsWindows() {
			Run("Activating the commit-msg hook", ctx, false, "chmod +x %s", cmFile)
		}
	}
}

type versionObject struct {
	Commit string
	Value  string
}
