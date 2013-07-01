package vmware

import (
	"bytes"
	"errors"
	"os/exec"
	"regexp"
)

// Interface to help find the host IP that is available from within
// the VMware virtual machines.
type HostIPFinder interface {
	HostIP() (string, error)
}

// IfconfigIPFinder finds the host IP based on the output of `ifconfig`.
type IfconfigIPFinder struct {
	Device string
}

func (f *IfconfigIPFinder) HostIP() (string, error) {
	ifconfigPath, err := exec.LookPath("ifconfig")
	if err != nil {
		return "", err
	}

	stdout := new(bytes.Buffer)

	cmd := exec.Command(ifconfigPath, f.Device)
	cmd.Stdout = stdout
	cmd.Stderr = new(bytes.Buffer)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	re := regexp.MustCompile(`inet\s*(?:addr:)?(.+?)\s`)
	matches := re.FindStringSubmatch(stdout.String())
	if matches == nil {
		return "", errors.New("IP not found in ifconfig output...")
	}

	return matches[1], nil
}
