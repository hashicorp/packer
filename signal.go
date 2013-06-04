package main

import (
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/packer/plugin"
	"log"
	"os"
	"os/signal"
)

// Prepares the signal handlers so that we handle interrupts properly.
// The signal handler exists in a goroutine.
func setupSignalHandlers(env packer.Environment) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	go func() {
		<-ch
		log.Println("First interrupt. Ignoring, will let plugins handle...")
		<-ch
		log.Println("Second interrupt. Exiting now.")

		env.Ui().Error("Interrupt signal received twice. Forcefully exiting now.")

		// Force kill all the plugins
		plugin.CleanupClients()
		os.Exit(1)
	}()
}
