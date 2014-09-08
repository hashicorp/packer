package rpc

import (
	"bufio"
	"github.com/mitchellh/packer/packer"
	"io"
	"reflect"
	"testing"
)

func TestCommunicatorRPC(t *testing.T) {
	// Create the interface to test
	c := new(packer.MockCommunicator)

	// Start the server
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterCommunicator(c)
	remote := client.Communicator()

	// The remote command we'll use
	stdin_r, stdin_w := io.Pipe()
	stdout_r, stdout_w := io.Pipe()
	stderr_r, stderr_w := io.Pipe()

	var cmd packer.RemoteCmd
	cmd.Command = "foo"
	cmd.Stdin = stdin_r
	cmd.Stdout = stdout_w
	cmd.Stderr = stderr_w

	// Send some data on stdout and stderr from the mock
	c.StartStdout = "outfoo\n"
	c.StartStderr = "errfoo\n"
	c.StartExitStatus = 42

	// Test Start
	err := remote.Start(&cmd)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test that we can read from stdout
	bufOut := bufio.NewReader(stdout_r)
	data, err := bufOut.ReadString('\n')
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if data != "outfoo\n" {
		t.Fatalf("bad data: %s", data)
	}

	// Test that we can read from stderr
	bufErr := bufio.NewReader(stderr_r)
	data, err = bufErr.ReadString('\n')
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if data != "errfoo\n" {
		t.Fatalf("bad data: %s", data)
	}

	// Test that we can write to stdin
	stdin_w.Write([]byte("info\n"))
	stdin_w.Close()
	cmd.Wait()
	if c.StartStdin != "info\n" {
		t.Fatalf("bad data: %s", c.StartStdin)
	}

	// Test that we can get the exit status properly
	if cmd.ExitStatus != 42 {
		t.Fatalf("bad exit: %d", cmd.ExitStatus)
	}

	// Test that we can upload things
	uploadR, uploadW := io.Pipe()
	go func() {
		defer uploadW.Close()
		uploadW.Write([]byte("uploadfoo\n"))
	}()
	err = remote.Upload("foo", uploadR, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !c.UploadCalled {
		t.Fatal("should have uploaded")
	}

	if c.UploadPath != "foo" {
		t.Fatalf("path: %s", c.UploadPath)
	}

	if c.UploadData != "uploadfoo\n" {
		t.Fatalf("bad: %s", c.UploadData)
	}

	// Test that we can upload directories
	dirDst := "foo"
	dirSrc := "bar"
	dirExcl := []string{"foo"}
	err = remote.UploadDir(dirDst, dirSrc, dirExcl)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if c.UploadDirDst != dirDst {
		t.Fatalf("bad: %s", c.UploadDirDst)
	}

	if c.UploadDirSrc != dirSrc {
		t.Fatalf("bad: %s", c.UploadDirSrc)
	}

	if !reflect.DeepEqual(c.UploadDirExclude, dirExcl) {
		t.Fatalf("bad: %#v", c.UploadDirExclude)
	}

	// Test that we can download things
	downloadR, downloadW := io.Pipe()
	downloadDone := make(chan bool)
	var downloadData string
	var downloadErr error

	go func() {
		bufDownR := bufio.NewReader(downloadR)
		downloadData, downloadErr = bufDownR.ReadString('\n')
		downloadDone <- true
	}()

	c.DownloadData = "download\n"
	err = remote.Download("bar", downloadW)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !c.DownloadCalled {
		t.Fatal("download should be called")
	}

	if c.DownloadPath != "bar" {
		t.Fatalf("bad: %s", c.DownloadPath)
	}

	<-downloadDone
	if downloadErr != nil {
		t.Fatalf("err: %s", downloadErr)
	}

	if downloadData != "download\n" {
		t.Fatalf("bad: %s", downloadData)
	}
}

func TestCommunicator_ImplementsCommunicator(t *testing.T) {
	var raw interface{}
	raw = Communicator(nil)
	if _, ok := raw.(packer.Communicator); !ok {
		t.Fatal("should be a Communicator")
	}
}
