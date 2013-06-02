package packer

import (
	"bytes"
	"testing"
	"time"
)

func TestRemoteCommand_ExitChan(t *testing.T) {
	t.Parallel()

	rc := &RemoteCommand{}
	exitChan := rc.ExitChan()

	// Set the exit data so that it is sent
	rc.ExitStatus = 42
	rc.Exited = true

	select {
	case exitCode := <-exitChan:
		if exitCode != 42 {
			t.Fatal("invalid exit code")
		}

		_, ok := <-exitChan
		if ok {
			t.Fatal("exit channel should be closed")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("exit channel never sent")
	}
}

func TestRemoteCommand_StdoutChan(t *testing.T) {
	expected := "DATA!!!"

	stdoutBuf := new(bytes.Buffer)
	stdoutBuf.WriteString(expected)

	rc := &RemoteCommand{}
	rc.Stdout = stdoutBuf

	outChan := rc.StdoutChan()

	results := new(bytes.Buffer)
	for data := range outChan {
		results.WriteString(data)
	}

	if results.String() != expected {
		t.Fatalf(
			"outputs didn't match:\ngot:\n%s\nexpected:\n%s",
			results.String(), stdoutBuf.String())
	}
}

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
