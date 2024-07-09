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

// aldevCompleteCmd represents a subcommand
var aldevCompleteCmd = &cobra.Command{
	Use:   "complete",
	Short: "Completes the app with additional generated code to speed up your dev",
	Long: "This generates additional code to provide you with useful handles on DB lists" +
		"business object classes & their properties, and more. And re-compiles the whole app.",
	Run: aldevCompleteRun,
}

var (
	compilationOnly bool
	verbose         bool
)

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevCompleteCmd)

	aldevCompleteCmd.Flags().BoolVarP(&compilationOnly, "compilation-only", "c", false, "does only the compilation of the code, no generation step")
	aldevCompleteCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "activates the verbose mode")
}

// ----------------------------------------------------------------------------
// Main logic
// ----------------------------------------------------------------------------

// TODO handle compilation errors = rollbacks on previous situation
func aldevCompleteRun(command *cobra.Command, args []string) {
	// it's only here that we have this variable valued
	if verbose {
		utils.SetVerbose()
	}

	start := time.Now()

	// reading the Aldev config one first time
	cfg := utils.ReadConfig(cmd.GetConfigFilename())

	// the context to build Go sources
	completeCtx := utils.InitAldevContext(100, nil).WithExecDir(cfg.GetSrcDir())

	// making sure we're applying what's decided in the go.mod file
	utils.Run("Making sure we're using the right set of dependencies", completeCtx, false, "go mod tidy")

	// control
	if cfg.GetBinDir() == "" {
		utils.Fatal(nil, "Aldev config item `.api.bindir` (relative path for the temp folder)  or `.lib.bindir` (if library) is empty!")
	}

	// repeated commands
	mainCompileCmd := fmt.Sprintf("go build -o %s/%s-api-local ./main", cfg.GetBinDir(), cfg.AppName)
	mainRunCmd := fmt.Sprintf("%s/%s-api-local -config %s -srcdir %s", cfg.GetResolvedBinDir(), cfg.AppName,
		path.Join(cfg.GetSrcDir(), cfg.GetConfigPath()), cfg.GetSrcDir())

	// compilation n°1 - this is needed to have the run command up-to-date
	utils.Run("Only compiling & formatting the code", completeCtx, false, mainCompileCmd)

	if compilationOnly {
		return
	}

	// generation step n°1
	utils.QuickRun("Generating stuff: DB list, BO registry...", mainRunCmd+" -codegen 1")

	// compilation n°2
	utils.Run("Does it still compile after codegen step 1?", completeCtx, false, mainCompileCmd)

	// generation step n°2
	utils.QuickRun("Generating stuff: BO classes...", mainRunCmd+" -codegen 2")

	// compilation n°3
	utils.Run("Does it still compile after codegen step 2?", completeCtx, false, mainCompileCmd)

	// generation step n°3
	utils.QuickRun("Generating stuff: BO utils...", mainRunCmd+" -codegen 3")

	// compilation n°4
	utils.Run("Does it still compile after codegen step 3?", completeCtx, false, mainCompileCmd)

	// formatting
	utils.QuickRun("Formatting the code", "gofumpt -w %s %s", path.Join(cfg.GetSrcDir(), "_generated"), path.Join(cfg.GetSrcDir(), "main"))

	// migrating the DBs if needed
	if !cfg.IsLibrary() {
		utils.QuickRun("DB automigration", mainRunCmd+" -migrate")
	}

	// bit of logging
	utils.Info("Aldev complete done in %s", time.Since(start))
}
