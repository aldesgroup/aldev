package update

import (
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aldesgroup/aldev/cmd"
	"github.com/aldesgroup/aldev/utils"
	"github.com/spf13/cobra"
)

// ----------------------------------------------------------------------------
// Command declaration
// ----------------------------------------------------------------------------

const aldevINITxREPO = "ALDEV_INITREPO"

// aldevInitCmd represents a subcommand
var aldevInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initiates a new aldev projects",
	Long: "Depends on an environment variable " + aldevINITxREPO + ", that should point to a git-clonable project " +
		"containing a kind of template for aldev projects, i.e. a project with a very generic aldev config file " +
		"(.aldev.yaml), a bit of API & web app code, that should work out of the box with the 'aldev' command. " +
		"e.g.: github.com/mydomain/myprojecttemplate",
	Run: aldevInitRun,
}

var (
	// cfgFileName  string
	verbose     bool
	projectName string
)

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevInitCmd)
	aldevInitCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "activates the verbose mode")
	aldevInitCmd.Flags().StringVarP(&projectName, "name", "n", "", "the name of the project to create; should not contain any space")
}

// ----------------------------------------------------------------------------
// Main logic
// ----------------------------------------------------------------------------

func aldevInitRun(command *cobra.Command, args []string) {
	// it's only here that we have this variable valued
	if verbose {
		utils.SetVerbose()
	}

	// checking the name
	if projectName == "" || strings.Contains(projectName, " ") || strings.Contains(projectName, "_") {
		utils.Fatal("Project name ('%s') must not be empty, or contain any space or _ characters", projectName)
	}

	// checking the context
	if _, exists := utils.DirExists(".git"); exists {
		utils.Fatal("Cannot run this from an actual Git project")
	}

	// LFG
	start := time.Now()

	// checking the environment
	cacheDir := os.Getenv(utils.AldevCacheDirENVVAR)
	if cacheDir == "" {
		utils.Fatal("The cache directory cannot be empty; Env var '%s' should be set (to '../tmp' for instance)",
			utils.AldevCacheDirENVVAR)
	}
	utils.EnsureDir(cacheDir)

	initURL := os.Getenv(aldevINITxREPO)
	if initURL == "" {
		utils.Fatal("The cache directory cannot be empty; Env var '%s' should be set (to '../tmp' for instance)", aldevINITxREPO)
	}

	// a context for executing stuff
	ctx := utils.NewBaseContext().WithStdErrWriter(io.Discard).WithExecDir(cacheDir)

	// git-cloning into the cache the template project
	firstSlashIndex := strings.Index(initURL, "/")
	utils.Run("git-cloning / caching the '"+initURL+"' repo", ctx, false,
		"git clone git@%s:%s.git %s", initURL[:firstSlashIndex], initURL[firstSlashIndex+1:], projectName)

	// removing the .git folder, and the .gitignore file
	projDir := path.Join(cacheDir, projectName)
	cachedProjCtx := ctx.WithExecDir(projDir)
	utils.Run("removing the .git folder", cachedProjCtx, false, "rm -fr .git")
	utils.Run("removing the .gitignore file", cachedProjCtx, false, "rm -f .gitignore")

	// some replacements to customize the project
	utils.ReplaceInFile(path.Join(projDir, ".aldev.yaml"), map[string]string{"fullstack-app-template": projectName})
	utils.ReplaceInFile(path.Join(projDir, "README.md"), map[string]string{"# fullstack-app-template": "# " + projectName})
	utils.ReplaceInFile(path.Join(projDir, "api", "go.mod"), map[string]string{"/libs/fullstack-app-template": "/web/" + projectName})
	utils.ReplaceInFile(path.Join(projDir, "api", "main", "1-start.go"), map[string]string{"/libs/fullstack-app-template": "/web/" + projectName})
	utils.ReplaceInFile(path.Join(projDir, "webapp", "vite.config.ts"), map[string]string{"fullstack-app-template": projectName})
	utils.ReplaceInFile(path.Join(projDir, "webapp", "src", "routes", "+__root.tsx"), map[string]string{"fullstack-app-template": projectName})

	// moving it
	utils.Run("moving the project", utils.NewBaseContext().WithStdErrWriter(io.Discard), false,
		"mv %s %s", path.Join(cacheDir, projectName), projectName)

	// we're done
	utils.Info("Done initialising an aldev project in %s", time.Since(start))
}
