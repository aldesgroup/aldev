package utils

import (
	"fmt"
	"runtime/debug"
)

func Recover(msg string, params ...any) {
	if err := recover(); err != nil {
		Error("Recovered from error [%v] while %s; Stack: %s", err, fmt.Sprintf(msg, params...), string(debug.Stack()))
	}
}
