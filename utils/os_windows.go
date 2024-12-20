//go:build windows
// +build windows

package utils

func IsWindows() bool {
	return true
}

func IsLinux() bool {
	return false
}
