package shell_local

import (
	"bytes"
	"runtime"
	"strings"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestCommunicator_impl(t *testing.T) {
	var _ packer.Communicator = new(Communicator)
}

func TestCommunicator(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("windows not supported for this test")
		return
	}

	c := &Communicator{
		ExecuteCommand: []string{"/bin/sh", "-c", "echo foo"},
	}

	var buf bytes.Buffer
	cmd := &packer.RemoteCmd{
		Stdout: &buf,
	}

	if err := c.Start(cmd); err != nil {
		t.Fatalf("err: %s", err)
	}

	cmd.Wait()

	if cmd.ExitStatus != 0 {
		t.Fatalf("err bad exit status: %d", cmd.ExitStatus)
	}

	if strings.TrimSpace(buf.String()) != "foo" {
		t.Fatalf("bad: %s", buf.String())
	}
}
