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

func Run(whyRunThis string, ctx CancelableContext, logStart bool, commandAsString string, params ...any) bool {
	// splitting the command elements as expected by the os/exec package
	commandElements := strings.Split(fmt.Sprintf(commandAsString, params...), " ")

	// running the command
	return runCmd(whyRunThis, ctx, logStart, exec.CommandContext(ctx, commandElements[0], commandElements[1:]...))
}

func QuickRun(whyRunThis string, commandAsString string, params ...any) bool {
	if verbose {
		return Run(whyRunThis, NewBaseContext().WithStdErrWriter(os.Stdout).WithStdOutWriter(os.Stdout), verbose, commandAsString, params...)
	}

	return Run(whyRunThis, NewBaseContext().WithStdErrWriter(io.Discard), false, commandAsString, params...)
}

func RunAndGet(whyRunThis string, execDir string, logStart bool, commandAsString string, params ...any) []byte {
	commandElements := strings.Split(fmt.Sprintf(commandAsString, params...), " ")
	buffer := new(bytes.Buffer)
	ctx := NewBaseContext().WithStdOutWriter(buffer).WithExecDir(execDir).WithAllowFailure(true)
	if !verbose {
		ctx.WithStdErrWriter(io.Discard)
	}
	runCmd(whyRunThis, ctx, logStart, exec.Command(commandElements[0], commandElements[1:]...))
	return buffer.Bytes()
}

func runCmd(whyRunThis string, ctxArg CancelableContext, logStart bool, cmd *exec.Cmd) bool {
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

	// passing the env vars, if any
	for _, envVar := range ctx.getEnvVars() {
		cmd.Env = append(os.Environ(), envVar)
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
			errMsg := fmt.Sprintf("Command [%s%s] failed: %v", fromDirString, cmd.String(), errRun.Error())
			// let's re-run to have more info, if not printed on stderr at first
			if ctx.isReRun() {
				Error("%s", errMsg)
				Run("Re-running the command to get the error logs",
					NewBaseContext().WithStdErrWriter(os.Stderr).WithExecDir(ctx.getExecDir()), true, "%s", cmd.String())
			} else {
				if !ctx.isAllowingFailure() {
					panic(errMsg)
				} else {
					Error("%s", errMsg)
				}
			}
		}

		return false
	}

	// bit of logging, only in verbose mode
	if verbose {
		if logStart {
			Step("--- [SH.RUN]> Finished%s: '%s' in %s", fromDirString, cmd.String(), time.Since(start))
		} else {
			StepWithPreamble(whyRunThis, "--- [SH.RUN]> Done%s: '%s' in %s", fromDirString, cmd.String(), time.Since(start))
		}
	}

	// it went fine
	return true
}
