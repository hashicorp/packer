// +build !windows

package googlecompute

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func NewTunnelDriver() TunnelDriver {
	return &TunnelDriverLinux{}
}

type TunnelDriverLinux struct {
	cmd *exec.Cmd
}

func (t *TunnelDriverLinux) StartTunnel(cancelCtx context.Context, tempScriptFileName string) error {
	// set stdout and stderr so we can read what's going on.
	var stdout, stderr bytes.Buffer

	cmd := exec.CommandContext(cancelCtx, tempScriptFileName)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err := cmd.Start()
	log.Printf("Waiting 30s for tunnel to create...")
	if err != nil {
		err := fmt.Errorf("Error calling gcloud sdk to launch IAP tunnel: %s",
			err)
		return err
	}

	time.Sleep(10 * time.Second)
	// read stdout
	scanner := bufio.NewScanner(&stderr)
	log.Printf("[start-iap-tunnel] stderr is:")
	for scanner.Scan() {
		line := scanner.Text()
		log.Println(line)

		if strings.Contains(line, "ERROR") {
			errIdx := strings.Index(line, "ERROR:")
			return fmt.Errorf("ERROR: %s", line[errIdx+7:])
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error scanning tunnel stderr: %s", err)
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
		// need to use the process _group_ id to kill this process and its
		// daemon child. We create the group ID with the syscall.SysProcAttr
		// call inside the retry loop above, and then store that ID on the
		// command so we can destroy it here.
		err := syscall.Kill(-t.cmd.Process.Pid, syscall.SIGINT)
		if err != nil {
			log.Printf("Issue stopping IAP tunnel: %s", err)
		}
	} else {
		log.Printf("Couldn't find IAP tunnel process to kill. Continuing.")
	}
}
