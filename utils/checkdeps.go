package utils

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	core "github.com/aldesgroup/corego"
)

// Comes from the doc: go help list
type Module struct {
	Path     string  // module path
	Query    string  // version query corresponding to this version
	Update   *Module // available update (with -u)
	Main     bool    // is this the main module?
	Indirect bool    // module is only indirectly needed by main module
	Version  string  // module version
	// Versions   []string   // available module versions
	// Replace    *Module    // replaced by this module
	// Time       *time.Time // time version was created
	// Dir        string     // directory holding local copy of files, if any
	// GoMod      string     // path to go.mod file describing module, if any
	// GoVersion  string     // go version used in module
	// Retracted  []string   // retraction information, if any (with -retracted or -u)
	// Deprecated string     // deprecation message, if any (with -u)
	// Sum        string     // checksum for path, version (as in go.sum)
	// GoModSum   string     // checksum for go.mod (as in go.sum)
	// Origin     any        // provenance of module
	// Reuse      bool       // reuse of old module info is safe
}

// Checks the project's dependencies in all its parts
func CheckDeps(ctx CancelableContext, printDeps bool) bool {
	// results
	var nativeDeps, apiDeps, webDeps string
	var goDep bool

	// checking the outdated deps in parallel
	var wg sync.WaitGroup
	wg.Go(func() { nativeDeps = checkNativeDeps() })
	wg.Go(func() { apiDeps, goDep = checkAPIDeps() })
	wg.Go(func() { webDeps = checkWebappDeps() })
	wg.Wait()

	// dealing with the dependencies - API part
	if apiDeps != "" {
		Error("Some of the API's dependencies are outdated!")
		if printDeps {
			println(apiDeps + "\n")
			if goDep {
				// suggestion about the Go version of the projects
				if GetGoSrcDir() == "." {
					Info("Fix it with 'go get go@latest && go mod tidy', then run 'aldev doctor' again.")
				} else {
					Info("Fix it with 'cd %s && go get go@latest && go mod tidy ; cd ..', then run 'aldev doctor' again.", GetGoSrcDir())
				}
			} else {
				// are we

				// suggestion about the packages used by the project
				if GetGoSrcDir() == "." {
					Info("To fix this, you can try: go get -u -t ./... && go mod tidy")
				} else {
					Info("To fix this, you can try: cd %s && go get -u -t ./... && go mod tidy ; cd ..", GetGoSrcDir())
				}
			}
		}
	}

	// dealing with the dependencies - web app part
	if webDeps != "" {
		Error("Some of the web app's dependencies are outdated!")
		if printDeps {
			println(webDeps)
			Warn("But for more control, you may run 'ncu --format group -i' instead.")
		}
	}

	// dealing with the dependencies - native part
	if nativeDeps != "" {
		Error("Some of native app's dependencies are outdated!")
		if printDeps {
			println(nativeDeps)
			Warn("1) But for more control, you must run, instead, the following command:")
			println(fmt.Sprintf("cd %s && ncu --format group -i --reject \"react,react-native*,@react-native*,react-test-renderer\" ; cd ..", Config().Native.SrcDir))
			Warn("2) In case this fails, rollback, and try changing your choices of updates:")
			println(fmt.Sprintf("git checkout %s/package.json %s/package-lock.json", Config().Native.SrcDir, Config().Native.SrcDir))
			Warn("3) Set the React Native version you want to update to in package.json > rnx-kit > alignDeps > requirements, then run:")
			println(fmt.Sprintf("cd %s && npx @rnx-kit/align-deps --write && npm i ; cd ..", Config().Native.SrcDir))
			Warn("4) If this fails, redo 3) with a different version of React Native")
			Warn("5) If 4) succeeds, apply all the changes suggested here: https://react-native-community.github.io/upgrade-helper/")
			Warn("6) 'cd %s/ios && pod install --repo-update && cd ../..' and / or 'cd %s/android && ./gradlew clean && cd ../..'", Config().Native.SrcDir, Config().Native.SrcDir)
			Warn("7) Test the app, and test, test, and test it again")
		}
	}

	// returning, for callersgo list
	return nativeDeps != "" || apiDeps != "" || webDeps != ""
}

// Checking the API's dependencies
func checkAPIDeps() (string, bool) {
	if IsDevAPI() || IsDevLibrary() {
		// controlling the Go version first
		goVersion := string(RunAndGet("Checking Go version", GetGoSrcDir(), true, "%s", "go list -mod=mod -m -u go"))
		if strings.HasSuffix((strings.TrimSpace(goVersion)), "]") {
			return goVersion, true
		}

		// this list all the dependencies, not only the outdated ones
		allDepsString := string(RunAndGet("Checking the API's deps", GetGoSrcDir(), false, "%s", "go list -mod=mod -u -m -json all"))
		allDepsString = "[" + strings.ReplaceAll(allDepsString, "\n", "") + "]"
		allDepsString = strings.ReplaceAll(allDepsString, "}{", "},{")

		// getting through all the dependencies, and only considering the direct, outdated ones
		allDeps := []*Module{}
		outdated := []*Module{}
		core.PanicMsgIfErr(json.Unmarshal([]byte(allDepsString), &allDeps), "Could not JSON-unmarshal the dependencies")
		for _, dep := range allDeps {
			if !dep.Main && !dep.Indirect && dep.Update != nil {
				outdated = append(outdated, dep)
			}
		}

		// let's only keep track of the outdated deps
		if len(outdated) > 0 {
			outdatedBytes, errJson := json.MarshalIndent(outdated, "", "  ")
			core.PanicMsgIfErr(errJson, "Could not JSON-marshal the outdated modules")
			return string(outdatedBytes), false
		}
	}

	return "", false
}

// Checking the web app's dependencies
func checkWebappDeps() string {
	if IsDevWebApp() {
		return string(RunAndGet("Checking the web app's deps", Config().Web.SrcDir, false, "%s", "ncu --format group"))
	}

	return ""
}

// Checking the native app's dependencies
func checkNativeDeps() string {
	if IsDevNative() {
		result := string(RunAndGet("Checking the native app's deps", Config().Native.SrcDir, false, "%s", "ncu --format group"))

		keptLib := false
		keptLines := []string{}

		for _, line := range strings.Split(result, "\n") {
			if line = strings.TrimSpace(line); strings.Contains(line, "  →  ") {
				if libName := strings.SplitN(line, " ", 2)[0]; !core.InSlice(Config().Native.IgnoreOutdated, libName) {
					keptLib = true
					keptLines = append(keptLines, line)
				}
			} else {
				keptLines = append(keptLines, line)
			}
		}

		if keptLib {
			return strings.Join(keptLines, "\n")
		}
	}

	return ""
}
