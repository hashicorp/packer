package chroot

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

func (da diskAttacher) WaitForDevice(ctx context.Context, lun int32) (device string, err error) {
	// This builder will always be running in Azure, where data disks show up
	// on scbus5 target 0. The camcontrol command always outputs LUNs in
	// unpadded hexadecimal format.
	regexStr := fmt.Sprintf(`at scbus5 target 0 lun %x \(.*?da([\d]+)`, lun)
	devRegex, err := regexp.Compile(regexStr)
	if err != nil {
		return "", err
	}
	for {
		cmd := exec.Command("camcontrol", "devlist")
		var out bytes.Buffer
		cmd.Stdout = &out
		err = cmd.Run()
		if err != nil {
			return "", err
		}
		outString := out.String()
		scanner := bufio.NewScanner(strings.NewReader(outString))
		for scanner.Scan() {
			line := scanner.Text()
			// Check if this is the correct bus, target, and LUN.
			if matches := devRegex.FindStringSubmatch(line); matches != nil {
				// If this function immediately returns, devfs won't have
				// created the device yet.
				time.Sleep(1000 * time.Millisecond)
				return fmt.Sprintf("/dev/da%s", matches[1]), nil
			}
		}
		if err = scanner.Err(); err != nil {
			return "", err
		}

		select {
		case <-time.After(100 * time.Millisecond):
			// continue
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}
