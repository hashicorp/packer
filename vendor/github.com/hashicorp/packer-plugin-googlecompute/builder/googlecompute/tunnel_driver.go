// +build !windows

package googlecompute

import (
	"context"
	"log"
	"os/exec"
	"syscall"
)

func NewTunnelDriver() TunnelDriver {
	return &TunnelDriverLinux{}
}

type TunnelDriverLinux struct {
	cmd *exec.Cmd
}

func (t *TunnelDriverLinux) StartTunnel(cancelCtx context.Context, tempScriptFileName string, timeout int) error {
	cmd := exec.CommandContext(cancelCtx, tempScriptFileName)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err := RunTunnelCommand(cmd, timeout)
	if err != nil {
		return err
	}

	// Store successful command on step so we can access it to cancel it
	// later.
	t.cmd = cmd
	return nil
}

func (t *TunnelDriverLinux) StopTunnel() {
	if t.cmd != nil && t.cmd.Process != nil {
		log.Printf("Cleaning up the IAP tunnel...")
		// Why not just cmd.Process.Kill()?  I'm glad you asked. The gcloud
		// call spawns a python subprocess that listens on the port, and you
		// need to use the process _group_ id to halt this process and its
		// daemon child. We create the group ID with the syscall.SysProcAttr
		// call inside the retry loop above, and then store that ID on the
		// command so we can halt it here.
		err := syscall.Kill(-t.cmd.Process.Pid, syscall.SIGINT)
		if err != nil {
			log.Printf("Issue stopping IAP tunnel: %s", err)
		}
	} else {
		log.Printf("Couldn't find IAP tunnel process to kill. Continuing.")
	}
}
