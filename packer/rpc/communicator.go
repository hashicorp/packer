package rpc

import (
	"errors"
	"github.com/mitchellh/packer/packer"
	"io"
	"log"
	"net"
	"net/rpc"
)

// An implementation of packer.Communicator where the communicator is actually
// executed over an RPC connection.
type communicator struct {
	client *rpc.Client
}

// CommunicatorServer wraps a packer.Communicator implementation and makes
// it exportable as part of a Golang RPC server.
type CommunicatorServer struct {
	c packer.Communicator
}

// RemoteCommandServer wraps a packer.RemoteCommand struct and makes it
// exportable as part of a Golang RPC server.
type RemoteCommandServer struct {
	rc *packer.RemoteCommand
}

type CommunicatorStartResponse struct {
	StdinAddress         string
	StdoutAddress        string
	StderrAddress        string
	RemoteCommandAddress string
}

type CommunicatorDownloadArgs struct {
	Path          string
	WriterAddress string
}

type CommunicatorUploadArgs struct {
	Path          string
	ReaderAddress string
}

func Communicator(client *rpc.Client) *communicator {
	return &communicator{client}
}

func (c *communicator) Start(cmd string) (rc *packer.RemoteCommand, err error) {
	var response CommunicatorStartResponse
	err = c.client.Call("Communicator.Start", &cmd, &response)
	if err != nil {
		return
	}

	// Connect to the three streams that will handle stdin, stdout,
	// and stderr and get net.Conns for them.
	stdinC, err := net.Dial("tcp", response.StdinAddress)
	if err != nil {
		return
	}

	stdoutC, err := net.Dial("tcp", response.StdoutAddress)
	if err != nil {
		return
	}

	stderrC, err := net.Dial("tcp", response.StderrAddress)
	if err != nil {
		return
	}

	// Connect to the RPC server for the remote command
	client, err := rpc.Dial("tcp", response.RemoteCommandAddress)
	if err != nil {
		return
	}

	// Build the response object using the streams we created
	rc = &packer.RemoteCommand{
		stdinC,
		stdoutC,
		stderrC,
		false,
		-1,
	}

	// In a goroutine, we wait for the process to exit, then we set
	// that it has exited.
	go func() {
		client.Call("RemoteCommand.Wait", new(interface{}), &rc.ExitStatus)
		rc.Exited = true
	}()

	return
}

func (c *communicator) Upload(path string, r io.Reader) (err error) {
	// We need to create a server that can proxy the reader data
	// over because we can't simply gob encode an io.Reader
	readerL := netListenerInRange(portRangeMin, portRangeMax)
	if readerL == nil {
		err = errors.New("couldn't allocate listener for upload reader")
		return
	}

	// Make sure at the end of this call, we close the listener
	defer readerL.Close()

	// Pipe the reader through to the connection
	go serveSingleCopy("uploadReader", readerL, nil, r)

	args := CommunicatorUploadArgs{
		path,
		readerL.Addr().String(),
	}

	err = c.client.Call("Communicator.Upload", &args, new(interface{}))
	return
}

func (c *communicator) Download(path string, w io.Writer) (err error) {
	// We need to create a server that can proxy that data downloaded
	// into the writer because we can't gob encode a writer directly.
	writerL := netListenerInRange(portRangeMin, portRangeMax)
	if writerL == nil {
		err = errors.New("couldn't allocate listener for download writer")
		return
	}

	// Make sure we close the listener once we're done because we'll be done
	defer writerL.Close()

	// Serve a single connection and a single copy
	go serveSingleCopy("downloadWriter", writerL, w, nil)

	args := CommunicatorDownloadArgs{
		path,
		writerL.Addr().String(),
	}

	err = c.client.Call("Communicator.Download", &args, new(interface{}))
	return
}

func (c *CommunicatorServer) Start(cmd *string, reply *CommunicatorStartResponse) (err error) {
	// Start executing the command.
	command, err := c.c.Start(*cmd)
	if err != nil {
		return
	}

	// If we didn't get a proper command... that's not right.
	if command == nil {
		return errors.New("communicator returned nil remote command")
	}

	// Next, we need to take the stdin/stdout and start a listener
	// for each because the client will connect to us via TCP and use
	// that connection as the io.Reader or io.Writer. These exist for
	// only a single connection that is persistent.
	stdinL := netListenerInRange(portRangeMin, portRangeMax)
	stdoutL := netListenerInRange(portRangeMin, portRangeMax)
	stderrL := netListenerInRange(portRangeMin, portRangeMax)
	go serveSingleCopy("stdin", stdinL, command.Stdin, nil)
	go serveSingleCopy("stdout", stdoutL, nil, command.Stdout)
	go serveSingleCopy("stderr", stderrL, nil, command.Stderr)

	// For the exit status, we use a simple RPC Server that serves
	// some of the RemoteComand methods.
	server := rpc.NewServer()
	server.RegisterName("RemoteCommand", &RemoteCommandServer{command})

	*reply = CommunicatorStartResponse{
		stdinL.Addr().String(),
		stdoutL.Addr().String(),
		stderrL.Addr().String(),
		serveSingleConn(server),
	}

	return
}

func (c *CommunicatorServer) Upload(args *CommunicatorUploadArgs, reply *interface{}) (err error) {
	readerC, err := net.Dial("tcp", args.ReaderAddress)
	if err != nil {
		return
	}

	defer readerC.Close()

	err = c.c.Upload(args.Path, readerC)
	return
}

func (c *CommunicatorServer) Download(args *CommunicatorDownloadArgs, reply *interface{}) (err error) {
	writerC, err := net.Dial("tcp", args.WriterAddress)
	if err != nil {
		return
	}

	defer writerC.Close()

	err = c.c.Download(args.Path, writerC)
	return
}

func (rc *RemoteCommandServer) Wait(args *interface{}, reply *int) error {
	rc.rc.Wait()
	*reply = rc.rc.ExitStatus
	return nil
}

func serveSingleCopy(name string, l net.Listener, dst io.Writer, src io.Reader) {
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		log.Printf("'%s' accept error: %s", name, err)
		return
	}

	// Be sure to close the connection after we're done copying so
	// that an EOF will successfully be sent to the remote side
	defer conn.Close()

	// The connection is the destination/source that is nil
	if dst == nil {
		dst = conn
	} else {
		src = conn
	}

	written, err := io.Copy(dst, src)
	log.Printf("%d bytes written for '%s'", written, name)
	if err != nil {
		log.Printf("'%s' copy error: %s", name, err)
	}
}
