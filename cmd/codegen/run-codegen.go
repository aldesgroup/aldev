package codegen

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/aldesgroup/aldev/cmd"
	"github.com/aldesgroup/aldev/utils"
	core "github.com/aldesgroup/corego"
	"github.com/spf13/cobra"
)

// ----------------------------------------------------------------------------
// Command declaration
// ----------------------------------------------------------------------------

// aldevCodegenCmd represents a subcommand
var aldevCodegenCmd = &cobra.Command{
	Use:   "codegen",
	Short: "Completes the app with additional generated code to speed up your dev",
	Long: "This generates additional code to provide you with useful handles on DB lists" +
		"business object classes & their properties, and more. And re-compiles the whole app.",
	Run: aldevCodegenRun,
}

var (
	compilationOnly bool
	noContainer     bool
)

func init() {
	// linking to the root command
	cmd.GetAldevCmd().AddCommand(aldevCodegenCmd)

	aldevCodegenCmd.Flags().BoolVarP(&compilationOnly, "compilation-only", "c", false, "does only the compilation of the code, no generation step")
	aldevCodegenCmd.Flags().BoolVarP(&noContainer, "no-container", "n", false, "if true, then does not build the binary for containerisation")
}

// ----------------------------------------------------------------------------
// Main logic
// ----------------------------------------------------------------------------

// TODO handle compilation errors = rollbacks on previous situation
func aldevCodegenRun(command *cobra.Command, args []string) {
	start := time.Now()

	// Reading this command's arguments, and reading the aldev YAML config file
	cmd.ReadCommonArgsAndConfig()

	// the contexts to build & codegen Go sources
	buildingCtx := utils.InitAldevContext(100, nil).WithExecDir(utils.GetGoSrcDir()).WithAllowFailure(true)
	codegenCtx := utils.InitAldevContext(100, nil).WithAllowFailure(true)

	// making sure we're applying what's decided in the go.mod file
	must(utils.Run("Making sure we're using the right set of dependencies", buildingCtx, true, "go mod tidy"))

	// control
	if utils.GetBinDir() == "" {
		core.PanicMsg("Aldev config item `.api.bindir` (relative path for the temp folder)  or `.lib.bindir` (if library) is empty!")
	}

	// repeated commands
	execExt := ""
	if core.IsWindows() {
		execExt = ".exe"
	}
	binName := core.PascalToKebab(utils.Config().AppName) + "-api"

	mainCompileCmd := fmt.Sprintf("go build -o %s/%s%s ./main", utils.GetBinDir(), binName, execExt)

	mainRunCmd := fmt.Sprintf("%s/%s%s -config %s -srcdir %s -bindir %s",
		utils.Config().ResolvedBinDir(), binName, execExt,
		path.Join(utils.GetGoSrcDir(), "conf-local.yaml"),
		utils.GetGoSrcDir(),
		path.Join(utils.GetGoSrcDir(), utils.GetBinDir()),
	)

	if utils.Config().API != nil {
		mainRunCmd = fmt.Sprintf("%s -docpath %s", mainRunCmd, utils.Config().API.DocPath)
		mainRunCmd = fmt.Sprintf("%s -othersrcdirs %s", mainRunCmd, strings.Join(utils.Config().API.WatchAlso, ","))
	}
	if utils.Config().Web != nil {
		mainRunCmd = fmt.Sprintf("%s -webdir %s", mainRunCmd, utils.Config().Web.SrcDir)
	}
	if utils.Config().Native != nil {
		mainRunCmd = fmt.Sprintf("%s -nativedir %s", mainRunCmd, utils.Config().Native.SrcDir)
	}

	// compilation n°1 - this is needed to have the run command up-to-date
	must(utils.Run("Compiling & formatting the code", buildingCtx, true, "%s", mainCompileCmd))

	if compilationOnly {
		return
	}

	regenArg := ""
	if utils.IsRegen() {
		regenArg = " -regen"
	}

	// generation step n°1
	must(utils.Run("Generating stuff: DB list, BOclasses, BO registry...", codegenCtx, true, "%s", mainRunCmd+" -codegen 1"+regenArg))

	// compilation n°2
	if codeHasChanged() {
		must(utils.Run("Does it still compile after codegen step 1?", buildingCtx, true, "%s", mainCompileCmd))
	}

	// generation step n°2
	must(utils.Run("Generating stuff: BO models...", codegenCtx, true, "%s", mainRunCmd+" -codegen 2"+regenArg))

	// compilation n°3
	if codeHasChanged() {
		must(utils.Run("Does it still compile after codegen step 2?", buildingCtx, true, "%s", mainCompileCmd))
	}

	// generation step n°3
	must(utils.Run("Generating stuff: BO vmaps, BO web models, etc...", codegenCtx, true, "%s", mainRunCmd+" -codegen 3"+regenArg))

	// generation step n°3-bis
	if codeHasChanged() {
		codeComplete()
	}

	// compilation n°4 -
	if codeHasChanged() {
		must(utils.Run("Does it still compile after codegen step 3?", buildingCtx, true, "%s", mainCompileCmd))
	}

	// generation step n°4 = verification
	must(utils.Run("Checking the code...", codegenCtx, true, "%s", mainRunCmd+" -codegen 4"+regenArg))

	// formatting
	must(utils.Run("Formatting the code", codegenCtx, true, "gofumpt -w %s %s", path.Join(utils.GetGoSrcDir(), "_include"), path.Join(utils.GetGoSrcDir(), "main")))

	// migrating the DBs if needed
	if !utils.IsDevLibrary() {
		must(utils.Run("DB automigration", codegenCtx, true, "%s", mainRunCmd+" -migrate"))
	}

	// under Windows, the executable for codegen and API serving is not the same - we need to build the executable for the containers
	if core.IsWindows() && !noContainer && codeHasChanged() {
		secondaryCompileCmd := fmt.Sprintf("go build -o %s/%s ./main", utils.GetBinDir(), binName)
		must(utils.Run("Compiling for containerization (Linux)", buildingCtx.WithEnvVars("GOOS=linux"), true, "%s", secondaryCompileCmd))
	}

	// linting / checking the code quality
	codeLint()

	// bit of logging
	utils.Info("Aldev codegen done in %s", time.Since(start))
}

// ----------------------------------------------------------------------------
// Utils - completing the code
// ----------------------------------------------------------------------------

func codeComplete() {
	// going over all the BO files
	completeRecentCode(time.Now().Add(-5*time.Second), utils.GetGoSrcDir())
}

var skipCodeCompleteForDirs = []string{"_include", "class"}

func completeRecentCode(newerThan time.Time, dir string) {
	if core.InSlice(skipCodeCompleteForDirs, dir) {
		return
	}

	for _, entry := range core.EnsureReadDir(dir) {
		if entry.IsDir() {
			completeRecentCode(newerThan, path.Join(dir, entry.Name()))
		} else {
			if strings.HasSuffix(entry.Name(), "--.go") {
				if filepath := path.Join(dir, entry.Name()); utils.IsRegen() || core.EnsureModTime(filepath).After(newerThan) {
					utils.Info("Completing code for file: %s", filepath)
					utils.QuickRun("Adding missing tags", "gomodifytags -file %s -all -add-tags json,io:in|i*|o*,desc: -transform camelcase -w --quiet -skip-unexported", filepath)
					utils.QuickRun("Aligning the tags", "formattag -file %s", filepath)

				}
			}
		}
	}
}

// ----------------------------------------------------------------------------
// Utils - linting
// ----------------------------------------------------------------------------

func codeLint() {
	// for now, only the API doc
	apiDocLint()
}

func apiDocLint() {
	// there's no API doc to lint if there's no API
	if utils.Config().API != nil {
		docPath := withDefault(utils.Config().API.DocPath, "data/api-doc.yaml")
		docReport := withDefault(utils.Config().API.DocReport, "tmp/api-doc-report.yaml")
		browser := withDefault(utils.Config().API.Browser, "chromium")

		// we can't generate the report if the API doc does not exist
		if core.FileExists(docPath) {
			// we need to generate it if it does not exist yet, or is older than the API doc
			if !core.FileExists(docReport) || core.EnsureModTime(docReport).Before(core.EnsureModTime(docPath)) {

				// making sure the folder path for the report exists
				core.EnsureDir(path.Dir(docReport))

				// Generating the report
				cmd := exec.Command("vacuum", "html-report", docPath, docReport)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				core.PanicIfErr(cmd.Run())

				slog.Info("API doc report generated: " + docReport)

				// Opening it automatically
				core.PanicIfErr(exec.Command(browser, docReport).Start())
			}
		}
	}
}

// ----------------------------------------------------------------------------
// Utils - misc
// ----------------------------------------------------------------------------

const dirtyFILENAME = "dirty"

func codeHasChanged() bool {
	return string(core.ReadFile(path.Join(utils.GetGoSrcDir(), utils.GetBinDir(), dirtyFILENAME), false)) == "true"
}

func must(result bool) {
	if !result {
		panic("Issue with code compilation or generation!")
	}
}

func withDefault(value, defaultValue string) string {
	return core.IfThenElse(value != "", value, defaultValue)
}
