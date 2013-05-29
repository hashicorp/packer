package packer

import (
	"testing"
	"time"
)

func TestRemoteCommand_WaitBlocks(t *testing.T) {
	t.Parallel()

	rc := &RemoteCommand{}

	complete := make(chan bool)

	// Make a goroutine that never exits. Since this is just in a test,
	// this should be okay.
	go func() {
		rc.Wait()
		complete <- true
	}()

	select {
	case <-complete:
		t.Fatal("It never should've completed")
	case <-time.After(500 * time.Millisecond):
		// All is well
	}
}

func TestRemoteCommand_WaitCompletes(t *testing.T) {
	t.Parallel()

	rc := &RemoteCommand{}

	complete := make(chan bool)
	go func() {
		rc.Wait()
		complete <- true
	}()

	// Flag that it completed
	rc.Exited = true

	select {
	case <-complete:
		// All is well
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for command completion.")
	}
}
