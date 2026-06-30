package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	core "github.com/aldesgroup/corego"
)

// ----------------------------------------------------------------------------
// Constants, global variables & types
// ----------------------------------------------------------------------------

type ReleaseType int

const (
	ReleasePatch ReleaseType = iota
	ReleaseMinor
	ReleaseMajor
)

const versionFilePath = "VERSION"

// ----------------------------------------------------------------------------
// Main logic
// ----------------------------------------------------------------------------

func askConfirm(prompt string) bool {
	fmt.Printf("%s (y/N): ", prompt)
	line, errRead := bufio.NewReader(os.Stdin).ReadString('\n')
	core.PanicIfErr(errRead)
	return strings.EqualFold(strings.TrimSpace(line), "y")
}

func runGitCheckCmd(whyRunThis string, commandAsString string) string {
	return strings.TrimSpace(string(RunAndGet(whyRunThis, ".", false, "%s", commandAsString)))
}

func MakeRelease(ctx CancelableContext, release ReleaseType) {
	// can't do this from any other branch than the main branch
	if currentBranch := runGitCheckCmd("Getting the current Git branch", "git branch --show-current"); currentBranch != "main" {
		Error("Can only make a release from the 'main' branch (not this '%s' branch)", currentBranch)
		return
	}

	// checking there are no uncommited changes
	if uncommitedChanges := runGitCheckCmd("Checking for uncommited changes", "git status --porcelain"); uncommitedChanges != "" {
		Error("There are uncommited changes, please commit or stash them before making a release")
		return
	}

	// refreshing the remote branches and tags
	if fetchOK := QuickRun("Fetching the remote branches and tags", "git fetch --prune --prune-tags"); !fetchOK {
		core.PanicMsg("Could not fetch the remote branches and tags")
	}

	// checking there are no unpushed commits
	if unpushedCommits := runGitCheckCmd("Checking for unpushed commits", "git log origin/main..HEAD --oneline"); unpushedCommits != "" {
		Error("There are unpushed commits, please push them before making a release")
		return
	}

	// checking there are no unpushed tags
	if unpushedTags := runGitCheckCmd("Checking for unpushed tags", "git push origin --tags --dry-run"); unpushedTags != "" {
		Error("There are unpushed tags, please push them before making a release")
		return
	}

	// checking there are no unpulled commits
	if unpulledCommits := runGitCheckCmd("Checking for unpulled commits", "git log HEAD..origin/main --oneline"); unpulledCommits != "" {
		Error("There are unpulled commits, please pull them before making a release")
		return
	}

	// checking origin/releases hasn't diverged
	if divergedCommits := runGitCheckCmd("Checking 'main' and 'releases' haven't diverged", "git log origin/main..origin/releases --oneline"); divergedCommits != "" {
		Error("There are commits in the 'releases' branch that are not in the 'main' branch: \n\n%s\n\n"+
			"Please update the 'main' branch with these commits before making a release", divergedCommits)
		return
	}

	// retrieving the current version from the VERSION file
	currentVersionFromFile := string(core.ReadFile(versionFilePath, false))

	// retrieving the current version using the git tags
	var releaseStr string
	switch release {
	case ReleasePatch:
		releaseStr = "patch"
	case ReleaseMinor:
		releaseStr = "minor"
	case ReleaseMajor:
		releaseStr = "major"
	}

	// getting the current version
	currentVersionFromGit := strings.TrimSpace(string(RunAndGet("Getting the current version", ".", false, "svu current")))

	// a bit of a sanity check to make sure the VERSION file is in sync with the git tags
	if currentVersionFromFile != "" && currentVersionFromFile != currentVersionFromGit {
		Error("VERSION file (%s) is not in sync with git tags (%s), this should never happen!", currentVersionFromFile, currentVersionFromGit)
		return
	}

	// computing the next desired version
	nextVersion := strings.TrimSpace(string(RunAndGet("Getting the next version", ".", false, "svu %s", releaseStr)))

	// asking for confirmation before making the release
	if !askConfirm(fmt.Sprintf("Do you want to go from %s to %s?", currentVersionFromGit, nextVersion)) {
		println("\nRelease cancelled")
		return
	}

	// new Git tag
	if newTagOK := QuickRun("New git tag", "git tag %s", nextVersion); !newTagOK {
		core.PanicMsg("Weird... Somehow we failed to create new Git tag '%s'", nextVersion)
	}

	// updating (maybe creating) the VERSION file
	core.WriteStringToFile(versionFilePath, "%s", nextVersion)

	// updating the API doc
	if Config().API != nil && Config().API.Doc.Path != "" {
		core.ReplaceInFile(Config().API.Doc.Path, map[string]string{"v0.0.1": nextVersion})
		// core.ReplaceInFile(Config().API.Doc.Report, map[string]string{"v0.0.1": nextVersion})
	}

	// git-adding the VERSION file
	if Config().API != nil && Config().API.Doc.Path != "" {
		// if addOK := QuickRun("Adding the updated VERSION file to Git + API docs", "git add %s %s %s", versionFilePath, Config().API.Doc.Path, Config().API.Doc.Report); !addOK {
		if addOK := QuickRun("Adding the updated VERSION file to Git + API docs", "git add %s %s %s", versionFilePath, Config().API.Doc.Path); !addOK {
			core.PanicMsg("Could not add the updated VERSION file + API docs to Git")
		}
	} else if addOK := QuickRun("Adding the updated VERSION file to Git + API docs", "git add %s", versionFilePath); !addOK {
		core.PanicMsg("Could not add the updated VERSION file + API docs to Git")

	}

	// commiting the updated VERSION file
	if commitOK := QuickRun("Committing the updated VERSION file + API docs", "git commit -m \"dev: bumped version %s to %s\"", currentVersionFromGit, nextVersion); !commitOK {
		core.PanicMsg("Could not commit the updated VERSION file + API docs to Git")
	}

	// pushing to the main branch
	if pushOK := QuickRun("Pushing to the remote 'main' branch", "git push"); !pushOK {
		core.PanicMsg("Could not push to the remote 'main' branch")
	}

	// pushing the new tag
	if pushOK := QuickRun("Pushing the new tag", "git push --tags"); !pushOK {
		core.PanicMsg("Could not push the new tag '%s'", nextVersion)
	}

	// pushing to the releases branch
	if pushOK := QuickRun("Pushing to the remote 'releases' branch", "git push origin HEAD:releases"); !pushOK {
		core.PanicMsg("Could not push to the remote 'releases' branch")
	}
}
