package useragent

import (
	"fmt"
	"runtime"
)

var (
	// projectURL is the project URL.
	projectURL = "https://www.packer.io/"

	// rt is the runtime - variable for tests.
	rt = runtime.Version()

	// goos is the os - variable for tests.
	goos = runtime.GOOS

	// goarch is the architecture - variable for tests.
	goarch = runtime.GOARCH
)

// String returns the consistent user-agent string for Packer.
func String(packerVersion string) string {
	return fmt.Sprintf("Packer/%s (+%s; %s; %s/%s)",
		packerVersion, projectURL, rt, goos, goarch)
}
