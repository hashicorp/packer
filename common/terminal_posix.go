// +build !windows

package common

// Imports for determining terminal information across platforms
import (
	"os"

	"golang.org/x/sys/unix"
)

// posix api
func platformGetTerminalDimensions() (width, height int, err error) {

	// grab the handle to stdin
	// XXX: in some cases, packer closes stdin, so the following can't be guaranteed
	/*
		tty := os.Stdin
	*/

	// open up a handle to the current tty
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return 0, 0, err
	}
	defer tty.Close()

	// convert the handle into a file descriptor
	fd := int(tty.Fd())

	// use it to make an Ioctl
	ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0, err
	}

	// return the width and height
	return int(ws.Col), int(ws.Row), nil
}
