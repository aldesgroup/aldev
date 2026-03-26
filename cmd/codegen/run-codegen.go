package codegen

import (
	"fmt"
	"path"
	"time"

	"github.com/aldesgroup/aldev/cmd"
	"github.com/aldesgroup/aldev/utils"
	core "github.com/aldesgroup/corego"
	"github.com/spf13/cobra"
)

// ----------------------------------------------------------------------------
// Command declaration
// ----------------------------------------------------------------------------

// aldevCodegenCmd represents a subcommand
var aldevCodegenCmd = &cobra.Command{
	Use:   "codegen",
	Short: "Completes the app with additional generated code to speed up your dev",
	Long: "This generates additional code to provide you with useful handles on DB lists" +
		"business object classes & their properties, and more. And re-compiles the whole app.",
	Run: aldevCodegenRun,
}

var (
	compilationOnly bool
	noContainer     bool
)

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevCodegenCmd)

	aldevCodegenCmd.Flags().BoolVarP(&compilationOnly, "compilation-only", "c", false, "does only the compilation of the code, no generation step")
	aldevCodegenCmd.Flags().BoolVarP(&noContainer, "no-container", "n", false, "if true, then does not build the binary for containerisation")
}

// ----------------------------------------------------------------------------
// Main logic
// ----------------------------------------------------------------------------

// TODO handle compilation errors = rollbacks on previous situation
func aldevCodegenRun(command *cobra.Command, args []string) {
	start := time.Now()

	// Reading this command's arguments, and reading the aldev YAML config file
	cmd.ReadCommonArgsAndConfig()

	// the contexts to build & codegen Go sources
	buildingCtx := utils.InitAldevContext(100, nil).WithExecDir(utils.GetGoSrcDir()).WithAllowFailure(true)
	codegenCtx := utils.InitAldevContext(100, nil).WithAllowFailure(true)

	// making sure we're applying what's decided in the go.mod file
	must(utils.Run("Making sure we're using the right set of dependencies", buildingCtx, false, "go mod tidy"))

	// control
	if utils.GetBinDir() == "" {
		core.PanicMsg("Aldev config item `.api.bindir` (relative path for the temp folder)  or `.lib.bindir` (if library) is empty!")
	}

	// repeated commands
	execExt := ""
	if core.IsWindows() {
		execExt = ".exe"
	}
	binName := core.PascalToKebab(utils.Config().AppName) + "-api"
	mainCompileCmd := fmt.Sprintf("go build -o %s/%s%s ./main", utils.GetBinDir(), binName, execExt)
	mainRunCmd := fmt.Sprintf("%s/%s%s -config %s -srcdir %s -bindir %s",
		utils.Config().ResolvedBinDir(), binName, execExt,
		path.Join(utils.GetGoSrcDir(), utils.GetConfigPath()),
		utils.GetGoSrcDir(), path.Join(utils.GetGoSrcDir(), utils.GetBinDir()))
	if utils.Config().Web != nil {
		mainRunCmd = fmt.Sprintf("%s -webdir %s", mainRunCmd, utils.Config().Web.SrcDir)
	}
	if utils.Config().Native != nil {
		mainRunCmd = fmt.Sprintf("%s -nativedir %s", mainRunCmd, utils.Config().Native.SrcDir)
	}

	// compilation n°1 - this is needed to have the run command up-to-date
	must(utils.Run("Only compiling & formatting the code", buildingCtx, false, "%s", mainCompileCmd))

	if compilationOnly {
		return
	}

	regenArg := ""
	if utils.IsRegen() {
		regenArg = " -regen"
	}

	// generation step n°1
	must(utils.Run("Generating stuff: DB list, BOclasses, BO registry...", codegenCtx, false, "%s", mainRunCmd+" -codegen 1"+regenArg))

	// compilation n°2
	if codeHasChanged() {
		must(utils.Run("Does it still compile after codegen step 1?", buildingCtx, false, "%s", mainCompileCmd))
	}

	// generation step n°2
	must(utils.Run("Generating stuff: BO specs...", codegenCtx, false, "%s", mainRunCmd+" -codegen 2"+regenArg))

	// compilation n°3
	if codeHasChanged() {
		must(utils.Run("Does it still compile after codegen step 2?", buildingCtx, false, "%s", mainCompileCmd))
	}

	// generation step n°3
	must(utils.Run("Generating stuff: BO vmaps, BO web models, etc...", codegenCtx, false, "%s", mainRunCmd+" -codegen 3"+regenArg))

	// compilation n°4
	if codeHasChanged() {
		must(utils.Run("Does it still compile after codegen step 3?", buildingCtx, false, "%s", mainCompileCmd))
	}

	// formatting
	must(utils.Run("Formatting the code", codegenCtx, false, "gofumpt -w %s %s", path.Join(utils.GetGoSrcDir(), "_include"), path.Join(utils.GetGoSrcDir(), "main")))

	// migrating the DBs if needed
	if !utils.IsDevLibrary() {
		must(utils.Run("DB automigration", codegenCtx, false, "%s", mainRunCmd+" -migrate"))
	}

	// under Windows, the executable for codegen and API serving is not the same - we need to build the executable for the containers
	if core.IsWindows() && !noContainer && codeHasChanged() {
		secondaryCompileCmd := fmt.Sprintf("go build -o %s/%s ./main", utils.GetBinDir(), binName)
		must(utils.Run("Compiling for containerization (Linux)", buildingCtx.WithEnvVars("GOOS=linux"), false, "%s", secondaryCompileCmd))
	}

	// bit of logging
	utils.Info("Aldev codegen done in %s", time.Since(start))
}

// ----------------------------------------------------------------------------
// Utils
// ----------------------------------------------------------------------------

const dirtyFILENAME = "dirty"

func codeHasChanged() bool {
	return string(core.ReadFile(path.Join(utils.GetGoSrcDir(), utils.GetBinDir(), dirtyFILENAME), false)) == "true"
}

func must(result bool) {
	if !result {
		panic("Issue with code compilation or generation!")
	}
}
