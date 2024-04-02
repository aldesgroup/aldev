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

func RunWithCtx(whyRunThis string, ctx CancelableContext, longCommand string, params ...any) {
	commandElements := strings.Split(fmt.Sprintf(longCommand, params...), " ")
	runCmd(whyRunThis, exec.CommandContext(ctx, commandElements[0], commandElements[1:]...), true, nil, Fatal)
}

func Run(whyRunThis string, longCommand string, params ...any) {
	commandElements := strings.Split(fmt.Sprintf(longCommand, params...), " ")
	runCmd(whyRunThis, exec.Command(commandElements[0], commandElements[1:]...), false, nil, Fatal)
}

func RunAndGet(whyRunThis string, errLogFn logFn, longCommand string, params ...any) []byte {
	commandElements := strings.Split(fmt.Sprintf(longCommand, params...), " ")
	buffer := new(bytes.Buffer)
	runCmd(whyRunThis, exec.Command(commandElements[0], commandElements[1:]...), false, buffer, errLogFn)
	return buffer.Bytes()
}

func runCmd(whyRunThis string, cmd *exec.Cmd, long bool, stdOutCapture io.Writer, errLogFn logFn) io.Writer {
	// making sure we're showing everything the command will throw
	cmd.Stderr = os.Stderr
	if stdOutCapture != nil {
		cmd.Stdout = stdOutCapture
	} else {
		cmd.Stdout = os.Stdout
	}

	// bit of logging
	if long {
		StepWithPreamble(whyRunThis, "--- [SH.RUN]> Starting: '%s'", cmd.String())
	}
	start := time.Now()

	// actually running the command
	if errRun := cmd.Run(); errRun != nil {
		exitErr, ok := errRun.(*exec.ExitError)
		if long && ok && exitErr.ExitCode() == -1 {
			fmt.Println("Command canceled due to context cancellation")
		} else {
			errLogFn("Command [%s] failed: %v", cmd.String(), errRun.Error())
		}
	}

	// bit of logging
	if long {
		Step("--- [SH.RUN]> Finished: '%s' in %s", cmd.String(), time.Since(start))
	} else {
		StepWithPreamble(whyRunThis, "--- [SH.RUN]> Done: '%s' in %s", cmd.String(), time.Since(start))
	}

	return stdOutCapture
}
