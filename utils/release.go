package utils

import (
	"bufio"
	"fmt"
	"log/slog"
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
		slog.Error(fmt.Sprintf("Can only make a release from the 'main' branch (not this '%s' branch)", currentBranch))
		return
	}

	// checking there are no uncommited changes
	if uncommitedChanges := runGitCheckCmd("Checking for uncommited changes", "git status --porcelain"); uncommitedChanges != "" {
		slog.Error("There are uncommited changes, please commit or stash them before making a release")
		return
	}

	// refreshing the remote branches and tags
	if fetchOK := QuickRun("Fetching the remote branches and tags", "git fetch --prune --prune-tags"); !fetchOK {
		core.PanicMsg("Could not fetch the remote branches and tags")
	}

	// checking there are no unpushed commits
	if unpushedCommits := runGitCheckCmd("Checking for unpushed commits", "git log origin/main..HEAD --oneline"); unpushedCommits != "" {
		slog.Error("There are unpushed commits, please push them before making a release")
		return
	}

	// checking there are no unpushed tags
	if unpushedTags := runGitCheckCmd("Checking for unpushed tags", "git push origin --tags --dry-run"); unpushedTags != "" {
		slog.Error("There are unpushed tags, please push them before making a release")
		return
	}

	// checking there are no unpulled commits
	if unpulledCommits := runGitCheckCmd("Checking for unpulled commits", "git log HEAD..origin/main --oneline"); unpulledCommits != "" {
		slog.Error("There are unpulled commits, please pull them before making a release")
		return
	}

	// checking origin/releases hasn't diverged
	if divergedCommits := runGitCheckCmd("Checking 'main' and 'releases' haven't diverged", "git log origin/main..origin/releases --oneline"); divergedCommits != "" {
		slog.Error("There are commits in the 'releases' branch that are not in the 'main' branch: \n\n" + divergedCommits + "\n\n" +
			"Please update the 'main' branch with these commits before making a release")
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

	currentVersionFromGit := strings.TrimSpace(string(RunAndGet("Getting the current version", ".", false, "svu current")))

	// a bit of a sanity check to make sure the VERSION file is in sync with the git tags
	if currentVersionFromFile != "" && currentVersionFromFile != currentVersionFromGit {
		slog.Error(fmt.Sprintf("VERSION file (%s) is not in sync with git tags (%s), this should never happen!", currentVersionFromFile, currentVersionFromGit))
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
	if addOK := QuickRun("Adding the updated VERSION file to Git", "git add %s", versionFilePath); !addOK {
		core.PanicMsg("Could not add the updated VERSION file to Git")
	}

	// commiting the updated VERSION file
	if commitOK := QuickRun("Committing the updated VERSION file", "git commit -m \"dev: bumped version %s to %s\"", currentVersionFromGit, nextVersion); !commitOK {
		core.PanicMsg("Could not commit the updated VERSION file")
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
