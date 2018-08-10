// +build !windows

package packer

// Imports for determining terminal information across platforms
import (
	"golang.org/x/sys/unix"
	"os"
)

// posix api
func GetTerminalDimensions() (width, height int, err error) {

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
