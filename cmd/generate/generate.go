package generate

import (
	"github.com/aldesgroup/aldev/cmd"
	"github.com/aldesgroup/aldev/utils"
	"github.com/spf13/cobra"
)

// ----------------------------------------------------------------------------
// Command declaration
// ----------------------------------------------------------------------------

// aldevGenerateCmd represents a subcommand
var aldevGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates config files, used notably for local & remote deployment",
	Long:  "Generates the Tiltfile, Docker & Kustomization files, and an .env-list for web app development",
	Run:   aldevGenerateRun,
}

var (
// no specific argument for now
)

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevGenerateCmd)
}

// ----------------------------------------------------------------------------
// Main logic
// ----------------------------------------------------------------------------

func aldevGenerateRun(command *cobra.Command, args []string) {
	// Reading this command's arguments, and reading the aldev YAML config file
	cmd.ReadCommonArgsAndConfig()

	// downloading various external resource in parallel
	utils.GenerateDeployConfigs(nil)
}
