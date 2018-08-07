// +build windows

package common

import (
	"syscall"
	"unsafe"
)

// windows constants and structures pulled from msdn
const (
	STD_INPUT_HANDLE  = -10
	STD_OUTPUT_HANDLE = -11
	STD_ERROR_HANDLE  = -12
)

type (
	SHORT int16
	WORD  uint16

	SMALL_RECT struct {
		Left, Top, Right, Bottom SHORT
	}
	COORD struct {
		X, Y SHORT
	}
	CONSOLE_SCREEN_BUFFER_INFO struct {
		dwSize, dwCursorPosition COORD
		wAttributes              WORD
		srWindow                 SMALL_RECT
		dwMaximumWindowSize      COORD
	}
)

// Low-level functions that call into Windows API for getting console info
var KERNEL32 = syscall.NewLazyDLL("kernel32.dll")
var KERNEL32_GetStdHandleProc = KERNEL32.NewProc("GetStdHandle")
var KERNEL32_GetConsoleScreenBufferInfoProc = KERNEL32.NewProc("GetConsoleScreenBufferInfo")

func KERNEL32_GetStdHandle(nStdHandle int32) (syscall.Handle, error) {
	res, _, err := KERNEL32_GetStdHandleProc.Call(uintptr(nStdHandle))
	if res == uintptr(syscall.InvalidHandle) {
		return syscall.InvalidHandle, error(err)
	}
	return syscall.Handle(res), nil
}

func KERNEL32_GetConsoleScreenBufferInfo(hConsoleOutput syscall.Handle, info *CONSOLE_SCREEN_BUFFER_INFO) error {
	ok, _, err := KERNEL32_GetConsoleScreenBufferInfoProc.Call(uintptr(hConsoleOutput), uintptr(unsafe.Pointer(info)))
	if int(ok) == 0 {
		return error(err)
	}
	return nil
}

// windows api
func GetTerminalDimensions() (width, height int, err error) {
	var (
		fd   syscall.Handle
		csbi CONSOLE_SCREEN_BUFFER_INFO
	)

	// grab the handle for stdout
	/*
		if fd, err = KERNEL32_GetStdHandle(STD_OUTPUT_HANDLE); err != nil {
			return 0, 0, err
		}
	*/

	if fd, err = syscall.Open("CONOUT$", syscall.O_RDWR, 0); err != nil {
		return 0, 0, err
	}
	defer syscall.Close(fd)

	// grab the dimensions for the console
	if err = KERNEL32_GetConsoleScreenBufferInfo(fd, &csbi); err != nil {
		return 0, 0, err
	}

	// whee...
	return int(csbi.dwSize.X), int(csbi.dwSize.Y), nil
}
