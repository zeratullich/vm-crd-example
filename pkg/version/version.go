package version

// version will be overridden with the current version at build time using the -X linker flag
var version = "v1.0.5-beta"

func GetVersion() string {
	return version
}