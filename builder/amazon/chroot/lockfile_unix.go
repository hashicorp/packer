// +build !windows

package chroot

import (
	"errors"
	"os"
	"syscall"
)

// See: http://linux.die.net/include/sys/file.h
const LOCK_EX = 2
const LOCK_NB = 4
const LOCK_UN = 8

func lockFile(f *os.File) error {
	err := syscall.Flock(int(f.Fd()), LOCK_EX|LOCK_NB)
	if err != nil {
		errno, ok := err.(syscall.Errno)
		if ok && errno == syscall.EWOULDBLOCK {
			return errors.New("file already locked")
		}

		return err
	}

	return nil
}

func unlockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), LOCK_UN)
}
