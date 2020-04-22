// +build windows

package googlecompute

import (
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
	return &TunnelDriverWindows{}
}

type TunnelDriverLinux struct {
	cmd *exec.Cmd
}

func (t *TunnelDriverWindows) StartTunnel(cancelCtx context.Context, tempScriptFileName string) error {
	return fmt.Errorf("Windows support for IAP tunnel not yet supported.")
}

func (t *TunnelDriverWindows) StopTunnel() {}
