package confgen

import (
	"github.com/aldesgroup/aldev/cmd"
	"github.com/aldesgroup/aldev/utils"
	"github.com/spf13/cobra"
)

// ----------------------------------------------------------------------------
// Command declaration
// ----------------------------------------------------------------------------

// aldevConfgenCmd represents a subcommand
var aldevConfgenCmd = &cobra.Command{
	Use:   "confgen",
	Short: "Generates config files, used notably for local & remote deployment",
	Run:   aldevConfgenRun,
}

var (
// no specific argument for now
)

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevConfgenCmd)
}

// ----------------------------------------------------------------------------
// Main logic
// ----------------------------------------------------------------------------

func aldevConfgenRun(command *cobra.Command, args []string) {
	// Reading this command's arguments, and reading the aldev YAML config file
	cmd.ReadCommonArgsAndConfig()

	// Generating all the deployment files
	utils.GenerateDeployFiles(nil)
}
