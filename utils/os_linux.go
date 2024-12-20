//go:build linux
// +build linux

package utils

func IsWindows() bool {
	return false
}

func IsLinux() bool {
	return true
}
