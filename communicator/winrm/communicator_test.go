package winrm

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/dylanmei/winrmtest"
	"github.com/mitchellh/packer/packer"
)

func newMockWinRMServer(t *testing.T) *winrmtest.Remote {
	wrm := winrmtest.NewRemote()

	wrm.CommandFunc(
		winrmtest.MatchText("echo foo"),
		func(out, err io.Writer) int {
			out.Write([]byte("foo"))
			return 0
		})

	wrm.CommandFunc(
		winrmtest.MatchPattern(`^echo c29tZXRoaW5n >> ".*"$`),
		func(out, err io.Writer) int {
			return 0
		})

	wrm.CommandFunc(
		winrmtest.MatchPattern(`^powershell.exe -EncodedCommand .*$`),
		func(out, err io.Writer) int {
			return 0
		})

	wrm.CommandFunc(
		winrmtest.MatchText("powershell"),
		func(out, err io.Writer) int {
			return 0
		})

	return wrm
}

func TestStart(t *testing.T) {
	wrm := newMockWinRMServer(t)
	defer wrm.Close()

	c, err := New(&Config{
		Host:     wrm.Host,
		Port:     wrm.Port,
		Username: "user",
		Password: "pass",
		Timeout:  30 * time.Second,
	})
	if err != nil {
		t.Fatalf("error creating communicator: %s", err)
	}

	var cmd packer.RemoteCmd
	stdout := new(bytes.Buffer)
	cmd.Command = "echo foo"
	cmd.Stdout = stdout

	err = c.Start(&cmd)
	if err != nil {
		t.Fatalf("error executing remote command: %s", err)
	}
	cmd.Wait()

	if stdout.String() != "foo" {
		t.Fatalf("bad command response: expected %q, got %q", "foo", stdout.String())
	}
}

func TestUpload(t *testing.T) {
	wrm := newMockWinRMServer(t)
	defer wrm.Close()

	c, err := New(&Config{
		Host:     wrm.Host,
		Port:     wrm.Port,
		Username: "user",
		Password: "pass",
		Timeout:  30 * time.Second,
	})
	if err != nil {
		t.Fatalf("error creating communicator: %s", err)
	}

	err = c.Upload("C:/Temp/terraform.cmd", bytes.NewReader([]byte("something")), nil)
	if err != nil {
		t.Fatalf("error uploading file: %s", err)
	}
}
