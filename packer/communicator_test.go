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
	uiOutput := new(bytes.Buffer)

	testComm := new(MockCommunicator)
	testComm.StartStdout = data
	testUi := &BasicUi{
		Reader: new(bytes.Buffer),
		Writer: uiOutput,
	}

	rc := &RemoteCmd{
		Command: "test",
		Stdout:  originalOutput,
	}

	err := rc.StartWithUi(testComm, testUi)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	rc.Wait()

	expected := strings.TrimSpace(data)
	if strings.TrimSpace(uiOutput.String()) != expected {
		t.Fatalf("bad output: '%s'", uiOutput.String())
	}

	if originalOutput.String() != expected {
		t.Fatalf("bad: %#v", originalOutput.String())
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
