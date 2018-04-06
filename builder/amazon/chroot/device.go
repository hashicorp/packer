package chroot

import (
	"errors"
	"fmt"
	"os"
)

// AvailableDevice finds an available device and returns it. Note that
// you should externally hold a flock or something in order to guarantee
// that this device is available across processes.
func AvailableDevice() (string, error) {
	var device string
	availablePrefixes := []string{"sd", "xvd"}
	letters := "fghijklmnop"
letter_loop:
	for _, letter := range letters {
		deviceStrings := make([]string, len(availablePrefixes))
		for idx, prefix := range availablePrefixes {
			deviceStrings[idx] = fmt.Sprintf("/dev/%s%c", prefix, letter)
		}

		// If the block device itself, i.e. /dev/sf, exists, then we
		// can't use any of the numbers either.
		for _, device = range deviceStrings {
			if _, err := os.Stat(device); err == nil {
				continue letter_loop
			}
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
