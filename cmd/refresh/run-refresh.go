package refresh

import (
	"github.com/aldesgroup/aldev/cmd"
	"github.com/aldesgroup/aldev/utils"
	"github.com/spf13/cobra"
)

// ----------------------------------------------------------------------------
// Command declaration
// ----------------------------------------------------------------------------

// aldevRefreshCmd represents a subcommand
var aldevRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refreshes the Aldev environment's required external resources",
	Long: "This downloads the required external resources in their latest version, " +
		"like the i18n files with up-to-date translations, or some required dependencies.",
	Run: aldevRefreshRun,
}

var (
	withTranslations bool
)

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevRefreshCmd)
	aldevRefreshCmd.Flags().BoolVarP(&withTranslations, "translations", "t", false,
		"if true, then the translations are also refreshed, which they are not by default")
}

// ----------------------------------------------------------------------------
// Main logic
// ----------------------------------------------------------------------------

func aldevRefreshRun(command *cobra.Command, args []string) {
	// Reading this command's arguments, and reading the aldev YAML config file
	cmd.ReadCommonArgsAndConfig()

	// the main cancelable context, that should stop everything
	aldevCtx := utils.InitAldevContext(100, nil)

	// TODO add installation / lib update step -> maybe rather in a `aldev install` subcommand
	// TODO add -deps to include go get -u -v ./main && git commit && git push

	// downloading various external resource in parallel
	utils.DownloadExternalResources(aldevCtx, withTranslations)

	// creates the required symlinks
	// TODO remove if not used - but could probably be used for linting
	// utils.CreateMissingSymlinks(aldevCtx)
}
