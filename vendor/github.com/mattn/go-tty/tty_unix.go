// +build !windows
// +build !plan9

package tty

import (
	"bufio"
	"os"
	"os/signal"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

type TTY struct {
	in      *os.File
	bin     *bufio.Reader
	out     *os.File
	termios syscall.Termios
	ws      chan WINSIZE
	ss      chan os.Signal
}

func open() (*TTY, error) {
	tty := new(TTY)

	in, err := os.Open("/dev/tty")
	if err != nil {
		return nil, err
	}
	tty.in = in
	tty.bin = bufio.NewReader(in)

	out, err := os.OpenFile("/dev/tty", syscall.O_WRONLY, 0)
	if err != nil {
		return nil, err
	}
	tty.out = out

	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(tty.in.Fd()), ioctlReadTermios, uintptr(unsafe.Pointer(&tty.termios)), 0, 0, 0); err != 0 {
		return nil, err
	}
	newios := tty.termios
	newios.Iflag &^= syscall.ISTRIP | syscall.INLCR | syscall.ICRNL | syscall.IGNCR | syscall.IXON | syscall.IXOFF
	newios.Lflag &^= syscall.ECHO | syscall.ICANON /*| syscall.ISIG*/
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(tty.in.Fd()), ioctlWriteTermios, uintptr(unsafe.Pointer(&newios)), 0, 0, 0); err != 0 {
		return nil, err
	}

	tty.ws = make(chan WINSIZE)
	tty.ss = make(chan os.Signal, 1)
	signal.Notify(tty.ss, syscall.SIGWINCH)
	go func() {
		for sig := range tty.ss {
			switch sig {
			case syscall.SIGWINCH:
				if w, h, err := tty.size(); err == nil {
					tty.ws <- WINSIZE{
						W: w,
						H: h,
					}
				}
			default:
			}
		}
	}()
	return tty, nil
}

func (tty *TTY) buffered() bool {
	return tty.bin.Buffered() > 0
}

func (tty *TTY) readRune() (rune, error) {
	r, _, err := tty.bin.ReadRune()
	return r, err
}

func (tty *TTY) close() error {
	close(tty.ss)
	close(tty.ws)
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(tty.in.Fd()), ioctlWriteTermios, uintptr(unsafe.Pointer(&tty.termios)), 0, 0, 0)
	return err
}

func (tty *TTY) size() (int, int, error) {
	var dim [4]uint16
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(tty.out.Fd()), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&dim)), 0, 0, 0); err != 0 {
		return -1, -1, err
	}
	return int(dim[1]), int(dim[0]), nil
}

func (tty *TTY) input() *os.File {
	return tty.in
}

func (tty *TTY) output() *os.File {
	return tty.out
}

func (tty *TTY) raw() (func() error, error) {
	termios, err := unix.IoctlGetTermios(int(tty.in.Fd()), ioctlReadTermios)
	if err != nil {
		return nil, err
	}

	termios.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	termios.Oflag &^= unix.OPOST
	termios.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	termios.Cflag &^= unix.CSIZE | unix.PARENB
	termios.Cflag |= unix.CS8
	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0
	if err := unix.IoctlSetTermios(int(tty.in.Fd()), ioctlWriteTermios, termios); err != nil {
		return nil, err
	}

	return func() error {
		if err := unix.IoctlSetTermios(int(tty.in.Fd()), ioctlWriteTermios, termios); err != nil {
			return err
		}
		return nil
	}, nil
}

func (tty *TTY) sigwinch() chan WINSIZE {
	return tty.ws
}
