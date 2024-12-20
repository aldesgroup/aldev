package utils

// ----------------------------------------------------------------------------
// The possible (al)development modes
// ----------------------------------------------------------------------------

type DevMode int

const (
	DevModeUNDEFINED DevMode = 0
	DevModeAPI       DevMode = 1
	DevModeLIB       DevMode = 2
	DevModeWEB       DevMode = 3
	DevModeNATIVE    DevMode = 4
)

var (
	devMode_ DevMode
)

func b2i(boolValue bool) int {
	if boolValue {
		return 1
	}

	return 0
}

func SetDevMode(api, lib, web, native bool) {
	if b2i(api)+b2i(web)+b2i(lib)+b2i(native) != 1 {
		Fatal(nil, "a UNIQUE dev mode must be provided: --api (-a), --lib (-l), --web (-w) or --native (-n); see the help (aldev -h)")
	}

	switch {
	case api:
		devMode_ = DevModeAPI
	case lib:
		devMode_ = DevModeLIB
	case web:
		devMode_ = DevModeWEB
	case native:
		devMode_ = DevModeNATIVE
	}
}

func getDevMode() DevMode {
	if devMode_ == DevModeUNDEFINED {
		Fatal(nil, "Development mode is undefined!")
	}

	return devMode_
}

func IsDevWeb() bool {
	return getDevMode() == DevModeWEB
}

func IsDevAPI() bool {
	return getDevMode() == DevModeAPI || getDevMode() == DevModeWEB
}

func IsDevLibrary() bool {
	return getDevMode() == DevModeLIB
}

func IsDevNative() bool {
	return getDevMode() == DevModeNATIVE
}

// ----------------------------------------------------------------------------
// Do we work with the local versions of some dependencies?
// ----------------------------------------------------------------------------

var (
	useLocalDeps bool
)

func SetUseLocalDeps() {
	useLocalDeps = true
}
