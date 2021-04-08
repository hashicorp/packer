// +build windows

package googlecompute

import (
	"context"
	"log"
	"os/exec"
)

func NewTunnelDriver() TunnelDriver {
	return &TunnelDriverWindows{}
}

type TunnelDriverWindows struct {
	cmd *exec.Cmd
}

func (t *TunnelDriverWindows) StartTunnel(cancelCtx context.Context, tempScriptFileName string, timeout int) error {
	args := []string{"/C", "call", tempScriptFileName}
	cmd := exec.CommandContext(cancelCtx, "cmd", args...)
	err := RunTunnelCommand(cmd, timeout)
	if err != nil {
		return err
	}
	// Store successful command on step so we can access it to cancel it
	// later.
	t.cmd = cmd
	return nil
}

func (t *TunnelDriverWindows) StopTunnel() {
	if t.cmd != nil && t.cmd.Process != nil {
		err := t.cmd.Process.Kill()
		if err != nil {
			log.Printf("Issue stopping IAP tunnel: %s", err)
		}
	} else {
		log.Printf("Couldn't find IAP tunnel process to kill. Continuing.")
	}
}
