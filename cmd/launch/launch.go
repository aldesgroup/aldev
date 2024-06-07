package deploy

import (
	"github.com/aldesgroup/aldev/cmd"
	"github.com/aldesgroup/aldev/utils"
	"github.com/spf13/cobra"
)

// ----------------------------------------------------------------------------
// Command declaration
// ----------------------------------------------------------------------------

// aldevLaunchCmd represents a subcommand
var aldevLaunchCmd = &cobra.Command{
	Use:   "launch",
	Short: "Locally deploys the app, i.e. its API and / or its client (web) app",
	Long:  "This simply launch the app(s), using Tilt, and the pre-existing config files (aldev generate)",
	Run:   aldevLaunchRun,
}

var (
	// cfgFileName  string
	verbose      bool
	useLocalDeps bool
)

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevLaunchCmd)
	aldevLaunchCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "activates the verbose mode")
	aldevLaunchCmd.Flags().BoolVarP(&useLocalDeps, "use-local-deps", "l", false,
		"to use the local dependencies declared in the config file")
}

// ----------------------------------------------------------------------------
// Main logic
// ----------------------------------------------------------------------------

func aldevLaunchRun(command *cobra.Command, args []string) {
	// it's only here that we have this variable valued
	if verbose {
		utils.SetVerbose()
	}

	if useLocalDeps {
		utils.SetUseLocalDeps()
	}

	// reading the Aldev config one first time
	cfg := utils.ReadConfig(cmd.GetConfigFilename())

	// the main cancelable context, that should stop everything
	aldevCtx := utils.InitAldevContext(100, nil)

	// launching!
	utils.Launch(aldevCtx, cfg)
}
