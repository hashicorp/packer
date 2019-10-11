package wrappedreadline

// getWidth impl for Solaris
func getWidth() int {
	return 80
}

// get width of the terminal
func getWidthFd(stdoutFd int) int {
	return getWidth()
}
