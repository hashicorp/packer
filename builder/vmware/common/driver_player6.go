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

func (d *Player6Driver) Clone(dst, src string) error {
	// TODO(rasa) check if running player+, not just player

	cmd := exec.Command(d.Player5Driver.VmrunPath,
		"-T", "ws",
		"clone", src, dst,
		"full")

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
