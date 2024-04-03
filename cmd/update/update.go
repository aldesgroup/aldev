package update

import (
	"github.com/aldesgroup/aldev/cmd"
	"github.com/aldesgroup/aldev/utils"
	"github.com/spf13/cobra"
)

// ----------------------------------------------------------------------------
// Command declaration
// ----------------------------------------------------------------------------

// aldevUpdateCmd represents the base command when called without any subcommands
var aldevUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates the Aldev environment's required external resources",
	Long: "This downloads the required external resources in their latest version, " +
		"like the i18n file with up-to-date translations. Later, this subcommand " +
		"will also help set up and maintain the whole dev environment, " +
		"by installing / updating all the locally needed software.",
	Run: aldevUpdateRun,
}

var (
// cfgFileName  string
// verbose      bool
// useLocalDeps bool
// options      string
)

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevUpdateCmd)
}

// ----------------------------------------------------------------------------
// Main logic watching for some files, and (re)starting the building &
// deploying function
// ----------------------------------------------------------------------------

func aldevUpdateRun(command *cobra.Command, args []string) {
	// reading the Aldev config one first time
	cfg := utils.ReadConfig(cmd.GetConfigFilename())

	// the main cancelable context, that should stop everything
	aldevCtx := utils.InitAldevContext()

	// TODO add installation / lib update step
	// TODO add -deps to include go get -u -v ./main && git commit && git push

	// proceed to download external resources
	utils.DownloadExternalResources(aldevCtx, cfg)
}
