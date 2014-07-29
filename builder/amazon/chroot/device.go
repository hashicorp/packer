package chroot

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AvailableDevice finds an available device and returns it. Note that
// you should externally hold a flock or something in order to guarantee
// that this device is available across processes.
func AvailableDevice() (string, error) {
	prefix, err := devicePrefix()
	if err != nil {
		return "", err
	}

	letters := "fghijklmnop"
	for _, letter := range letters {
		device := fmt.Sprintf("/dev/%s%c", prefix, letter)

		// If the block device itself, i.e. /dev/sf, exists, then we
		// can't use any of the numbers either.
		if _, err := os.Stat(device); err == nil {
			continue
		}

		// To be able to build both Paravirtual and HVM images, the unnumbered
		// device and the first numbered one must be available.
		// E.g. /dev/xvdf  and  /dev/xvdf1
		numbered_device := fmt.Sprintf("%s%d", device, 1)
		if _, err := os.Stat(numbered_device); err != nil {
			return device, nil
		}
	}

	return "", errors.New("available device could not be found")
}

// devicePrefix returns the prefix ("sd" or "xvd" or so on) of the devices
// on the system.
func devicePrefix() (string, error) {
	available := []string{"sd", "xvd"}

	f, err := os.Open("/sys/block")
	if err != nil {
		return "", err
	}
	defer f.Close()

	dirs, err := f.Readdirnames(-1)
	if dirs != nil && len(dirs) > 0 {
		for _, dir := range dirs {
			dirBase := filepath.Base(dir)
			for _, prefix := range available {
				if strings.HasPrefix(dirBase, prefix) {
					return prefix, nil
				}
			}
		}
	}

	if err != nil {
		return "", err
	}

	return "", errors.New("device prefix could not be detected")
}
