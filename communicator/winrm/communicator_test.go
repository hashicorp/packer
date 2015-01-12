// +build !race

package winrm

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"os"
	"testing"
	"time"
)

func TestCommIsCommunicator(t *testing.T) {
	var raw interface{}
	raw = &comm{}
	if _, ok := raw.(packer.Communicator); !ok {
		t.Fatalf("comm must be a communicator")
	}
}

func TestStart(t *testing.T) {
	// This test hits an already running Windows VM
	// You can comment this line out temporarily during development
	t.Skip()

	comm, err := New("localhost:5985", "vagrant", "vagrant", time.Duration(1)*time.Minute)
	if err != nil {
		t.Fatalf("error connecting to WinRM: %s", err)
	}

	var cmd packer.RemoteCmd
	var outWriter, errWriter bytes.Buffer

	cmd.Command = "dir"
	cmd.Stdout = &outWriter
	cmd.Stderr = &errWriter

	err = comm.Start(&cmd)
	if err != nil {
		t.Fatalf("error starting cmd: %s", err)
	}
	cmd.Wait()

	fmt.Println(outWriter.String())
	fmt.Println(errWriter.String())

	if err != nil {
		t.Fatalf("error running cmd: %s", err)
	}

	if cmd.ExitStatus != 0 {
		t.Fatalf("exit status was non-zero: %d", cmd.ExitStatus)
	}
}

func TestStartElevated(t *testing.T) {
	// This test hits an already running Windows VM
	// You can comment this line out temporarily during development
	t.Skip()

	comm, err := New("localhost:5985", "vagrant", "vagrant", time.Duration(1)*time.Minute)
	if err != nil {
		t.Fatalf("error connecting to WinRM: %s", err)
	}

	var cmd packer.RemoteCmd
	var outWriter, errWriter bytes.Buffer

	cmd.Command = "dir"
	cmd.Stdout = &outWriter
	cmd.Stderr = &errWriter

	err = comm.StartElevated(&cmd)
	if err != nil {
		t.Fatalf("error starting cmd: %s", err)
	}
	cmd.Wait()

	fmt.Println(outWriter.String())
	fmt.Println(errWriter.String())

	if err != nil {
		t.Fatalf("error running cmd: %s", err)
	}

	if cmd.ExitStatus != 0 {
		t.Fatalf("exit status was non-zero: %d", cmd.ExitStatus)
	}
}

func TestUpload(t *testing.T) {
	// This test hits an already running Windows VM
	// You can comment this line out temporarily during development
	t.Skip()

	comm, err := New("localhost:5985", "vagrant", "vagrant", time.Duration(1)*time.Minute)
	if err != nil {
		t.Fatalf("error connecting to WinRM: %s", err)
	}

	f, err := os.Open("~/Downloads/photo1.jpg")
	if err != nil {
		t.Fatalf("error opening file: %s", err)
	}
	defer f.Close()

	err = comm.Upload("c:\\photo1.jpg", f, nil)
	if err != nil {
		t.Fatalf("error uploading file: %s", err)
	}
}

func TestUploadDir(t *testing.T) {
	// This test hits an already running Windows VM
	// You can comment this line out temporarily during development
	t.Skip()

	comm, err := New("localhost:5985", "vagrant", "vagrant", time.Duration(1)*time.Minute)
	if err != nil {
		t.Fatalf("error connecting to WinRM: %s", err)
	}

	err = comm.UploadDir("c:\\src\\chef-repo", "~/src/chef-repo", nil)
	if err != nil {
		t.Fatalf("error uploading dir: %s", err)
	}
}
