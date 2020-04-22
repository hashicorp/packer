// +build windows

package googlecompute

import (
	"context"
	"fmt"
)

func NewTunnelDriver() TunnelDriver {
	return &TunnelDriverWindows{}
}

type TunnelDriverWindows struct {
}

func (t *TunnelDriverWindows) StartTunnel(cancelCtx context.Context, tempScriptFileName string) error {
	return fmt.Errorf("Windows support for IAP tunnel not yet supported.")
}

func (t *TunnelDriverWindows) StopTunnel() {}
