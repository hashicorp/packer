package ssh

import (
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"github.com/mitchellh/packer/packer"
	"log"
	"net"
)

type comm struct {
	client *ssh.ClientConn
}

// Creates a new packer.Communicator implementation over SSH. This takes
// an already existing TCP connection and SSH configuration.
func New(c net.Conn, config *ssh.ClientConfig) (result *comm, err error) {
	client, err := ssh.Client(c, config)
	result = &comm{client}
	return
}

func (c *comm) Start(cmd string) (remote *packer.RemoteCommand, err error) {
	session, err := c.client.NewSession()
	if err != nil {
		return
	}

	// Create the buffers to store our stdin/stdout/stderr
	stdin := new(bytes.Buffer)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	// Setup our session
	session.Stdin = stdin
	session.Stdout = stdout
	session.Stderr = stderr

	// Setup the remote command
	remote = &packer.RemoteCommand{
		stdin,
		stdout,
		stderr,
		false,
		-1,
	}

	log.Printf("starting remote command: %s", cmd)
	err = session.Start(cmd + "\n")
	if err != nil {
		return
	}

	// Start a goroutine to wait for the session to end and set the
	// exit boolean and status.
	go func() {
		defer session.Close()

		err := session.Wait()
		remote.ExitStatus = 0
		if err != nil {
			exitErr, ok := err.(*ssh.ExitError)
			if ok {
				remote.ExitStatus = exitErr.ExitStatus()
			}
		}

		remote.Exited = true
	}()

	return
}
