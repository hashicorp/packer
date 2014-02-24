package common

import (
        "os/exec"
        "bytes"
        "regexp"
        "fmt"
        "log"
        "strings"
        "runtime"
)

// Workstation10Driver is a driver that can run VMware Workstation 10
// installations. Current only tested for UNIX

type Workstation10Driver struct {
        Workstation9Driver
}

func (d *Workstation10Driver) Clone(dst, src string) error {
        cmd := exec.Command(d.Workstation9Driver.VmrunPath,
                "-T", "ws",
                "clone", src, dst,
                "full")

        if _, _, err := runAndLog(cmd); err != nil {
                return err
        }

        return nil
}

func (d *Workstation10Driver) Verify() error {
        if runtime.GOOS != "linux" {
                return fmt.Errorf("can't used driver WS 10 not yet supported on: %s", runtime.GOOS)
        }

        if err := d.Workstation9Driver.Verify(); err != nil {
                return err
        }


        //TODO(pmyjavec) there is a better way to find this, how?
        //the default will suffice for now.
        vmxpath := "/usr/lib/vmware/bin/vmware-vmx"

	var stderr bytes.Buffer
	cmd := exec.Command(vmxpath, "-v")
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	versionRe := regexp.MustCompile(`(?i)VMware Workstation (\d+\.\d+\.\d+)\s`)
	matches := versionRe.FindStringSubmatch(stderr.String())
	if matches == nil {
		return fmt.Errorf(
			"Couldn't find VMware WS version in output: %s", stderr.String())
	}
	log.Printf("Detected VMware WS version: %s", matches[1])

	if !strings.HasPrefix(matches[1], "10.") {
		return fmt.Errorf(
			"WS 10 not detected. Got version: %s", matches[1])
	}

        return  nil
}
