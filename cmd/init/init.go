package update

import (
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aldesgroup/aldev/cmd"
	"github.com/aldesgroup/aldev/templates"
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
	Short: "Initiates a new aldev project",
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
	apiOnly     bool
)

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevInitCmd)
	aldevInitCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "activates the verbose mode")
	aldevInitCmd.Flags().StringVarP(&projectName, "name", "n", "", "the name of the project to create; should not contain any space")
	aldevInitCmd.Flags().BoolVarP(&apiOnly, "apiOnly", "a", false, "no web app part, i.e. pure API mode")
}

// ----------------------------------------------------------------------------
// Main logic
// ----------------------------------------------------------------------------

func aldevInitRun(command *cobra.Command, args []string) {
	// handling the verbosity
	utils.SetVerbose(verbose)

	// checking the name
	if projectName == "" || strings.Contains(projectName, " ") || strings.Contains(projectName, "_") {
		utils.Fatal(nil, "Project name ('%s') must not be empty, or contain any space or _ characters", projectName)
	}

	// checking the context
	if _, exists := utils.DirExists(".git"); exists {
		utils.Fatal(nil, "Cannot run this from an actual Git project")
	}

	// LFG!s
	start := time.Now()

	// checking the environment
	cacheDir := os.Getenv(utils.AldevCacheDirENVVAR)
	if cacheDir == "" {
		utils.Fatal(nil, "The cache directory cannot be empty; Env var '%s' should be set (to '../tmp' for instance)",
			utils.AldevCacheDirENVVAR)
	}
	utils.EnsureDir(nil, cacheDir)

	initURL := os.Getenv(aldevINITxREPO)
	if initURL == "" {
		utils.Fatal(nil, "The cache directory cannot be empty; Env var '%s' should be set (to '../tmp' for instance)", aldevINITxREPO)
	}

	// a context for executing stuff
	ctx := utils.NewBaseContext().WithStdErrWriter(io.Discard).WithExecDir(cacheDir)

	// git-cloning into the cache the template project
	firstSlashIndex := strings.Index(initURL, "/")
	// TODO handle public repos
	utils.Run("git-cloning / caching the '"+initURL+"' repo", ctx, false,
		"git clone git@%s:%s.git %s", initURL[:firstSlashIndex], initURL[firstSlashIndex+1:], projectName)

	// removing the .git folder, and the .gitignore file
	cachedProjDir := path.Join(cacheDir, projectName)
	cachedProjCtx := ctx.WithExecDir(cachedProjDir)
	utils.Run("removing the .git folder", cachedProjCtx, false, "rm -fr .git")
	utils.Run("removing the .gitignore file", cachedProjCtx, false, "rm -f .gitignore")
	utils.Run("removing the README file", cachedProjCtx, false, "rm -f README.md")

	// some replacements to customize the project
	// TODO handle custom api / webapp name,
	// TODO + git target repo instead of just project name, to handle auto git init & push
	utils.ReplaceInFile(nil, path.Join(cachedProjDir, ".aldev.yaml"), map[string]string{"fullstack-app-template": projectName})
	utils.ReplaceInFile(nil, path.Join(cachedProjDir, ".aldev.yaml"), map[string]string{"apionly: false": "apionly: true"})
	utils.ReplaceInFile(nil, path.Join(cachedProjDir, "api", "go.mod"), map[string]string{"/libs/fullstack-app-template": "/web/" + projectName})
	utils.ReplaceInFile(nil, path.Join(cachedProjDir, "api", "main", "1-start.go"), map[string]string{"/libs/fullstack-app-template": "/web/" + projectName})
	utils.ReplaceInFile(nil, path.Join(cachedProjDir, "webapp", "vite.config.ts"), map[string]string{"fullstack-app-template": projectName})
	utils.ReplaceInFile(nil, path.Join(cachedProjDir, "webapp", "src", "routes", "+__root.tsx"), map[string]string{"fullstack-app-template": projectName})

	// pure API mode
	if apiOnly {
		utils.Run("removing the webapp part", cachedProjCtx, false, "rm -fr webapp")
	}

	// moving it
	utils.Run("moving the project", utils.NewBaseContext().WithStdErrWriter(io.Discard), false,
		"mv %s %s", path.Join(cacheDir, projectName), projectName)

	// TODO make it more customizable
	newProjCtx := ctx.WithExecDir(projectName)
	utils.Run("initializing Git", newProjCtx, false, "git init")
	utils.EnsureFileFromTemplate(path.Join(projectName, ".gitignore"), templates.GitIgnore)
	utils.Run("adding the files", newProjCtx, false, "git add .")
	utils.Run("committing the files", newProjCtx, false, "git commit -m \"dev: new aldev project\"") // TODO this fails for now
	utils.Run("pushing the first commit", newProjCtx, false,
		"git push --set-upstream git@git-ext.aldes.com:prodigtools/web/%s.git main", projectName)

	// we're done
	utils.Info("Done initialising an aldev project in %s", time.Since(start))
}
