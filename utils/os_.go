//go:build !windows && !linux
// +build !windows,!linux

package utils

import "runtime"

func IsWindows() bool {
	Fatal(nil, "Platform '%s' is currently not supported right now!", runtime.GOOS)
	return false
}

func IsLinux() bool {
	Fatal(nil, "Platform '%s' is currently not supported right now!", runtime.GOOS)
	return false
}
