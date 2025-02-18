package deploylocal

import (
	"github.com/aldesgroup/aldev/cmd"
	"github.com/aldesgroup/aldev/utils"
	"github.com/spf13/cobra"
)

// ----------------------------------------------------------------------------
// Command declaration
// ----------------------------------------------------------------------------

// aldevDeployLocalCmd represents a subcommand
var aldevDeployLocalCmd = &cobra.Command{
	Use:   "deploylocal",
	Short: "Locally deploys the app, i.e. its API and / or its client (web) app",
	Long:  "This simply launch the app(s), using Tilt, and the pre-existing config files (aldev confgen)",
	Run:   aldevDeployLocalRun,
}

var (
	// cfgFileName  string
	verbose  bool
	swapCode bool
)

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevDeployLocalCmd)
	aldevDeployLocalCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "activates the verbose mode")
	aldevDeployLocalCmd.Flags().BoolVarP(&swapCode, "swap", "s", false,
		"use swapping of code, to use the local version of some dependencies for instance")
}

// ----------------------------------------------------------------------------
// Main logic
// ----------------------------------------------------------------------------

func aldevDeployLocalRun(command *cobra.Command, args []string) {
	// Reading this command's arguments, and reading the aldev YAML config file
	cmd.ReadCommonArgsAndConfig()

	if swapCode {
		utils.UseCodeSwaps()
	}

	// the main cancelable context, that should stop everything
	aldevCtx := utils.InitAldevContext(100, nil)

	// launching!
	utils.DeployToLocalCluster(aldevCtx)
}
