// +build windows

package googlecompute

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"time"
)

func NewTunnelDriver() TunnelDriver {
	log.Printf("Megan created driver for windows")
	return &TunnelDriverWindows{}
}

type TunnelDriverWindows struct {
	cmd *exec.Cmd
}

func (t *TunnelDriverWindows) StartTunnel(cancelCtx context.Context, tempScriptFileName string) error {
	log.Printf("Megan inside StartTunnel for windows")
	// set stdout and stderr so we can read what's going on.
	var stdout, stderr bytes.Buffer

	args := []string{"/C", "call", tempScriptFileName}

	cmd := exec.CommandContext(cancelCtx, "cmd", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Start()
	if err != nil {
		log.Printf("Megan: error calling start.")
		err := fmt.Errorf("Error calling gcloud sdk to launch IAP tunnel: %s",
			err)
		return err
	}

	// Give tunnel 30 seconds to either launch, or return an error.
	// Unfortunately, the SDK doesn't provide any official acknowledgment that
	// the tunnel is launched when it's not being run through a TTY so we
	// are just trusting here that 30s is enough to know whether the tunnel
	// launch was going to fail. Yep, feels icky to me too. But I spent an
	// afternoon trying to figure out how to get the SDK to actually send
	// the "Listening on port [n]" line I see when I run it manually, and I
	// can't justify spending more time than that on aesthetics.
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)

		lineStderr, err := stderr.ReadString('\n')
		if err != nil && err != io.EOF {
			log.Printf("Err from scanning stderr is %s", err)
			return fmt.Errorf("Error reading stderr from tunnel launch: %s", err)
		}
		if lineStderr != "" {
			log.Printf("stderr: %s", lineStderr)
		}

		lineStdout, err := stdout.ReadString('\n')
		if err != nil && err != io.EOF {
			log.Printf("Err from scanning stdout is %s", err)
			return fmt.Errorf("Error reading stdout from tunnel launch: %s", err)
		}
		if lineStdout != "" {
			log.Printf("stdout: %s", lineStdout)
		}

		if strings.Contains(lineStderr, "ERROR") {
			// 4033: Either you don't have permission to access the instance,
			// the instance doesn't exist, or the instance is stopped.
			// The two sub-errors we may see while the permissions settle are
			// "not authorized" and "failed to connect to backend," but after
			// about a minute of retries this goes away and we're able to
			// connect.
			if strings.Contains(lineStderr, "4033") {
				return RetryableTunnelError{lineStderr}
			} else {
				log.Printf("NOT RETRYABLE: %s", lineStderr)
				return fmt.Errorf("Non-retryable tunnel error: %s", lineStderr)
			}
		}
	}

	log.Printf("No error detected after tunnel launch; continuing...")

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
