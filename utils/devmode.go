package utils

// ----------------------------------------------------------------------------
// Properties coming from the command flags
// ----------------------------------------------------------------------------

var (
	onlyAPI bool
)

func SetOnlyAPI(isOnlyAPI bool) {
	onlyAPI = isOnlyAPI
}

// ----------------------------------------------------------------------------
// The possible (al)development modes
// ----------------------------------------------------------------------------

func IsDevWebApp() bool {
	return !onlyAPI && Config().Web != nil && Config().Web.SrcDir != ""
}

func IsDevAPI() bool {
	return Config().API != nil && Config().API.SrcDir != ""
}

func IsDevLibrary() bool {
	return !onlyAPI && Config().Lib != nil && Config().Lib.SrcDir != ""
}

func IsDevGoSrc() bool {
	return IsDevAPI() || IsDevLibrary()
}

func IsDevNative() bool {
	return !onlyAPI && Config().Native != nil && Config().Native.SrcDir != ""
}

// // ----------------------------------------------------------------------------
// // Do we work with the local versions of some dependencies?
// // ----------------------------------------------------------------------------

// var swapCode bool

// func UseCodeSwaps() {
// 	swapCode = true
// }
