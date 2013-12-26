package common

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Fusion6Driver is a driver that can run VMWare Fusion 5.
type Fusion6Driver struct {
	Fusion5Driver
}

func (d *Fusion6Driver) Clone(dst, src string) error {
	return errors.New("Cloning is not supported with Fusion 5. Please use Fusion 6+.")
}

func (d *Fusion6Driver) Verify() error {
	if err := d.Fusion5Driver.Verify(); err != nil {
		return err
	}

	vmxpath := filepath.Join(d.AppPath, "Contents", "Library", "vmware-vmx")
	if _, err := os.Stat(vmxpath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("vmware-vmx could not be found at path: %s",
				vmxpath)
		}

		return err
	}

	var stderr bytes.Buffer
	cmd := exec.Command(vmxpath, "-v")
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	versionRe := regexp.MustCompile(`(?i)VMware [a-z0-9-]+ (\d+\.\d+\.\d+)\s`)
	matches := versionRe.FindStringSubmatch(stderr.String())
	if matches == nil {
		return fmt.Errorf(
			"Couldn't find VMware version in output: %s", stderr.String())
	}
	log.Printf("Detected VMware version: %s", matches[1])

	if !strings.HasPrefix(matches[1], "6.") {
		return fmt.Errorf(
			"Fusion 6 not detected. Got version: %s", matches[1])
	}

	return nil
}
