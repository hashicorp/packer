// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func handleTermInterrupt(ui packersdk.Ui) (context.Context, func()) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	// Handle interrupts for this build
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	cleanup := func() {
		cancelCtx()
		signal.Stop(sigCh)
		close(sigCh)
	}
	go func() {
		select {
		case sig := <-sigCh:
			if sig == nil {
				// context got cancelled and this closed chan probably
				// triggered first
				return
			}
			ui.Error(fmt.Sprintf("Cancelling build after receiving %s", sig))
			cancelCtx()
		case <-ctx.Done():
		}
	}()
	return ctx, cleanup
}
