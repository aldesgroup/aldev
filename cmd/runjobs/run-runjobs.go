package refresh

import (
	"github.com/aldesgroup/aldev/cmd"
	"github.com/aldesgroup/aldev/utils"
	"github.com/spf13/cobra"
)

// ----------------------------------------------------------------------------
// Command declaration
// ----------------------------------------------------------------------------

// aldevRunJobsCmd represents a subcommand
var aldevRunJobsCmd = &cobra.Command{
	Use:   "runjobs",
	Short: "Run the jobs defined in the aldev.yaml file",
	Long: "This runs the jobs defined in the aldev.yaml file. The jobs either run in parallel, " +
		"or their order is determined by the order of the jobs in the aldev.yaml file.",
	Run: aldevRunJobsRun,
}

var (
	parallel bool
)

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevRunJobsCmd)
	aldevRunJobsCmd.Flags().BoolVarP(&parallel, "parallel", "p", false, "run the jobs in parallel")
}

// ----------------------------------------------------------------------------
// Main logic
// ----------------------------------------------------------------------------

func aldevRunJobsRun(command *cobra.Command, args []string) {
	// Reading this command's arguments, and reading the aldev YAML config file
	cmd.ReadCommonArgsAndConfig()

	// the main cancelable context, that should stop everything
	aldevCtx := utils.InitAldevContext(100, nil)

	// downloading various external resource in parallel
	utils.RunJobs(aldevCtx, parallel)
}
