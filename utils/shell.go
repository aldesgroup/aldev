// ----------------------------------------------------------------------------
// The code here is about executing shell commands
// ----------------------------------------------------------------------------
package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

func Run(whyRunThis string, ctx CancelableContext, logStart bool, commandAsString string, params ...any) {
	// splitting the command elements as expected by the os/exec package
	commandElements := strings.Split(fmt.Sprintf(commandAsString, params...), " ")

	// running the command
	runCmd(whyRunThis, ctx, logStart, exec.CommandContext(ctx, commandElements[0], commandElements[1:]...))
}

func QuickRun(whyRunThis string, commandAsString string, params ...any) {
	Run(whyRunThis, NewBaseContext().WithStdErrWriter(io.Discard), false, commandAsString, params...)
}

func RunAndGet(whyRunThis string, execDir string, logStart bool, commandAsString string, params ...any) []byte {
	commandElements := strings.Split(fmt.Sprintf(commandAsString, params...), " ")
	buffer := new(bytes.Buffer)
	runCmd(whyRunThis, NewBaseContext().WithStdOutWriter(buffer).WithExecDir(execDir),
		logStart, exec.Command(commandElements[0], commandElements[1:]...))
	return buffer.Bytes()
}

func runCmd(whyRunThis string, ctxArg CancelableContext, logStart bool, cmd *exec.Cmd) {
	// making sure we have a non-nil context here
	ctx := ctxArg
	if ctx == nil {
		ctx = NewBaseContext()
	}

	// making sure we're showing everything the command will throw
	if ctx.getStdOutWriter() != nil {
		cmd.Stdout = ctx.getStdOutWriter()
	} else {
		cmd.Stdout = os.Stdout
	}

	if ctx.getStdErrWriter() != nil {
		cmd.Stderr = ctx.getStdErrWriter()
	} else {
		cmd.Stderr = os.Stderr
	}

	// changing the execution directory if needed
	fromDirString := ""
	if ctx.getExecDir() != "" {
		cmd.Dir = ctx.getExecDir()
		fromDirString = " [from " + cmd.Dir + "]"
	}

	// bit of logging
	if logStart {
		// but only in verbose mode
		if verbose {
			StepWithPreamble(whyRunThis, "--- [SH.RUN]> Starting%s: '%s'", fromDirString, cmd.String())
		}
	}

	start := time.Now()

	// actually running the command
	if errRun := cmd.Run(); errRun != nil {
		exitErr, ok := errRun.(*exec.ExitError)
		if logStart && ok && exitErr.ExitCode() == -1 {
			Info("Command canceled due to context cancellation")
		} else {
			// let's re-run to have more info, if not printed on stderr at first
			if ctx.getStdErrWriter() != os.Stderr {
				Error("Command [%s] failed: %v", cmd.String(), errRun.Error())
				Run("Re-running the command to get the error logs",
					NewBaseContext().WithStdErrWriter(os.Stderr).WithExecDir(ctx.getExecDir()), true, cmd.String())
			} else {
				ctx.getErrLogFn()(ctx, "Command [%s] failed: %v", cmd.String(), errRun.Error())
			}

		}
	}

	// bit of logging, only in verbose mode
	if verbose {
		if logStart {
			Step("--- [SH.RUN]> Finished%s: '%s' in %s", fromDirString, cmd.String(), time.Since(start))
		} else {
			StepWithPreamble(whyRunThis, "--- [SH.RUN]> Done%s: '%s' in %s", fromDirString, cmd.String(), time.Since(start))
		}
	}
}
