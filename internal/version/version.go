package version

var (
	version = "< version filled by build >"
    buildInfo = "< build info filled by build >"
)

// Print version Details
func Print(name string) string {
	var versionInfo string
	versionInfo += name
	versionInfo += "\nversion: " + version
	versionInfo += "\nbuild info: " + buildInfo
	return versionInfo
}