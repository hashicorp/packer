package main

import (
	"io"
	"log"
	"os"
	"os/signal"
)

// setupStdin switches out stdin for a pipe. We do this so that we can
// close the writer end of the pipe when we receive an interrupt so plugins
// blocked on reading from stdin are unblocked.
func setupStdin() {
	// Create the pipe and swap stdin for the reader end
	r, w, _ := os.Pipe()
	originalStdin := os.Stdin
	os.Stdin = r

	// Create a goroutine that copies data from the original stdin
	// into the writer end of the pipe forever.
	go func() {
		defer w.Close()
		io.Copy(w, originalStdin)
	}()

	// Register a signal handler for interrupt in order to close the
	// writer end of our pipe so that readers get EOF downstream.
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	go func() {
		defer signal.Stop(ch)
		defer w.Close()
		<-ch
		log.Println("Closing stdin because interrupt received.")
	}()
}
