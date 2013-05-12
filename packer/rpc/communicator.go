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
	StdinAddress string
	StdoutAddress string
	StderrAddress string
	RemoteCommandAddress string
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

func (c *communicator) Upload(string, io.Reader) {}

func (c *communicator) Download(string, io.Writer) {}

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

func (rc *RemoteCommandServer) Wait(args *interface{}, reply *int) error {
	rc.rc.Wait()
	*reply = rc.rc.ExitStatus
	return nil
}

func serveSingleCopy(name string, l net.Listener, dst io.Writer, src io.Reader) {
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		return
	}

	// The connection is the destination/source that is nil
	if dst == nil {
		dst = conn
	} else {
		src = conn
	}

	written, err := io.Copy(dst, src)
	log.Printf("%d bytes written for '%s'", written, name)
	if err != nil {
		log.Printf("'%s' copy error: %s", name, err.Error())
	}
}
