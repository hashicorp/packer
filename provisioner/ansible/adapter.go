package ansible

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/google/shlex"
	"github.com/hashicorp/packer/packer"
	"golang.org/x/crypto/ssh"
)

// An adapter satisfies SSH requests (from an Ansible client) by delegating SSH
// exec and subsystem commands to a packer.Communicator.
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
	log.Printf("SSH proxy: serving on %s", c.l.Addr())

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
	log.Print("SSH proxy: accepted connection")
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
				log.Println("ansible provisioner pty-req request")
				// accept pty-req requests, but don't actually do anything. Necessary for OpenSSH and sudo.
				req.Reply(true, nil)

			case "env":
				req, err := newEnvRequest(req)
				if err != nil {
					c.ui.Error(err.Error())
					req.Reply(false, nil)
					continue
				}
				env = append(env, req.Payload)
				log.Printf("new env request: %s", req.Payload)
				req.Reply(true, nil)
			case "exec":
				req, err := newExecRequest(req)
				if err != nil {
					c.ui.Error(err.Error())
					req.Reply(false, nil)
					close(done)
					continue
				}

				log.Printf("new exec request: %s", req.Payload)

				if len(req.Payload) == 0 {
					req.Reply(false, nil)
					close(done)
					return
				}

				go func(channel ssh.Channel) {
					exit := c.exec(string(req.Payload), channel, channel, channel.Stderr())

					exitStatus := make([]byte, 4)
					binary.BigEndian.PutUint32(exitStatus, uint32(exit))
					channel.SendRequest("exit-status", false, exitStatus)
					close(done)
				}(channel)
				req.Reply(true, nil)
			case "subsystem":
				req, err := newSubsystemRequest(req)
				if err != nil {
					c.ui.Error(err.Error())
					req.Reply(false, nil)
					continue
				}

				log.Printf("new subsystem request: %s", req.Payload)
				switch req.Payload {
				case "sftp":
					sftpCmd := c.sftpCmd
					if len(sftpCmd) == 0 {
						sftpCmd = "/usr/lib/sftp-server -e"
					}

					log.Print("starting sftp subsystem")
					go func() {
						_ = c.remoteExec(sftpCmd, channel, channel, channel.Stderr())
						close(done)
					}()
					req.Reply(true, nil)
				default:
					c.ui.Error(fmt.Sprintf("unsupported subsystem requested: %s", req.Payload))
					req.Reply(false, nil)
				}
			default:
				log.Printf("rejecting %s request", req.Type)
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

func (c *adapter) exec(command string, in io.Reader, out io.Writer, err io.Writer) int {
	var exitStatus int
	switch {
	case strings.HasPrefix(command, "scp ") && serveSCP(command[4:]):
		err := c.scpExec(command[4:], in, out)
		if err != nil {
			log.Println(err)
			exitStatus = 1
		}
	default:
		exitStatus = c.remoteExec(command, in, out, err)
	}
	return exitStatus
}

func serveSCP(args string) bool {
	opts, _ := scpOptions(args)
	return bytes.IndexAny(opts, "tf") >= 0
}

func (c *adapter) scpExec(args string, in io.Reader, out io.Writer) error {
	opts, rest := scpOptions(args)

	// remove the quoting that ansible added to rest for shell safety.
	shargs, err := shlex.Split(rest)
	if err != nil {
		return err
	}
	rest = strings.Join(shargs, "")

	if i := bytes.IndexByte(opts, 't'); i >= 0 {
		return scpUploadSession(opts, rest, in, out, c.comm)
	}

	if i := bytes.IndexByte(opts, 'f'); i >= 0 {
		return scpDownloadSession(opts, rest, in, out, c.comm)
	}
	return errors.New("no scp mode specified")
}

func (c *adapter) remoteExec(command string, in io.Reader, out io.Writer, err io.Writer) int {
	cmd := &packer.RemoteCmd{
		Stdin:   in,
		Stdout:  out,
		Stderr:  err,
		Command: command,
	}

	if err := c.comm.Start(cmd); err != nil {
		c.ui.Error(err.Error())
		return cmd.ExitStatus
	}

	cmd.Wait()

	return cmd.ExitStatus
}

type envRequest struct {
	*ssh.Request
	Payload envRequestPayload
}

type envRequestPayload struct {
	Name  string
	Value string
}

func (p envRequestPayload) String() string {
	return fmt.Sprintf("%s=%s", p.Name, p.Value)
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

func (p execRequestPayload) String() string {
	return string(p)
}

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

func (p subsystemRequestPayload) String() string {
	return string(p)
}

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
