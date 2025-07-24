package bootstrap

import (
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aldesgroup/aldev/cmd"
	"github.com/aldesgroup/aldev/utils"
	core "github.com/aldesgroup/corego"
	"github.com/spf13/cobra"
)

// ----------------------------------------------------------------------------
// Command declaration
// ----------------------------------------------------------------------------

const genericTemplate = "github.com/aldesgroup/devotion-project-template"

// aldevBootstrapCmd represents a subcommand
var aldevBootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstraps a new Devotion project",
	Long: "Using a link to a Devotion project template, and depending on given options, " +
		"this creates a new Devotion project with an API part and / or a WEB app " +
		"and / or a NATIVE app that can then be run & developed with the `aldev` command." +
		"The default `aldev boostrap -MyNewProject` command creates a running web app + API, " +
		"using the generic " + genericTemplate,
	Run: aldevBootstrapRun,
}

var (
	verbose           bool
	projectNamePascal string
	templateLink      string
	templatePrivate   bool
	noAPI             bool
	noWeb             bool
	native            bool
)

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevBootstrapCmd)
	aldevBootstrapCmd.Example = "To generate a new Devotion native / mobile app (in verbose mode): " +
		"aldev bootstrap -verbose --mobile --name MyDevotionApp; " +
		"or aldev bootstrap -v -m -n MyDevotionApp"
	aldevBootstrapCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "activates the verbose mode")
	aldevBootstrapCmd.Flags().StringVarP(&projectNamePascal, "name", "n", "", "the name of the project to create; should not contain any space, and be in PascalCase")
	aldevBootstrapCmd.Flags().StringVarP(&templateLink, "template", "t", genericTemplate, "a link to a Git repo containing a Devotion project template, e.g. "+genericTemplate)
	aldevBootstrapCmd.Flags().BoolVarP(&templatePrivate, "private", "p", false, "is the project template being used private?")
	aldevBootstrapCmd.Flags().BoolVarP(&noAPI, "api-less", "a", false, "does not create the API part")
	aldevBootstrapCmd.Flags().BoolVarP(&noWeb, "web-less", "w", false, "does not create the Web part")
	aldevBootstrapCmd.Flags().BoolVarP(&native, "mobile", "m", false, "create the native (mobile) part")
}

// ----------------------------------------------------------------------------
// Main logic
// ----------------------------------------------------------------------------

func aldevBootstrapRun(command *cobra.Command, args []string) {
	// handling the base execution parameter
	utils.SetVerbose(verbose)
	utils.SetCacheDir(core.IfThenElse(utils.GetCacheDir() != "", utils.GetCacheDir(), "../tmp"))

	// checking the Project's name, and that it does not exist yet
	if projectNamePascal == "" || strings.Contains(projectNamePascal, " ") || strings.Contains(projectNamePascal, "_") {
		core.PanicMsg("Project name ('%s') must not be empty, or contain any space or _ characters", projectNamePascal)
	}
	projectNameKebab := core.PascalToKebab(projectNamePascal)
	if projectNamePascal != core.KebabToPascal(projectNameKebab) {
		core.PanicMsg("Project name ('%s') should be in PascalCase!", projectNamePascal)
	}
	if core.DirExists(projectNameKebab) {
		core.PanicMsg("There's already a project '%s' ('%s') here!", projectNamePascal, projectNameKebab)
	}

	// a context for executing stuff, and recovery
	out := io.Discard
	if verbose {
		out = os.Stdout
	}
	ctx := utils.NewBaseContext().WithStdErrWriter(out).WithStdOutWriter(out).WithExecDir(utils.GetCacheDir())
	defer utils.Recover(ctx, "Bootstrapping a new app called '%s' (project '%s')", projectNamePascal, projectNameKebab)

	// checking the context
	if core.DirExists(".git") {
		core.PanicMsg("Cannot run this from an actual Git project")
	}

	// LFG!
	start := time.Now()

	// checking the environment
	core.EnsureDir(utils.GetCacheDir())
	core.EnsureNoDir(utils.GetCacheDir(), projectNameKebab)

	// git-cloning into the cache the template project
	if templatePrivate {
		firstSlashIndex := strings.Index(templateLink, "/")
		utils.Run("git-cloning / caching the private '"+templateLink+"' repo", ctx, false,
			"git clone git@%s:%s.git %s", templateLink[:firstSlashIndex], templateLink[firstSlashIndex+1:], projectNameKebab)
	} else {
		utils.Run("git-cloning / caching the public '"+templateLink+"' repo", ctx, false,
			"git clone https://%s %s", templateLink, projectNameKebab)
	}

	// removing the .git folder, and the .gitignore file
	cachedProjDir := path.Join(utils.GetCacheDir(), projectNameKebab)
	cachedProjCtx := ctx.WithExecDir(cachedProjDir)
	utils.Run("removing the .git folder", cachedProjCtx, false, "%s", core.RemoveCmd()+" .git")

	// some replacements to customize the project
	// TODO handle custom api / webapp name,ls

	// TODO + git target repo instead of just project name, to handle auto git init & push
	// core.ReplaceInFile(path.Join(cachedProjDir, ".aldev.yaml"), map[string]string{"devotion--template": projectNamePascal})
	// core.ReplaceInFile(path.Join(cachedProjDir, ".aldev.yaml"), map[string]string{"apionly: false": "apionly: true"})
	// core.ReplaceInFile(path.Join(cachedProjDir, "api", "go.mod"), map[string]string{"/libs/devotion--template": "/web/" + projectNamePascal})
	// core.ReplaceInFile(path.Join(cachedProjDir, "api", "main", "1-start.go"), map[string]string{"/libs/devotion--template": "/web/" + projectNamePascal})
	// core.ReplaceInFile(path.Join(cachedProjDir, "webapp", "vite.config.ts"), map[string]string{"devotion--template": projectNamePascal})
	// core.ReplaceInFile(path.Join(cachedProjDir, "webapp", "src", "routes", "+__root.tsx"), map[string]string{"devotion--template": projectNamePascal})
	// TODO replaces in README
	// TODO replaces in .gitignore

	// // pure API mode
	// if apiOnly {
	// 	utils.Run("removing the webapp part", cachedProjCtx, false, "rm -fr webapp")
	// }

	// // TODO make it more customizable
	// newProjCtx := ctx.WithExecDir(projectName)
	// utils.Run("initializing Git", newProjCtx, false, "git init")
	// utils.EnsureFileFromTemplate(path.Join(projectName, ".gitignore"), templates.GitIgnore)
	// utils.Run("adding the files", newProjCtx, false, "git add .")
	// utils.Run("committing the files", newProjCtx, false, "git commit -m \"dev: new aldev project\"") // TODO this fails for now
	// utils.Run("pushing the first commit", newProjCtx, false,
	// 	"git push --set-upstream git@git-ext.aldes.com:prodigtools/web/%s.git main", projectName)

	// trimming the clone as much as necessary, before moving it
	if noAPI {
		utils.Run("removing the API part", cachedProjCtx, false, "%s", core.RemoveCmd()+" api")
	}
	if noWeb {
		utils.Run("removing the Web part", cachedProjCtx, false, "%s", core.RemoveCmd()+" web")
	}
	if !native {
		utils.Run("removing the native part", cachedProjCtx, false, "%s", core.RemoveCmd()+" native")
	}

	// moving it
	utils.Run("moving the project", ctx.WithExecDir("."), false, core.MoveCmd()+" %s %s", path.Join(utils.GetCacheDir(), projectNameKebab), projectNameKebab)

	// tweaking the config
	configFilePath := path.Join(projectNameKebab, ".aldev.yaml")
	if !noAPI {
		core.ReplaceInFile(configFilePath, map[string]string{"#IFAPI ": ""})
	}
	if !noWeb {
		core.ReplaceInFile(configFilePath, map[string]string{"#IFWEB ": ""})
	}
	if native {
		core.ReplaceInFile(configFilePath, map[string]string{"#IFNATIVE ": ""})
	}

	// simple reading of the config - for real config loading, use utils.ReadConfig
	conf := core.ReadFileFromYAML(configFilePath, &utils.AldevConfig{}, true)

	// customising each app part of the project
	if native {
		bootstrapNativeApp(ctx.WithExecDir(projectNameKebab), conf)
	}

	// we're done
	utils.Info("Done initialising an aldev project in %s", time.Since(start))
}

// bootstraps the native app of the project
func bootstrapNativeApp(ctx utils.CancelableContext, conf *utils.AldevConfig) {
	nativeTemplate := conf.Native.SrcDir + "_template"
	utils.Run("renaming the native folder from the template", ctx, false, core.MoveCmd()+" %s %s", conf.Native.SrcDir, nativeTemplate)

	// utils.Run("removing the old installation if any", ctx, false, "%s", "npm uninstall -g react-native-cli @react-native-community/cli")
	// utils.Run("init a new React Native project", ctx, false, "npx @react-native-community/cli@latest init %s", projectNamePascal)
	// - [ ] npm install:
	//   - npm i @react-navigation/bottom-tabs @react-navigation/native-stack react-native-gesture-handler react-native-screens
	//   - npm i react-native-orientation-locker
	//   - npm i jotai form-atoms @react-native-async-storage/async-storage
	//   - npm i nativewind react-native-reanimated react-native-safe-area-context@5.4.0 tailwindcss-animate lucide-react-native class-variance-authority tailwind-merge clsx
	//   - npm i i18next react-i18next react-native-localize
	//   - npm i react-native-fs react-native-html-to-pdf react-native-share
	//   - npm i react-native-config
	//   - npm i react-native-ble-manager react-native-permissions => only if using the BLE
	//   - npm i react-native-svg
	//   - npm i async-mutex
	//TODO   - npm i react-native-reanimated-carousel @react-native-community/slider @rn-primitives/slot @rn-primitives/switch

	// - [ ] npm install --save-dev:
	//   - npm i -D @rnx-kit/metro-resolver-symlinks
	//   - npm i -D prettier@^3.0 tailwindcss@^3.4.17 prettier-plugin-tailwindcss@^0.5.11
	//   - npm i -D babel-plugin-module-resolver
	//   - npm i -D @types/react-native-html-to-pdf => only if printing PDF

	// replace names in:
	// - package.json

	// addings:
	// android/app/build.gradle :

	// Make sure a virtual device is ready : Android Studio > three dots menu > Virtual Device Manager > start 1 device (play button)
	// adb devices should return something

	// Auto-commit after the installation, before the adding of

	// aldev refresh -v

	// TODO : document : how to prepare a release, with build.gradle, a key etc : better to give a link

	// TODO : Android main Manifest stuff : we can copy for now
}
