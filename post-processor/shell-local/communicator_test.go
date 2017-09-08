package shell_local

import (
	"bytes"
	"io/ioutil"
	"os"
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

	vars := []string{}
	shebang := []string{"/bin/sh", "-e"}
	c := &Communicator{
		vars,
		shebang,
	}

	// create a temporary script file
	tmpfile, err := ioutil.TempFile("", "script")
	if err != nil {
		t.Fatal(err)
	}

	defer os.Remove(tmpfile.Name())

	content := []byte("/bin/echo foo")

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	cmd := &packer.RemoteCmd{
		Command: tmpfile.Name(),
		Stdout:  &buf,
	}

	if err := c.Start(cmd); err != nil {
		t.Fatalf("err: %s", err)
	}

	cmd.Wait()

	// if cmd.ExitStatus != 0 {
	// 	t.Fatalf("err bad exit status: %d", cmd.ExitStatus)
	// }

	if strings.TrimSpace(buf.String()) != "foo" {
		t.Fatalf("bad: %s", buf.String())
	}
}
