package ansible

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/mitchellh/packer/packer"
	"golang.org/x/crypto/ssh"
)

type adapter struct {
	done    <-chan struct{}
	l       net.Listener
	config  *ssh.ServerConfig
	sftpCmd string
	ui      packer.Ui
	comm    packer.Communicator
}

func newAdapter(done <-chan struct{}, l net.Listener, config *ssh.ServerConfig, sftpCmd string, ui packer.Ui, comm packer.Communicator) *adapter {
	return &adapter{
		done:    done,
		l:       l,
		config:  config,
		sftpCmd: sftpCmd,
		ui:      ui,
		comm:    comm,
	}
}

func (c *adapter) Serve() {
	c.ui.Say(fmt.Sprintf("SSH proxy: serving on %s", c.l.Addr()))

	for {
		// Accept will return if either the underlying connection is closed or if a connection is made.
		// after returning, check to see if c.done can be received. If so, then Accept() returned because
		// the connection has been closed.
		conn, err := c.l.Accept()
		select {
		case <-c.done:
			return
		default:
			if err != nil {
				c.ui.Error(fmt.Sprintf("listen.Accept failed: %v", err))
				continue
			}
			go func(conn net.Conn) {
				if err := c.Handle(conn, c.ui); err != nil {
					c.ui.Error(err.Error())
				}
			}(conn)
		}
	}
}

func (c *adapter) Handle(conn net.Conn, ui packer.Ui) error {
	c.ui.Message("SSH proxy: accepted connection")
	_, chans, reqs, err := ssh.NewServerConn(conn, c.config)
	if err != nil {
		return errors.New("failed to handshake")
	}

	// discard all global requests
	go ssh.DiscardRequests(reqs)

	// Service the incoming NewChannels
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		go func(ch ssh.NewChannel) {
			if err := c.handleSession(ch); err != nil {
				c.ui.Error(err.Error())
			}
		}(newChannel)
	}

	return nil
}

func (c *adapter) handleSession(newChannel ssh.NewChannel) error {
	channel, requests, err := newChannel.Accept()
	if err != nil {
		return err
	}
	defer channel.Close()

	done := make(chan struct{})

	// Sessions have requests such as "pty-req", "shell", "env", and "exec".
	// see RFC 4254, section 6
	go func(in <-chan *ssh.Request) {
		env := make([]envRequestPayload, 4)
		for req := range in {
			switch req.Type {
			case "pty-req":
				// accept pty-req requests, but don't actually do anything. Necessary for OpenSSH and sudo.
				req.Reply(true, nil)

			case "env":
				req.Reply(true, nil)

				req, err := newEnvRequest(req)
				if err != nil {
					c.ui.Error(err.Error())
					continue
				}
				env = append(env, req.Payload)
			case "exec":
				req.Reply(true, nil)

				req, err := newExecRequest(req)
				if err != nil {
					c.ui.Error(err.Error())
					close(done)
					continue
				}

				if len(req.Payload) > 0 {
					cmd := &packer.RemoteCmd{
						Stdin:   channel,
						Stdout:  channel,
						Stderr:  channel.Stderr(),
						Command: string(req.Payload),
					}

					if err := c.comm.Start(cmd); err != nil {
						c.ui.Error(err.Error())
						close(done)
						return
					}
					go func(cmd *packer.RemoteCmd, channel ssh.Channel) {
						cmd.Wait()

						exitStatus := make([]byte, 4)
						binary.BigEndian.PutUint32(exitStatus, uint32(cmd.ExitStatus))
						channel.SendRequest("exit-status", false, exitStatus)
						close(done)
					}(cmd, channel)
				}

			case "subsystem":
				req, err := newSubsystemRequest(req)
				if err != nil {
					c.ui.Error(err.Error())
					continue
				}

				switch req.Payload {
				case "sftp":
					c.ui.Say("starting sftp subsystem")
					req.Reply(true, nil)
					sftpCmd := c.sftpCmd
					if len(sftpCmd) == 0 {
						sftpCmd = "/usr/lib/sftp-server -e"
					}
					cmd := &packer.RemoteCmd{
						Stdin:   channel,
						Stdout:  channel,
						Stderr:  channel.Stderr(),
						Command: sftpCmd,
					}

					if err := c.comm.Start(cmd); err != nil {
						c.ui.Error(err.Error())
					}

					go func() {
						cmd.Wait()
						close(done)
					}()

				default:
					req.Reply(false, nil)

				}
			default:
				c.ui.Message(fmt.Sprintf("rejecting %s request", req.Type))
				req.Reply(false, nil)
			}
		}
	}(requests)

	<-done
	return nil
}

func (c *adapter) Shutdown() {
	c.l.Close()
}

type envRequest struct {
	*ssh.Request
	Payload envRequestPayload
}

type envRequestPayload struct {
	Name  string
	Value string
}

func newEnvRequest(raw *ssh.Request) (*envRequest, error) {
	r := new(envRequest)
	r.Request = raw

	if err := ssh.Unmarshal(raw.Payload, &r.Payload); err != nil {
		return nil, err
	}

	return r, nil
}

func sshString(buf io.Reader) (string, error) {
	var size uint32
	err := binary.Read(buf, binary.BigEndian, &size)
	if err != nil {
		return "", err
	}

	b := make([]byte, size)
	err = binary.Read(buf, binary.BigEndian, b)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

type execRequest struct {
	*ssh.Request
	Payload execRequestPayload
}

type execRequestPayload string

func newExecRequest(raw *ssh.Request) (*execRequest, error) {
	r := new(execRequest)
	r.Request = raw
	buf := bytes.NewReader(r.Request.Payload)

	var err error
	var payload string
	if payload, err = sshString(buf); err != nil {
		return nil, err
	}

	r.Payload = execRequestPayload(payload)
	return r, nil
}

type subsystemRequest struct {
	*ssh.Request
	Payload subsystemRequestPayload
}

type subsystemRequestPayload string

func newSubsystemRequest(raw *ssh.Request) (*subsystemRequest, error) {
	r := new(subsystemRequest)
	r.Request = raw
	buf := bytes.NewReader(r.Request.Payload)

	var err error
	var payload string
	if payload, err = sshString(buf); err != nil {
		return nil, err
	}

	r.Payload = subsystemRequestPayload(payload)
	return r, nil
}
