package common

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"regexp"
)

// IfconfigIPFinder finds the host IP based on the output of `ip address` or `ifconfig`.
type IfconfigIPFinder struct {
	Device string
}

func (f *IfconfigIPFinder) HostIP() (string, error) {
	ip, err := ipaddress(f.Device)
	if err != nil || ip == "" {
		return ifconfig(f.Device)
	}
	return ip, err
}

func ipaddress(device string) (string, error) {
	var ipPath string

	// On some systems, ip is in /sbin which is generally not
	// on the PATH for a standard user, so we just check that first.
	if _, err := os.Stat("/sbin/ip"); err == nil {
		ipPath = "/sbin/ip"
	}

	if ipPath == "" {
		var err error
		ipPath, err = exec.LookPath("ip")
		if err != nil {
			return "", err
		}
	}

	stdout := new(bytes.Buffer)
	cmd := exec.Command(ipPath, "address", "show", "dev", device)
	// Force LANG=C so that the output is what we expect it to be
	// despite the locale.
	cmd.Env = append(cmd.Env, "LANG=C")
	cmd.Env = append(cmd.Env, os.Environ()...)

	cmd.Stdout = stdout
	cmd.Stderr = new(bytes.Buffer)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	re := regexp.MustCompile(`inet[^\d]+([\d\.]+)/`)
	matches := re.FindStringSubmatch(stdout.String())
	if matches == nil {
		return "", errors.New("IP not found in ip a output...")
	}

	return matches[1], nil
}

func ifconfig(device string) (string, error) {
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

	stdout := new(bytes.Buffer)

	cmd := exec.Command(ifconfigPath, device)
	// Force LANG=C so that the output is what we expect it to be
	// despite the locale.
	cmd.Env = append(cmd.Env, "LANG=C")
	cmd.Env = append(cmd.Env, os.Environ()...)

	cmd.Stdout = stdout
	cmd.Stderr = new(bytes.Buffer)
	if err := cmd.Run(); err != nil {
		return "", err
	}

	re := regexp.MustCompile(`inet[^\d]+([\d\.]+)\s`)
	matches := re.FindStringSubmatch(stdout.String())
	if matches == nil {
		return "", errors.New("IP not found in ifconfig output...")
	}

	return matches[1], nil
}
