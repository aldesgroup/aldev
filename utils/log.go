// ----------------------------------------------------------------------------
// The code here is about logging in the shell
// ----------------------------------------------------------------------------
package utils

import (
	"fmt"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

var (
	verbose bool
)

func SetVerbose(isVerbose bool) {
	verbose = isVerbose
	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
}

func Debug(str string, params ...any) {
	if verbose {
		slog.Debug(fmt.Sprintf(str, params...))
	}
}

func Info(str string, params ...any) {
	slog.Info(fmt.Sprintf(str, params...))
}

type logFn func(string, ...any)

type errLogFn func(CancelableContext, string, ...any)

func log(preambleMsg string, fn logFn, separator, str string, params ...any) {
	println("")
	msg := fmt.Sprintf(str, params...)
	sep := strings.Repeat(separator, max(len(preambleMsg), len(msg)))
	fn(sep)
	if preambleMsg != "" {
		fn(preambleMsg)
	}
	fn(msg)
	fn(sep)
}

func Step(str string, params ...any) {
	log("", slog.Info, "-", str, params...)
}

func StepWithPreamble(preamble, str string, params ...any) {
	log(preamble, slog.Info, "-", str, params...)
}

func Warn(str string, params ...any) {
	log("", slog.Warn, "=", str, params...)
}

func Error(str string, params ...any) {
	log("", slog.Error, "*", str, params...)
}

func Fatal(ctx CancelableContext, str string, params ...any) {
	Error(str, params...)
	Info("Stack: %s", string(debug.Stack()))
	if ctx != nil {
		ctx.CancelAll()
	}
	Debug("Waiting a bit for other processes to finish")
	time.Sleep(2 * time.Second)
	os.Exit(1)
}

func FatalErr(ctx CancelableContext, err error) {
	Fatal(ctx, "An error has occurred: %s", err)
}

func FatalIfErr(ctx CancelableContext, err error) {
	if err != nil {
		FatalErr(ctx, err)
	}
}
