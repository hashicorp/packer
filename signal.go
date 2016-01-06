package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/packer/plugin"
)

// Prepares the signal handlers so that we handle interrupts properly.
// The signal handler exists in a goroutine.
func setupSignalHandlers(ui packer.Ui) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	signal.Notify(ch, syscall.SIGTERM)

	go func() {
		// First interrupt. We mostly ignore this because it allows the
		// plugins time to cleanup.
		<-ch
		log.Println("First interrupt. Ignoring to allow plugins to clean up.")

		ui.Error("Interrupt signal received. Cleaning up...")

		// Second interrupt. Go down hard.
		<-ch
		log.Println("Second interrupt. Exiting now.")

		ui.Error("Interrupt signal received twice. Forcefully exiting now.")

		// Force kill all the plugins, but mark that we're killing them
		// first so that we don't get panics everywhere.
		plugin.CleanupClients()
		os.Exit(1)
	}()
}
