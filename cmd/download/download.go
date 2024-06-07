package update

import (
	"github.com/aldesgroup/aldev/cmd"
	"github.com/aldesgroup/aldev/utils"
	"github.com/spf13/cobra"
)

// ----------------------------------------------------------------------------
// Command declaration
// ----------------------------------------------------------------------------

// aldevDownloadCmd represents a subcommand
var aldevDownloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Downloads the Aldev environment's required external resources",
	Long: "This downloads the required external resources in their latest version, " +
		"like the i18n file with up-to-date translations, or some required dependencies.",
	Run: aldevDownloadRun,
}

var (
	// cfgFileName  string
	verbose bool

// useLocalDeps bool
// options      string
)

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevDownloadCmd)
	aldevDownloadCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "activates the verbose mode")
}

// ----------------------------------------------------------------------------
// Main logic watching for some files, and (re)starting the building &
// deploying function
// ----------------------------------------------------------------------------

func aldevDownloadRun(command *cobra.Command, args []string) {
	// it's only here that we have this variable valued
	if verbose {
		utils.SetVerbose()
	}

	// reading the Aldev config one first time
	cfg := utils.ReadConfig(cmd.GetConfigFilename())

	// the main cancelable context, that should stop everything
	aldevCtx := utils.InitAldevContext(100, nil)

	// TODO add installation / lib update step
	// TODO add -deps to include go get -u -v ./main && git commit && git push

	// downloading various external resource in parallel
	utils.DownloadExternalResources(aldevCtx, cfg)
}
