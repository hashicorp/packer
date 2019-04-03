package packer

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
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
	ctx := context.TODO()

	err := rc.RunWithUi(ctx, testComm, testUi)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// sometimes cmd has returned and everything can be printed later on
	time.Sleep(1 * time.Second)

	expected := strings.TrimSpace(data)
	if diff := cmp.Diff(strings.TrimSpace(uiOutput.String()), expected); diff != "" {
		t.Fatalf("bad output: %s", diff)
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
