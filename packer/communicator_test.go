package packer

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRemoteCmd_StartWithUi(t *testing.T) {
	data := "hello\nworld\nthere"

	originalOutput := new(bytes.Buffer)
	rcOutput := new(bytes.Buffer)
	uiOutput := new(bytes.Buffer)
	rcOutput.WriteString(data)

	testComm := &MockCommunicator{
		Stdout: rcOutput,
	}

	testUi := &BasicUi{
		Reader: new(bytes.Buffer),
		Writer: uiOutput,
	}

	rc := &RemoteCmd{
		Command: "test",
		Stdout:  originalOutput,
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		rc.SetExited(0)
	}()

	err := rc.StartWithUi(testComm, testUi)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if uiOutput.String() != strings.TrimSpace(data)+"\n" {
		t.Fatalf("bad output: '%s'", uiOutput.String())
	}

	if originalOutput.String() != data {
		t.Fatalf("original is bad: '%s'", originalOutput.String())
	}
}

func TestRemoteCmd_Wait(t *testing.T) {
	var cmd RemoteCmd

	result := make(chan bool)
	go func() {
		cmd.Wait()
		result <- true
	}()

	cmd.SetExited(42)

	select {
	case <-result:
		// Success
	case <-time.After(500 * time.Millisecond):
		t.Fatal("never got exit notification")
	}
}
