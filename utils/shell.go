// ----------------------------------------------------------------------------
// The code here is about executing shell commands
// ----------------------------------------------------------------------------
package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"
)

func RunWithCtx(ctx context.Context, longCommand string, params ...any) {
	commandElements := strings.Split(fmt.Sprintf(longCommand, params...), " ")
	runCmd(exec.CommandContext(ctx, commandElements[0], commandElements[1:]...), true, nil, Fatal)
}

func Run(longCommand string, params ...any) {
	commandElements := strings.Split(fmt.Sprintf(longCommand, params...), " ")
	runCmd(exec.Command(commandElements[0], commandElements[1:]...), false, nil, Fatal)
}

func RunAndGet(errLogFn logFn, longCommand string, params ...any) []byte {
	commandElements := strings.Split(fmt.Sprintf(longCommand, params...), " ")
	buffer := new(bytes.Buffer)
	runCmd(exec.Command(commandElements[0], commandElements[1:]...), false, buffer, errLogFn)
	return buffer.Bytes()
}

func runCmd(cmd *exec.Cmd, long bool, stdOutCapture io.Writer, errLogFn logFn) io.Writer {
	// making sure we're showing everything the command will throw
	cmd.Stderr = os.Stderr
	if stdOutCapture != nil {
		cmd.Stdout = stdOutCapture
	} else {
		cmd.Stdout = os.Stdout
	}

	// bit of logging
	if long {
		Step("--- [SH.RUN]> Starting: '%s'", cmd.String())
	}
	start := time.Now()

	// actually running the command
	if errRun := cmd.Run(); errRun != nil {
		exitErr, ok := errRun.(*exec.ExitError)
		if long && ok && exitErr.ExitCode() == -1 {
			fmt.Println("Command canceled due to context cancellation")
		} else {
			errLogFn("Command failed: %v", errRun.Error())
		}
	}

	// bit of logging
	prefix := "Done"
	if long {
		prefix = "Finished"
	}
	Step("--- [SH.RUN]> %s: '%s' in %s", prefix, cmd.String(), time.Since(start))

	return stdOutCapture
}
