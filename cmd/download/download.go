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
// no specific argument for now
)

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevDownloadCmd)
}

// ----------------------------------------------------------------------------
// Main logic
// ----------------------------------------------------------------------------

func aldevDownloadRun(command *cobra.Command, args []string) {
	// Reading this command's arguments, and reading the aldev YAML config file
	cmd.ReadCommonArgsAndConfig()

	// the main cancelable context, that should stop everything
	aldevCtx := utils.InitAldevContext(100, nil)

	// TODO add installation / lib update step -> maybe rather in a `aldev install` subcommand
	// TODO add -deps to include go get -u -v ./main && git commit && git push

	// downloading various external resource in parallel
	utils.DownloadExternalResources(aldevCtx)
}
