// +build windows

package chroot

import "errors"

func lockFile(*os.File) error {
	return errors.New("not supported on Windows")
}

func unlockFile(f *os.File) error {
	return nil
}
