// ----------------------------------------------------------------------------
// Handling Git hooks & more
// ----------------------------------------------------------------------------

package utils

import (
	"strings"

	"github.com/aldesgroup/aldev/templates"
	core "github.com/aldesgroup/corego"
)

const (
	TagHOTSWAPPED = "___HOTSWAPPED___"
)

// Doing some Git setup, like installing hooks, to unify our practices across projects
func SetupGit(ctx CancelableContext) {

	// installing hooks
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

	// checking the remote has a "releases" branch, and if not, creating it
	if !core.DirExists(".git") {
		return
	}
	if !releasesBranchExist() {
		Run("Creating the 'releases' branch on the remote", ctx, true, "git push origin HEAD:releases")
	}
}

func releasesBranchExist() bool {
	output := RunAndGet("Getting the list of remote branches", ".", false, "git branch --remote")
	for _, branch := range strings.Split(string(output), "\n") {
		if strings.TrimSpace(branch) == "origin/releases" {
			return true
		}
	}
	return false
}

type versionObject struct {
	Commit string
	Value  string
}
