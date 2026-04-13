package utils

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"
)

const waitBeforeExit = 2 * time.Second

func Recover(ctx CancelableContext, msg string, params ...any) {
	if err := recover(); err != nil {
		Error("%v", err)
		Info("Recovered from error (%v) while %s; will cancel the whole process now; Stack: %s", err, fmt.Sprintf(msg, params...), string(debug.Stack()))
		time.Sleep(waitBeforeExit)
		Debug("Waiting %s for other processes to finish, then EXITING", waitBeforeExit)
		ctx.CancelAll()
		time.Sleep(waitBeforeExit)
		os.Exit(1)
	}
}
