package iso

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"regexp"
)

// IfconfigIPFinder finds the host IP based on the output of `ifconfig`.
type IfconfigIPFinder struct {
	Devices []string
}

func (f *IfconfigIPFinder) HostIP() (string, error) {
	var ifconfigPath string

	// On some systems, ifconfig is in /sbin which is generally not
	// on the PATH for a standard user, so we just check that first.
	if _, err := os.Stat("/sbin/ifconfig"); err == nil {
		ifconfigPath = "/sbin/ifconfig"
	}

	if ifconfigPath == "" {
		var err error
		ifconfigPath, err = exec.LookPath("ifconfig")
		if err != nil {
			return "", err
		}
	}

	for _, device := range f.Devices {
		stdout := new(bytes.Buffer)

		cmd := exec.Command(ifconfigPath, device)
		cmd.Env = append(cmd.Env, os.Environ()...)

		// Force LANG=C so that the output is what we expect it to be
		// despite the locale.
		cmd.Env = append(cmd.Env, "LANG=C")

		cmd.Stdout = stdout
		cmd.Stderr = new(bytes.Buffer)

		if err := cmd.Run(); err == nil {
			re := regexp.MustCompile(`inet\s+(?:addr:)?(.+?)\s`)
			matches := re.FindStringSubmatch(stdout.String())
			if matches != nil {
				return matches[1], nil
			}
		}
	}
	return "", errors.New("IP not found in ifconfig output...")
}
