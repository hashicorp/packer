package rpc

import (
	"bufio"
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"io"
	"net/rpc"
	"testing"
	"time"
)

type testCommunicator struct {
	startCalled bool
	startCmd    *packer.RemoteCmd

	uploadCalled bool
	uploadPath   string
	uploadData   string

	downloadCalled bool
	downloadPath   string
}

func (t *testCommunicator) Start(cmd *packer.RemoteCmd) error {
	t.startCalled = true
	t.startCmd = cmd
	return nil
}

func (t *testCommunicator) Upload(path string, reader io.Reader) (err error) {
	t.uploadCalled = true
	t.uploadPath = path
	t.uploadData, err = bufio.NewReader(reader).ReadString('\n')
	return
}

func (t *testCommunicator) Download(path string, writer io.Writer) error {
	t.downloadCalled = true
	t.downloadPath = path
	writer.Write([]byte("download\n"))

	return nil
}

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
	remote := Communicator(client)

	// The remote command we'll use
	stdin_r, stdin_w := io.Pipe()
	stdout_r, stdout_w := io.Pipe()
	stderr_r, stderr_w := io.Pipe()

	var cmd packer.RemoteCmd
	cmd.Command = "foo"
	cmd.Stdin = stdin_r
	cmd.Stdout = stdout_w
	cmd.Stderr = stderr_w

	// Test Start
	err = remote.Start(&cmd)
	assert.Nil(err, "should not have an error")

	// Test that we can read from stdout
	c.startCmd.Stdout.Write([]byte("outfoo\n"))
	bufOut := bufio.NewReader(stdout_r)
	data, err := bufOut.ReadString('\n')
	assert.Nil(err, "should have no problem reading stdout")
	assert.Equal(data, "outfoo\n", "should be correct stdout")

	// Test that we can read from stderr
	c.startCmd.Stderr.Write([]byte("errfoo\n"))
	bufErr := bufio.NewReader(stderr_r)
	data, err = bufErr.ReadString('\n')
	assert.Nil(err, "should have no problem reading stderr")
	assert.Equal(data, "errfoo\n", "should be correct stderr")

	// Test that we can write to stdin
	stdin_w.Write([]byte("infoo\n"))
	bufIn := bufio.NewReader(c.startCmd.Stdin)
	data, err = bufIn.ReadString('\n')
	assert.Nil(err, "should have no problem reading stdin")
	assert.Equal(data, "infoo\n", "should be correct stdin")

	// Test that we can get the exit status properly
	c.startCmd.SetExited(42)

	for i := 0; i < 5; i++ {
		if cmd.Exited {
			assert.Equal(cmd.ExitStatus, 42, "should have proper exit status")
			break
		}

		time.Sleep(50 * time.Millisecond)
	}

	assert.True(cmd.Exited, "should have exited")

	// Test that we can upload things
	uploadR, uploadW := io.Pipe()
	go uploadW.Write([]byte("uploadfoo\n"))
	err = remote.Upload("foo", uploadR)
	assert.Nil(err, "should not error")
	assert.True(c.uploadCalled, "should be called")
	assert.Equal(c.uploadPath, "foo", "should be correct path")
	assert.Equal(c.uploadData, "uploadfoo\n", "should have the proper data")

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

	err = remote.Download("bar", downloadW)
	assert.Nil(err, "should not error")
	assert.True(c.downloadCalled, "should have called download")
	assert.Equal(c.downloadPath, "bar", "should have correct download path")

	<-downloadDone
	assert.Nil(downloadErr, "should not error reading download data")
	assert.Equal(downloadData, "download\n", "should have the proper data")
}

func TestCommunicator_ImplementsCommunicator(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var r packer.Communicator
	c := Communicator(nil)

	assert.Implementor(c, &r, "should be a Communicator")
}
