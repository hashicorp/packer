package ssh

import (
	"bufio"
	"bytes"
	"code.google.com/p/go.crypto/ssh"
	"errors"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"io"
	"log"
	"net"
	"path/filepath"
)

type comm struct {
	client *ssh.ClientConn
	config *Config
	conn   net.Conn
}

// Config is the structure used to configure the SSH communicator.
type Config struct {
	// The configuration of the Go SSH connection
	SSHConfig *ssh.ClientConfig

	// Connection returns a new connection. The current connection
	// in use will be closed as part of the Close method, or in the
	// case an error occurs.
	Connection func() (net.Conn, error)
}

// Creates a new packer.Communicator implementation over SSH. This takes
// an already existing TCP connection and SSH configuration.
func New(config *Config) (result *comm, err error) {
	// Establish an initial connection and connect
	result = &comm{
		config: config,
	}

	if err = result.reconnect(); err != nil {
		result = nil
		return
	}

	return
}

func (c *comm) Start(cmd *packer.RemoteCmd) (err error) {
	session, err := c.newSession()
	if err != nil {
		return
	}

	// Setup our session
	session.Stdin = cmd.Stdin
	session.Stdout = cmd.Stdout
	session.Stderr = cmd.Stderr

	// Request a PTY
	termModes := ssh.TerminalModes{
		ssh.ECHO:          0,     // do not echo
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err = session.RequestPty("xterm", 80, 40, termModes); err != nil {
		return
	}

	log.Printf("starting remote command: %s", cmd.Command)
	err = session.Start(cmd.Command + "\n")
	if err != nil {
		return
	}

	// Start a goroutine to wait for the session to end and set the
	// exit boolean and status.
	go func() {
		defer session.Close()
		err := session.Wait()
		exitStatus := 0
		if err != nil {
			exitErr, ok := err.(*ssh.ExitError)
			if ok {
				exitStatus = exitErr.ExitStatus()
			}
		}

		log.Printf("remote command exited with '%d': %s", exitStatus, cmd.Command)
		cmd.SetExited(exitStatus)
	}()

	return
}

func (c *comm) Upload(path string, input io.Reader) error {
	session, err := c.newSession()
	if err != nil {
		return err
	}

	defer session.Close()

	// Get a pipe to stdin so that we can send data down
	w, err := session.StdinPipe()
	if err != nil {
		return err
	}

	// We only want to close once, so we nil w after we close it,
	// and only close in the defer if it hasn't been closed already.
	defer func() {
		if w != nil {
			w.Close()
		}
	}()

	// Get a pipe to stdout so that we can get responses back
	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		return err
	}
	stdoutR := bufio.NewReader(stdoutPipe)

	// Set stderr to a bytes buffer
	stderr := new(bytes.Buffer)
	session.Stderr = stderr

	// The target directory and file for talking the SCP protocol
	target_dir := filepath.Dir(path)
	target_file := filepath.Base(path)

	// On windows, filepath.Dir uses backslash seperators (ie. "\tmp").
	// This does not work when the target host is unix.  Switch to forward slash
	// which works for unix and windows
	target_dir = filepath.ToSlash(target_dir)

	// Start the sink mode on the other side
	// TODO(mitchellh): There are probably issues with shell escaping the path
	log.Println("Starting remote scp process in sink mode")
	if err = session.Start("scp -vt " + target_dir); err != nil {
		return err
	}

	// Determine the length of the upload content by copying it
	// into an in-memory buffer. Note that this means what we upload
	// must fit into memory.
	log.Println("Copying input data into in-memory buffer so we can get the length")
	input_memory := new(bytes.Buffer)
	if _, err = io.Copy(input_memory, input); err != nil {
		return err
	}

	// Start the protocol
	log.Println("Beginning file upload...")
	fmt.Fprintln(w, "C0644", input_memory.Len(), target_file)
	err = checkSCPStatus(stdoutR)
	if err != nil {
		return err
	}

	if _, err := io.Copy(w, input_memory); err != nil {
		return err
	}

	fmt.Fprint(w, "\x00")
	err = checkSCPStatus(stdoutR)
	if err != nil {
		return err
	}

	// Close the stdin, which sends an EOF, and then set w to nil so that
	// our defer func doesn't close it again since that is unsafe with
	// the Go SSH package.
	log.Println("Upload complete, closing stdin pipe")
	w.Close()
	w = nil

	// Wait for the SCP connection to close, meaning it has consumed all
	// our data and has completed. Or has errored.
	log.Println("Waiting for SSH session to complete")
	err = session.Wait()
	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			// Otherwise, we have an ExitErorr, meaning we can just read
			// the exit status
			log.Printf("non-zero exit status: %d", exitErr.ExitStatus())

			// If we exited with status 127, it means SCP isn't available.
			// Return a more descriptive error for that.
			if exitErr.ExitStatus() == 127 {
				return errors.New(
					"SCP failed to start. This usually means that SCP is not\n" +
						"properly installed on the remote system.")
			}
		}

		return err
	}

	log.Printf("scp stderr (length %d): %s", stderr.Len(), stderr.String())

	return nil
}

func (c *comm) Download(string, io.Writer) error {
	panic("not implemented yet")
}

func (c *comm) newSession() (session *ssh.Session, err error) {
	log.Println("opening new ssh session")
	if c.client == nil {
		err = errors.New("client not available")
	} else {
		session, err = c.client.NewSession()
	}

	if err != nil {
		log.Printf("ssh session open error: '%s', attempting reconnect", err)
		if err := c.reconnect(); err != nil {
			return nil, err
		}

		return c.client.NewSession()
	}

	return session, nil
}

func (c *comm) reconnect() (err error) {
	if c.conn != nil {
		c.conn.Close()
	}

	// Set the conn and client to nil since we'll recreate it
	c.conn = nil
	c.client = nil

	log.Printf("reconnecting to TCP connection for SSH")
	c.conn, err = c.config.Connection()
	if err != nil {
		log.Printf("reconnection error: %s", err)
		return
	}

	log.Printf("handshaking with SSH")
	c.client, err = ssh.Client(c.conn, c.config.SSHConfig)
	if err != nil {
		log.Printf("handshake error: %s", err)
	}

	return
}

// checkSCPStatus checks that a prior command sent to SCP completed
// successfully. If it did not complete successfully, an error will
// be returned.
func checkSCPStatus(r *bufio.Reader) error {
	code, err := r.ReadByte()
	if err != nil {
		return err
	}

	if code != 0 {
		// Treat any non-zero (really 1 and 2) as fatal errors
		message, _, err := r.ReadLine()
		if err != nil {
			return fmt.Errorf("Error reading error message: %s", err)
		}

		return errors.New(string(message))
	}

	return nil
}
