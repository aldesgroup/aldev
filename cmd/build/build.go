package build

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

// aldevBuildCmd represents a subcommand
var aldevBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Builds the app with additional generated code to speed up your dev",
	Long: "This generates additional code to provide you with useful handles on DB lists" +
		"business object classes & their properties, and more. And rebuilds the whole app.",
	Run: aldevBuildRun,
}

var (
	compilationOnly bool
	verbose         bool
)

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevBuildCmd)

	aldevBuildCmd.Flags().BoolVarP(&compilationOnly, "compilation-only", "c", false, "does only the compilation of the code, no generation step")
	aldevBuildCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "activates the verbose mode")
}

// ----------------------------------------------------------------------------
// Main logic watching for some files, and (re)starting the building &
// deploying function
// ----------------------------------------------------------------------------

// TODO handle compilation errors = rollbacks on previous situation
func aldevBuildRun(command *cobra.Command, args []string) {
	// it's only here that we have this variable valued
	if verbose {
		utils.SetVerbose()
	}

	start := time.Now()

	// reading the Aldev config one first time
	cfg := utils.ReadConfig(cmd.GetConfigFilename())

	// the context to build Go sources
	buildCtx := utils.InitAldevContext(100, nil).WithExecDir(cfg.API.SrcDir)

	// making sure we're applying what's decided in the go.mod file
	utils.Run("Making sure we're using the right set of dependencies", buildCtx, false, "go mod tidy")

	// control
	if cfg.API.Build.BinDir == "" {
		utils.Fatal("Aldev config item `.api.build.bindir` (relative path for the bin folder from the API directory) is empty!")
	}

	// // some args for the code generation part
	// libraryArg := ""
	// if cfg.API.Build.Library {
	// 	libraryArg = " -library"
	// }

	// repeated commands
	mainBuildCmd := fmt.Sprintf("go build -o %s/%s-api-local ./main", cfg.API.Build.BinDir, cfg.AppName)
	mainRunCmd := fmt.Sprintf("%s/%s-api-local -config %s -srcdir %s", //+libraryArg,
		cfg.API.Build.ResolvedBinDir, cfg.AppName, path.Join(cfg.API.SrcDir, cfg.API.Config), cfg.API.SrcDir)

	// compilation n°1 - this is needed to have the run command up-to-date
	utils.Run("Only compiling & formatting the code", buildCtx, false, mainBuildCmd)

	if compilationOnly {
		return
	}

	// generation step n°1
	utils.QuickRun("Generating stuff: DB list, BO registry...", mainRunCmd+" -codegen 1")

	// compilation n°2
	utils.Run("Does it still compile after codegen step 1?", buildCtx, false, mainBuildCmd)

	// generation step n°2
	utils.QuickRun("Generating stuff: BO classes...", mainRunCmd+" -codegen 2")

	// compilation n°3
	utils.Run("Does it still compile after codegen step 2?", buildCtx, false, mainBuildCmd)

	// Exit:
	// formatting
	utils.QuickRun("Formatting the code", "gofumpt -w %s", path.Join(cfg.API.SrcDir, "_generated"))

	// bit of logging
	utils.Info("Aldev build done in %s", time.Since(start))
}
