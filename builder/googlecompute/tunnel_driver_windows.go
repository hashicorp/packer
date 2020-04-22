// +build windows

package googlecompute

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

func NewTunnelDriver() TunnelDriver {
	return &TunnelDriverWindows{}
}

type TunnelDriverWindows struct {
	cmd *exec.Cmd
}

func (t *TunnelDriverWindows) StartTunnel(cancelCtx context.Context, tempScriptFileName string) error {
	// set stdout and stderr so we can read what's going on.
	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(cancelCtx, tempScriptFileName)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Start()
	log.Printf("Waiting 30s for tunnel to create...")
	if err != nil {
		err := fmt.Errorf("Error calling gcloud sdk to launch IAP tunnel: %s",
			err)
		return err
	}
	// Wait for tunnel to launch and gather response. TODO: do this without
	// a sleep.
	time.Sleep(30 * time.Second)

	// Track stdout.
	sout := stdout.String()
	if sout != "" {
		log.Printf("[start-iap-tunnel] stdout is:")
	}

	log.Printf("[start-iap-tunnel] stderr is:")
	serr := stderr.String()
	log.Println(serr)
	if strings.Contains(serr, "ERROR") {
		errIdx := strings.Index(serr, "ERROR:")
		return fmt.Errorf("ERROR: %s", serr[errIdx+7:])
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
