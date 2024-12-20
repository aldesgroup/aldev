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
	devMode DevMode
)

func b2i(boolValue bool) int {
	if boolValue {
		return 1
	}

	return 0
}

var modeToString = map[DevMode]string{
	DevModeAPI:    "--api",
	DevModeLIB:    "--lib",
	DevModeWEB:    "--web",
	DevModeNATIVE: "--native",
}

// returns the current dev mode as a string
func DevModeToString() string {
	return modeToString[devMode]
}

func SetDevMode(api, lib, web, native bool) {
	if b2i(api)+b2i(web)+b2i(lib)+b2i(native) != 1 {
		Fatal(nil, "a UNIQUE dev mode must be provided: --api (-a), --lib (-l), --web (-w) or --native (-n); see the help (aldev -h)")
	}

	switch {
	case api:
		devMode = DevModeAPI
	case lib:
		devMode = DevModeLIB
	case web:
		devMode = DevModeWEB
	case native:
		devMode = DevModeNATIVE
	}
}

func getDevMode() DevMode {
	if devMode == DevModeUNDEFINED {
		Fatal(nil, "Development mode is undefined!")
	}

	return devMode
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
	swapCode bool
)

func UseCodeSwaps() {
	swapCode = true
}
