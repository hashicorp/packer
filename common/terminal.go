package common

// call into one of the platform-specific implementations to get the current terminal dimensions
func GetTerminalDimensions() (width, height int, err error) {
	return platformGetTerminalDimensions()
}
