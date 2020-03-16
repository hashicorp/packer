// +build windows

package common

import (
	"syscall"
	"unsafe"
)

// windows constants and structures pulled from msdn
const (
	_STD_INPUT_HANDLE  = -10
	_STD_OUTPUT_HANDLE = -11
	_STD_ERROR_HANDLE  = -12
)

type (
	_SHORT int16
	_WORD  uint16

	_SMALL_RECT struct {
		Left, Top, Right, Bottom _SHORT
	}
	_COORD struct {
		X, Y _SHORT
	}
	_CONSOLE_SCREEN_BUFFER_INFO struct {
		dwSize, dwCursorPosition _COORD
		wAttributes              _WORD
		srWindow                 _SMALL_RECT
		dwMaximumWindowSize      _COORD
	}
)

// Low-level functions that call into Windows API for getting console info
var kernel32 = syscall.NewLazyDLL("kernel32.dll")
var kernel32_GetStdHandleProc = kernel32.NewProc("GetStdHandle")
var kernel32_GetConsoleScreenBufferInfoProc = kernel32.NewProc("GetConsoleScreenBufferInfo")

func kernel32_GetStdHandle(nStdHandle int32) (syscall.Handle, error) {
	res, _, err := kernel32_GetStdHandleProc.Call(uintptr(nStdHandle))
	if res == uintptr(syscall.InvalidHandle) {
		return syscall.InvalidHandle, error(err)
	}
	return syscall.Handle(res), nil
}

func kernel32_GetConsoleScreenBufferInfo(hConsoleOutput syscall.Handle, info *_CONSOLE_SCREEN_BUFFER_INFO) error {
	ok, _, err := kernel32_GetConsoleScreenBufferInfoProc.Call(uintptr(hConsoleOutput), uintptr(unsafe.Pointer(info)))
	if int(ok) == 0 {
		return error(err)
	}
	return nil
}

// windows api to get the console screen buffer info
func getConsoleScreenBufferInfo(csbi *_CONSOLE_SCREEN_BUFFER_INFO) (err error) {
	var (
		bi _CONSOLE_SCREEN_BUFFER_INFO
		fd syscall.Handle
	)

	// Re-open CONOUT$ as in some instances, stdout may be closed and guaranteed an stdout
	if fd, err = syscall.Open("CONOUT$", syscall.O_RDWR, 0); err != nil {
		return err
	}
	defer syscall.Close(fd)

	// grab the dimensions for the console
	if err = kernel32_GetConsoleScreenBufferInfo(fd, &bi); err != nil {
		return err
	}

	*csbi = bi
	return nil
}

func platformGetTerminalDimensions() (width, height int, err error) {
	var csbi _CONSOLE_SCREEN_BUFFER_INFO

	if err = getConsoleScreenBufferInfo(&csbi); err != nil {
		return 0, 0, err
	}

	return int(csbi.dwSize.X), int(csbi.dwSize.Y), nil
}
