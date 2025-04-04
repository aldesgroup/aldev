package codegen

import (
	"fmt"
	"path"
	"time"

	"github.com/aldesgroup/aldev/cmd"
	"github.com/aldesgroup/aldev/utils"
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
	regen           bool
	noContainer     bool
)

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevCodegenCmd)

	aldevCodegenCmd.Flags().BoolVarP(&compilationOnly, "compilation-only", "c", false, "does only the compilation of the code, no generation step")
	aldevCodegenCmd.Flags().BoolVarP(&regen, "regen", "r", false, "forces the regeneration")
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

	// the context to build Go sources
	completeCtx := utils.InitAldevContext(100, nil).WithExecDir(utils.GetGoSrcDir())

	// making sure we're applying what's decided in the go.mod file
	utils.Run("Making sure we're using the right set of dependencies", completeCtx, false, "go mod tidy")

	// control
	if utils.GetBinDir() == "" {
		utils.Fatal(nil, "Aldev config item `.api.bindir` (relative path for the temp folder)  or `.lib.bindir` (if library) is empty!")
	}

	// repeated commands
	execExt := ""
	if utils.IsWindows() {
		execExt = ".exe"
	}
	mainCompileCmd := fmt.Sprintf("go build -o %s/%s-api-local%s ./main", utils.GetBinDir(), utils.Config().AppName, execExt)
	mainRunCmd := fmt.Sprintf("%s/%s-api-local%s -config %s -srcdir %s -bindir %s",
		utils.Config().ResolvedBinDir(), utils.Config().AppName, execExt,
		path.Join(utils.GetGoSrcDir(), utils.GetConfigPath()),
		utils.GetGoSrcDir(), path.Join(utils.GetGoSrcDir(), utils.GetBinDir()))
	if utils.Config().Web != nil {
		mainRunCmd = fmt.Sprintf("%s -webdir %s", mainRunCmd, utils.Config().Web.SrcDir)
	}
	if utils.Config().Native != nil {
		mainRunCmd = fmt.Sprintf("%s -nativedir %s", mainRunCmd, utils.Config().Native.SrcDir)
	}

	// compilation n°1 - this is needed to have the run command up-to-date
	utils.Run("Only compiling & formatting the code", completeCtx, false, "%s", mainCompileCmd)

	if compilationOnly {
		return
	}

	regenArg := ""
	if regen {
		regenArg = " -regen"
	}

	// generation step n°1
	utils.QuickRun("Generating stuff: DB list, BOclasses, BO registry...", "%s", mainRunCmd+" -codegen 1"+regenArg)

	// compilation n°2
	if codeHasChanged() {
		utils.Run("Does it still compile after codegen step 1?", completeCtx, false, "%s", mainCompileCmd)
	}

	// generation step n°2
	utils.QuickRun("Generating stuff: BO specs...", "%s", mainRunCmd+" -codegen 2"+regenArg)

	// compilation n°3
	if codeHasChanged() {
		utils.Run("Does it still compile after codegen step 2?", completeCtx, false, "%s", mainCompileCmd)
	}

	// generation step n°3
	utils.QuickRun("Generating stuff: BO vmaps, BO web models, etc...", "%s", mainRunCmd+" -codegen 3"+regenArg)

	// compilation n°4
	if codeHasChanged() {
		utils.Run("Does it still compile after codegen step 3?", completeCtx, false, "%s", mainCompileCmd)
	}

	// formatting
	utils.QuickRun("Formatting the code", "gofumpt -w %s %s", path.Join(utils.GetGoSrcDir(), "_include"), path.Join(utils.GetGoSrcDir(), "main"))

	// migrating the DBs if needed
	if !utils.IsDevLibrary() {
		utils.QuickRun("DB automigration", "%s", mainRunCmd+" -migrate")
	}

	// under Windows, the executable for codegen and API serving is not the same - we need to build the executable for the containers
	if utils.IsWindows() && !noContainer && codeHasChanged() {
		secondaryCompileCmd := fmt.Sprintf("go build -o %s/%s-api-local ./main", utils.GetBinDir(), utils.Config().AppName)
		utils.Run("Compiling for Docker (Linux)", completeCtx.WithEnvVars("GOOS=linux"), false, "%s", secondaryCompileCmd)
	}

	// bit of logging
	utils.Info("Aldev codegen done in %s", time.Since(start))
}

// ----------------------------------------------------------------------------
// Utils
// ----------------------------------------------------------------------------

const dirtyFILENAME = "dirty"

func codeHasChanged() bool {
	return string(utils.ReadFile(nil, path.Join(utils.GetGoSrcDir(), utils.GetBinDir(), dirtyFILENAME), false)) == "true"
}
