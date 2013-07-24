package packer

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
)

type TestCommunicator struct {
	Stderr io.Reader
	Stdout io.Reader
}

func (c *TestCommunicator) Start(rc *RemoteCmd) error {
	go func() {
		if rc.Stdout != nil && c.Stdout != nil {
			io.Copy(rc.Stdout, c.Stdout)
		}

		if rc.Stderr != nil && c.Stderr != nil {
			io.Copy(rc.Stderr, c.Stderr)
		}
	}()

	return nil
}

func (c *TestCommunicator) Upload(string, io.Reader) error {
	return nil
}

func (c *TestCommunicator) Download(string, io.Writer) error {
	return nil
}

func TestRemoteCmd_StartWithUi(t *testing.T) {
	data := "hello\nworld\nthere"

	rcOutput := new(bytes.Buffer)
	uiOutput := new(bytes.Buffer)
	rcOutput.WriteString(data)

	testComm := &TestCommunicator{
		Stdout: rcOutput,
	}

	testUi := &ReaderWriterUi{
		Reader: new(bytes.Buffer),
		Writer: uiOutput,
	}

	rc := &RemoteCmd{
		Command: "test",
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		rc.Exited = true
	}()

	err := rc.StartWithUi(testComm, testUi)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if uiOutput.String() != strings.TrimSpace(data)+"\n" {
		t.Fatalf("bad output: '%s'", uiOutput.String())
	}
}

func TestRemoteCmd_Wait(t *testing.T) {
	var cmd RemoteCmd

	result := make(chan bool)
	go func() {
		cmd.Wait()
		result <- true
	}()

	cmd.ExitStatus = 42
	cmd.Exited = true

	select {
	case <-result:
		// Success
	case <-time.After(500 * time.Millisecond):
		t.Fatal("never got exit notification")
	}
}
