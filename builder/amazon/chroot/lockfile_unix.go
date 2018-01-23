// +build !windows

package chroot

import (
	"os"

	"golang.org/x/sys/unix"
)

// See: http://linux.die.net/include/sys/file.h
const LOCK_EX = 2
const LOCK_NB = 4
const LOCK_UN = 8

func lockFile(f *os.File) error {
	err := unix.Flock(int(f.Fd()), LOCK_EX)
	if err != nil {
		return err
	}

	return nil
}

func unlockFile(f *os.File) error {
	return unix.Flock(int(f.Fd()), LOCK_UN)
}
