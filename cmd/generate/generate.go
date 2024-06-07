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
	Short: "Downloads the Aldev environment's required external resources",
	Long: "This downloads the required external resources in their latest version, " +
		"like the i18n file with up-to-date translations, or some required dependencies.",
	Run: aldevGenerateRun,
}

var (
	// cfgFileName  string
	verbose bool

// useLocalDeps bool
// options      string
)

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevGenerateCmd)
	aldevGenerateCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "activates the verbose mode")
}

// ----------------------------------------------------------------------------
// Main logic
// ----------------------------------------------------------------------------

func aldevGenerateRun(command *cobra.Command, args []string) {
	// it's only here that we have this variable valued
	if verbose {
		utils.SetVerbose()
	}

	// reading the Aldev config one first time
	cfg := utils.ReadConfig(cmd.GetConfigFilename())

	// downloading various external resource in parallel
	utils.GenerateConfigs(cfg)
}
