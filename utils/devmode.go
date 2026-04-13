package utils

// ----------------------------------------------------------------------------
// Properties coming from the command flags
// ----------------------------------------------------------------------------

// ----------------------------------------------------------------------------
// The possible (al)development modes
// ----------------------------------------------------------------------------

func IsDevWebApp() bool {
	return Config().Web != nil && Config().Web.SrcDir != ""
}

func IsDevAPI() bool {
	return Config().API != nil && Config().API.SrcDir != ""
}

func IsDevLibrary() bool {
	return Config().Lib != nil && Config().Lib.SrcDir != ""
}

func IsDevGoSrc() bool {
	return IsDevAPI() || IsDevLibrary()
}

func IsDevNative() bool {
	return Config().Native != nil && Config().Native.SrcDir != ""
}
