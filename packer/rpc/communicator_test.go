package rpc

import (
	"bufio"
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"io"
	"net/rpc"
	"testing"
)

type testCommunicator struct {
	startCalled bool
	startCmd string

	startIn *io.PipeReader
	startOut *io.PipeWriter
	startErr *io.PipeWriter
	startExited *bool
	startExitStatus *int
}

func (t *testCommunicator) Start(cmd string) (*packer.RemoteCommand, error) {
	t.startCalled = true
	t.startCmd = cmd

	var stdin *io.PipeWriter
	var stdout, stderr *io.PipeReader

	t.startIn, stdin = io.Pipe()
	stdout, t.startOut = io.Pipe()
	stderr, t.startErr = io.Pipe()

	rc := &packer.RemoteCommand{
		stdin,
		stdout,
		stderr,
		false,
		0,
	}

	t.startExited = &rc.Exited
	t.startExitStatus = &rc.ExitStatus

	return rc, nil
}

func (t *testCommunicator) Upload(string, io.Reader) {}

func (t *testCommunicator) Download(string, io.Writer) {}

func TestCommunicatorRPC(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	// Create the interface to test
	c := new(testCommunicator)

	// Start the server
	server := rpc.NewServer()
	RegisterCommunicator(server, c)
	address := serveSingleConn(server)

	// Create the client over RPC and run some methods to verify it works
	client, err := rpc.Dial("tcp", address)
	assert.Nil(err, "should be able to connect")

	// Test Start
	remote := Communicator(client)
	rc, err := remote.Start("foo")
	assert.Nil(err, "should not have an error")

	// Test that we can read from stdout
	bufOut := bufio.NewReader(rc.Stdout)
	c.startOut.Write([]byte("outfoo\n"))
	data, err := bufOut.ReadString('\n')
	assert.Nil(err, "should have no problem reading stdout")
	assert.Equal(data, "outfoo\n", "should be correct stdout")

	// Test that we can read from stderr
	bufErr := bufio.NewReader(rc.Stderr)
	c.startErr.Write([]byte("errfoo\n"))
	data, err = bufErr.ReadString('\n')
	assert.Nil(err, "should have no problem reading stdout")
	assert.Equal(data, "errfoo\n", "should be correct stdout")

	// Test that we can write to stdin
	bufIn := bufio.NewReader(c.startIn)
	rc.Stdin.Write([]byte("infoo\n"))
	data, err = bufIn.ReadString('\n')
	assert.Nil(err, "should have no problem reading stdin")
	assert.Equal(data, "infoo\n", "should be correct stdin")

	// Test that we can get the exit status properly
	*c.startExitStatus = 42
	*c.startExited = true
	rc.Wait()
	assert.Equal(rc.ExitStatus, 42, "should have proper exit status")
}

func TestCommunicator_ImplementsCommunicator(t *testing.T) {
	//assert := asserts.NewTestingAsserts(t, true)

	//var r packer.Communicator
	//c := Communicator(nil)

	//assert.Implementor(c, &r, "should be a Communicator")
}
