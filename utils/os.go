package utils

import "runtime"

func IsWindows() bool {
	return runtime.GOOS[:7] == "windows"
}

func IsLinux() bool {
	return runtime.GOOS[:5] == "linux"
}
