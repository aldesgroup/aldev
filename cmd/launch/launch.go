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
	aldevLaunchCmd.Flags().BoolVarP(&useLocalDeps, "use-local-deps", "u", false,
		"to use the local versions of the dependencies declared in the config file")
}

// ----------------------------------------------------------------------------
// Main logic
// ----------------------------------------------------------------------------

func aldevLaunchRun(command *cobra.Command, args []string) {
	// Reading this command's arguments, and reading the aldev YAML config file
	cmd.ReadCommonArgsAndConfig()

	if useLocalDeps {
		utils.SetUseLocalDeps()
	}

	// the main cancelable context, that should stop everything
	aldevCtx := utils.InitAldevContext(100, nil)

	// launching!
	utils.DeployToLocalCluster(aldevCtx)
}
