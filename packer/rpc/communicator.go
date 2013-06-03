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

type CommunicatorStartArgs struct {
	Command       string
	StdinAddress string
	StdoutAddress string
	StderrAddress string
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

func (c *communicator) Start(cmd *packer.RemoteCmd) (err error) {
	var args CommunicatorStartArgs
	args.Command = cmd.Command

	if cmd.Stdin != nil {
		stdinL := netListenerInRange(portRangeMin, portRangeMax)
		args.StdinAddress = stdinL.Addr().String()
		go serveSingleCopy("stdin", stdinL, nil, cmd.Stdin)
	}

	if cmd.Stdout != nil {
		stdoutL := netListenerInRange(portRangeMin, portRangeMax)
		args.StdoutAddress = stdoutL.Addr().String()
		go serveSingleCopy("stdout", stdoutL, cmd.Stdout, nil)
	}

	if cmd.Stderr != nil {
		stderrL := netListenerInRange(portRangeMin, portRangeMax)
		args.StderrAddress = stderrL.Addr().String()
		go serveSingleCopy("stderr", stderrL, cmd.Stderr, nil)
	}

	err = c.client.Call("Communicator.Start", &args, new(interface{}))
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

func (c *CommunicatorServer) Start(args *CommunicatorStartArgs, reply *interface{}) (err error) {
	// Build the RemoteCmd on this side so that it all pipes over
	// to the remote side.
	var cmd packer.RemoteCmd
	cmd.Command = args.Command

	if args.StdinAddress != "" {
		stdinC, err := net.Dial("tcp", args.StdinAddress)
		if err != nil {
			return err
		}

		cmd.Stdin = stdinC
	}

	if args.StdoutAddress != "" {
		stdoutC, err := net.Dial("tcp", args.StdoutAddress)
		if err != nil {
			return err
		}

		cmd.Stdout = stdoutC
	}

	if args.StderrAddress != "" {
		stderrC, err := net.Dial("tcp", args.StderrAddress)
		if err != nil {
			return err
		}

		cmd.Stderr = stderrC
	}

	// Start the actual command
	err = c.c.Start(&cmd)
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
