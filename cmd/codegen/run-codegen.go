package complete

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
)

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevCodegenCmd)

	aldevCodegenCmd.Flags().BoolVarP(&compilationOnly, "compilation-only", "c", false, "does only the compilation of the code, no generation step")
	aldevCodegenCmd.Flags().BoolVarP(&regen, "regen", "r", false, "forces the regeneration")
}

// ----------------------------------------------------------------------------
// Main logic
// ----------------------------------------------------------------------------

// TODO handle compilation errors = rollbacks on previous situation
func aldevCodegenRun(command *cobra.Command, args []string) {
	start := time.Now()

	// Reading this command's arguments, and reading the aldev YAML config file
	cmd.ReadCommonArgsAndConfig()

	// Controlling we're not using this in native dev
	if utils.IsDevNative() {
		utils.Fatal(nil, "This is not available for native development")
	}

	// the context to build Go sources
	completeCtx := utils.InitAldevContext(100, nil).WithExecDir(utils.GetSrcDir())

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
	mainRunCmd := fmt.Sprintf("%s/%s-api-local%s -config %s -srcdir %s", utils.Config().ResolvedBinDir(), utils.Config().AppName, execExt,
		path.Join(utils.GetSrcDir(), utils.GetConfigPath()), utils.GetSrcDir())
	if utils.Config().Web != nil {
		mainRunCmd = fmt.Sprintf("%s -webdir %s", mainRunCmd, utils.Config().Web.SrcDir)
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
	utils.QuickRun("Generating stuff: DB list, BO registry...", "%s", mainRunCmd+" -codegen 1"+regenArg)

	// compilation n°2
	utils.Run("Does it still compile after codegen step 1?", completeCtx, false, "%s", mainCompileCmd)

	// generation step n°2
	utils.QuickRun("Generating stuff: BO classes...", "%s", mainRunCmd+" -codegen 2"+regenArg)

	// compilation n°3
	utils.Run("Does it still compile after codegen step 2?", completeCtx, false, "%s", mainCompileCmd)

	// generation step n°3
	utils.QuickRun("Generating stuff: BO vmaps, BO web models, etc...", "%s", mainRunCmd+" -codegen 3"+regenArg)

	// compilation n°4
	utils.Run("Does it still compile after codegen step 3?", completeCtx, false, "%s", mainCompileCmd)

	// formatting
	utils.QuickRun("Formatting the code", "gofumpt -w %s %s", path.Join(utils.GetSrcDir(), "_include"), path.Join(utils.GetSrcDir(), "main"))

	// migrating the DBs if needed
	if !utils.IsDevLibrary() {
		utils.QuickRun("DB automigration", "%s", mainRunCmd+" -migrate")
	}

	// under Windows, the executable for codegen and API serving is not the same - we need to build the executable for the containers
	if utils.IsWindows() {
		secondaryCompileCmd := fmt.Sprintf("go build -o %s/%s-api-local ./main", utils.GetBinDir(), utils.Config().AppName)
		utils.Run("Compiling for Docker (Linux)", completeCtx.WithEnvVars("GOOS=linux"), false, "%s", secondaryCompileCmd)
	}

	// bit of logging
	utils.Info("Aldev complete done in %s", time.Since(start))
}
