package release

// Here, I want to be able to run: aldev release patch, or aldev release minor, or aldev release major
// Using cobra.Command

import (
	"github.com/aldesgroup/aldev/cmd"
	"github.com/aldesgroup/aldev/utils"
	"github.com/spf13/cobra"
)

// ----------------------------------------------------------------------------
// Command declaration
// ----------------------------------------------------------------------------

var aldevReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Make a new Git tag, and push the code to the remote `releases` branch.",
	Long:  "Use with a subcommand: patch (vM.N.P -> vM.N.P+1), minor (vM.N.P -> vM.N+1.0) or major (vM.N.P -> vM+1.0.0)",
}

var releaseMajorCmd = &cobra.Command{
	Use:   "major",
	Short: "Release a major version",
	Run: func(cmd *cobra.Command, args []string) {
		handleReleaseRun(utils.ReleaseMajor)
	},
}

var releaseMinorCmd = &cobra.Command{
	Use:   "minor",
	Short: "Release a minor version",
	Run: func(cmd *cobra.Command, args []string) {
		handleReleaseRun(utils.ReleaseMinor)
	},
}

var releasePatchCmd = &cobra.Command{
	Use:   "patch",
	Short: "Release a patch version",
	Run: func(cmd *cobra.Command, args []string) {
		handleReleaseRun(utils.ReleasePatch)
	},
}

func init() {
	aldevReleaseCmd.AddCommand(releaseMajorCmd)
	aldevReleaseCmd.AddCommand(releaseMinorCmd)
	aldevReleaseCmd.AddCommand(releasePatchCmd)
	cmd.GetAldevCmd().AddCommand(aldevReleaseCmd)
}

// ----------------------------------------------------------------------------
// Main logic
// ----------------------------------------------------------------------------

func handleReleaseRun(release utils.ReleaseType) {
	// Reading this command's arguments, and reading the aldev YAML config file
	cmd.ReadCommonArgsAndConfig()

	// the main cancelable context, that should stop everything
	aldevCtx := utils.InitAldevContext(100, nil)

	// creating the release in parallel
	utils.MakeRelease(aldevCtx, release)
}
