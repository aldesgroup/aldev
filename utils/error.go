package utils

import (
	"fmt"
	"runtime/debug"
)

func Recover(ctx CancelableContext, msg string, params ...any) {
	if err := recover(); err != nil {
		Error("Recovered from error (%v) while %s; will cancel the whole process now; Stack: %s", err, fmt.Sprintf(msg, params...), string(debug.Stack()))
		ctx.CancelAll()
	}
}
