// ----------------------------------------------------------------------------
// The code here is responsible for fetching source code and copy-pasting it
// into our project, as configured
// ----------------------------------------------------------------------------
package utils

import (
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	core "github.com/aldesgroup/corego"
)

const (
	versionFILENAME = "version.json"
)

// Fetching the required vendored libraries
func fetchVendoredLibraries(ctx CancelableContext) {
	// making sure we recover any big crashing error
	defer Recover(ctx, "fetching / refreshing the vendors")

	// LFG
	start := time.Now()

	// fetching / refreshing all the vendors in parallel
	wg := new(sync.WaitGroup)
	for _, vendor := range Config().Vendors {
		wg.Add(1)
		go func(v *VendorConfig) {
			defer wg.Done()
			fetchVendor(ctx, v)
		}(vendor)
	}
	wg.Wait()

	// we're done
	Info("Done fetching / refreshing the vendors in %s", time.Since(start))
}

func fetchVendor(ctx CancelableContext, vendor *VendorConfig) {
	// making sure we recover any big crashing error
	defer Recover(ctx, "fetching / refreshing vendor '%s'", vendor)

	// getting the repo name
	repoName := path.Base(vendor.Repo)

	// checking if vendor already exists or not
	repoCachePath := path.Join(GetCacheDir(), repoName)
	repoExistsInCache := core.DirExists(repoCachePath)

	// if it exists, git pulling within it
	if repoExistsInCache {
		Run("Ensuring the main branch in the '"+repoName+"' repo",
			NewBaseContext().WithStdErrWriter(io.Discard).WithExecDir(repoCachePath).WithStdOutWriter(io.Discard),
			false,
			"git checkout main")

		Run("refreshing the cached '"+repoName+"' repo",
			NewBaseContext().WithStdErrWriter(io.Discard).WithExecDir(repoCachePath).WithStdOutWriter(io.Discard),
			false,
			"git pull")

	} else { // if not, git clone it into temp folder
		firstSlashIndex := strings.Index(vendor.Repo, "/")
		Run("git-cloning / caching the '"+repoName+"' repo",
			NewBaseContext().WithStdErrWriter(io.Discard).WithExecDir(GetCacheDir()),
			false,
			"git clone git@%s:%s.git", vendor.Repo[:firstSlashIndex], vendor.Repo[firstSlashIndex+1:]) // TODO handle https for public repos
	}

	// switch branch, if one if specified
	if vendor.Branch != "" {
		Run("checking out the '"+vendor.Branch+"' branch",
			NewBaseContext().WithStdErrWriter(io.Discard).WithExecDir(repoCachePath),
			false,
			"git checkout %s", vendor.Branch)
		Run("refreshing the branch's code",
			NewBaseContext().WithStdErrWriter(io.Discard).WithExecDir(repoCachePath).WithStdOutWriter(io.Discard),
			false,
			"git pull")
	}

	// get the latest version
	allVersions := strings.Split(string(RunAndGet("Getting the latest version", repoCachePath, false, "git tag -l --sort -version:refname")), "\n")
	latestVersion := &versionObject{Value: allVersions[0], Commit: lastCommit(repoCachePath, core.IfThenElse(vendor.Branch != "", vendor.Branch, "main"))}

	// the target directory
	vendorDir := core.EnsureDir(vendor.To, repoName)

	// checking the current version
	versionFileName := path.Join(vendorDir, versionFILENAME)
	currentVersion := core.ReadFileFromJSON(versionFileName, &versionObject{}, false)

	// will there be a next version different from the current one?
	var nextVersion *versionObject
	if vendor.Version == "latest" {
		if currentVersion == nil ||
			latestVersion.Commit != currentVersion.Commit ||
			latestVersion.Value != currentVersion.Value {
			nextVersion = latestVersion
		}
	} else {
		// warning if the required version is not the latest one
		if vendor.Version != latestVersion.Value {
			Error("Required '%s' version '%s' is not the latest available one (%s)!", repoName, vendor.Version, latestVersion.Value)
		}

		// there's some change to do if the current version is not the required one
		if currentVersion == nil || currentVersion.Value != vendor.Version {
			// checking the required version exists!
			if !core.InSlice(allVersions, vendor.Version) {
				core.PanicMsg("Required version '%s' does not exist in project '%s'. Here are the available versions: %s",
					vendor.Version, repoName, strings.Join(allVersions, "\n"))
			}

			// checking out the required version
			Run("checking out the right '"+repoName+"' version",
				NewBaseContext().WithStdErrWriter(io.Discard).WithExecDir(repoCachePath),
				false,
				"git checkout %s", vendor.Version)

			// setting up the next version
			nextVersion = &versionObject{Value: vendor.Version, Commit: lastCommit(repoCachePath, vendor.Version)}
		}
	}

	if nextVersion != nil {
		// removing the previous vendor version first
		Debug("Cleaning '%s' first, if needed", vendor.To)
		core.PanicIfErr(os.RemoveAll(path.Join(vendor.To, repoName)))

		// copying the new vendor code + version file
		QuickRun("Copying this repo into project: "+repoName, "%s %s/%s/. %s", core.CopyCmd(), repoCachePath, vendor.From, vendorDir)
		core.WriteJsonObjToFile(versionFileName, nextVersion)

		// removing the vendor's vendors, if any
		Debug("Removing project %s's vendor folder, if any", repoName)
		core.PanicIfErr(os.RemoveAll(path.Join(vendorDir, "vendor")))

		// bit of logging
		if currentVersion != nil {
			Info("Update for vendor '%s': '%s' (%s) -> %s (%s)", repoName,
				currentVersion.Value, currentVersion.Commit,
				nextVersion.Value, nextVersion.Commit)
		} else {
			Info("Installed vendor '%s': '%s' (%s)", repoName,
				nextVersion.Value, nextVersion.Commit)
		}
	}
}

// get the latest commit
func lastCommit(repoPath string, branchOrTag string) string {
	commitBytes := RunAndGet("Getting the latest commit", repoPath, false, "%s", "git rev-parse "+branchOrTag)
	return string(commitBytes[:len(commitBytes)-1])
}
