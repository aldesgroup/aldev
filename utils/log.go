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
)

var (
	verbose bool
)

func SetVerbose() {
	verbose = true
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func Noop(str string, params ...any) {
	// does nothing
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

func Fatal(str string, params ...any) {
	Error(str, params...)
	Info(string(debug.Stack()))
	os.Exit(1)
}

func FatalErr(err error) {
	Fatal("An error has occurred: %s", err)
}

func FatalIfErr(err error) {
	if err != nil {
		FatalErr(err)
	}
}
