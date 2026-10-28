package utils

import core "github.com/aldesgroup/corego"

// ----------------------------------------------------------------------------
// Properties coming from the command flags
// ----------------------------------------------------------------------------

// ----------------------------------------------------------------------------
// The possible (al)development modes
// ----------------------------------------------------------------------------

func IsDevWebApp() bool {
	return Config().Web != nil && Config().Web.Build != nil && Config().Web.Build.SrcDir != ""
}

func IsDevAPI() bool {
	return Config().API != nil && Config().API.Build != nil && Config().API.Build.SrcDir != ""
}

func IsDevNative() bool {
	return Config().Native != nil && Config().Native.SrcDir != ""
}

func IsDevLibrary() bool {
	return Config().Lib != nil && Config().Lib.SrcDir != ""
}

func IsDevGoSrc() bool {
	return core.FileExists(getGoSrcDir(), "go.mod")
}

func IsDevWebSrc() bool {
	return (IsDevLibrary() && core.FileExists("package.json")) ||
		(IsDevWebApp() && core.FileExists(getWebSrcDir(), "package.json"))
}

// ----------------------------------------------------------------------------
// Some utils
// ----------------------------------------------------------------------------

func getGoSrcDir() string {
	if IsDevAPI() {
		return Config().API.Build.SrcDir
	}
	if IsDevLibrary() {
		return Config().Lib.SrcDir
	}
	return ""
}

func getWebSrcDir() string {
	if IsDevWebApp() {
		return Config().Web.Build.SrcDir
	}
	if IsDevLibrary() {
		return Config().Lib.SrcDir
	}
	return ""
}
