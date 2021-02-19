package common

import (
	"os/exec"
)

const VMWARE_PLAYER_VERSION = "6"

// Player6Driver is a driver that can run VMware Player 6
// installations.

type Player6Driver struct {
	Player5Driver
}

func NewPlayer6Driver(config *SSHConfig) Driver {
	return &Player6Driver{
		Player5Driver: Player5Driver{
			SSHConfig: config,
		},
	}
}

func (d *Player6Driver) Clone(dst, src string, linked bool, snapshot string) error {
	// TODO(rasa) check if running player+, not just player

	var cloneType string
	if linked {
		cloneType = "linked"
	} else {
		cloneType = "full"
	}

	args := []string{"-T", "ws", "clone", src, dst, cloneType}
	if snapshot != "" {
		args = append(args, "-snapshot", snapshot)
	}
	cmd := exec.Command(d.Player5Driver.VmrunPath, args...)
	if _, _, err := runAndLog(cmd); err != nil {
		return err
	}

	return nil
}

func (d *Player6Driver) Verify() error {
	if err := d.Player5Driver.Verify(); err != nil {
		return err
	}

	return playerVerifyVersion(VMWARE_PLAYER_VERSION)
}

func (d *Player6Driver) GetVmwareDriver() VmwareDriver {
	return d.Player5Driver.VmwareDriver
}
