package utils

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

func IsDevNative() bool {
	return Config().Native != nil && Config().Native.SrcDir != ""
}

// ----------------------------------------------------------------------------
// Do we work with the local versions of some dependencies?
// ----------------------------------------------------------------------------

var swapCode bool

func UseCodeSwaps() {
	swapCode = true
}
