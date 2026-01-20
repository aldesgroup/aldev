package refresh

import (
	"github.com/aldesgroup/aldev/cmd"
	"github.com/aldesgroup/aldev/utils"
	"github.com/spf13/cobra"
)

// ----------------------------------------------------------------------------
// Command declaration
// ----------------------------------------------------------------------------

// aldevDoctorCmd represents a subcommand
var aldevDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Checks the project's health, regarding the dependencies (for now)",
	Long: "Checks, for each existing part of the project (api, web, native), " +
		"if the dependencies are up to date, and proposes to fix it all",
	Run: aldevDoctorRun,
}

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevDoctorCmd)
	// aldevDoctorCmd.Flags().BoolVarP(&withTranslations, "translations", "t", false,
	// 	"if true, then the translations are also refreshed, which they are not by default")
}

// ----------------------------------------------------------------------------
// Main logic
// ----------------------------------------------------------------------------

func aldevDoctorRun(command *cobra.Command, args []string) {
	// Reading this command's arguments, and reading the aldev YAML config file
	cmd.ReadCommonArgsAndConfig()

	// the main cancelable context, that should stop everything
	aldevCtx := utils.InitAldevContext(100, nil)

	// downloading various external resource in parallel
	utils.CheckDeps(aldevCtx, true)
}
