// Package ssh implements the SSH communicator. Plugin maintainers should not
// import this package directly, instead using the tooling in the
// "packer-plugin-sdk/communicator" module.
package ssh

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// ErrHandshakeTimeout is returned from New() whenever we're unable to establish
// an ssh connection within a certain timeframe. By default the handshake time-
// out period is 1 minute. You can change it with Config.HandshakeTimeout.
var ErrHandshakeTimeout = fmt.Errorf("Timeout during SSH handshake")

type comm struct {
	client  *ssh.Client
	config  *Config
	conn    net.Conn
	address string
}

// TunnelDirection is the supported tunnel directions
type TunnelDirection int

const (
	UnsetTunnel TunnelDirection = iota
	RemoteTunnel
	LocalTunnel
)

// TunnelSpec represents a request to map a port on one side of the SSH connection to the other
type TunnelSpec struct {
	Direction   TunnelDirection
	ListenType  string
	ListenAddr  string
	ForwardType string
	ForwardAddr string
}

// Config is the structure used to configure the SSH communicator.
type Config struct {
	// The configuration of the Go SSH connection
	SSHConfig *ssh.ClientConfig

	// Connection returns a new connection. The current connection
	// in use will be closed as part of the Close method, or in the
	// case an error occurs.
	Connection func() (net.Conn, error)

	// Pty, if true, will request a pty from the remote end.
	Pty bool

	// DisableAgentForwarding, if true, will not forward the SSH agent.
	DisableAgentForwarding bool

	// HandshakeTimeout limits the amount of time we'll wait to handshake before
	// saying the connection failed.
	HandshakeTimeout time.Duration

	// UseSftp, if true, sftp will be used instead of scp for file transfers
	UseSftp bool

	// KeepAliveInterval sets how often we send a channel request to the
	// server. A value < 0 disables.
	KeepAliveInterval time.Duration

	// Timeout is how long to wait for a read or write to succeed.
	Timeout time.Duration

	Tunnels []TunnelSpec
}

// Creates a new packersdk.Communicator implementation over SSH. This takes
// an already existing TCP connection and SSH configuration.
func New(address string, config *Config) (result *comm, err error) {
	// Establish an initial connection and connect
	result = &comm{
		config:  config,
		address: address,
	}

	if err = result.reconnect(); err != nil {
		result = nil
		return
	}

	return
}

func (c *comm) Start(ctx context.Context, cmd *packersdk.RemoteCmd) (err error) {
	session, err := c.newSession()
	if err != nil {
		return
	}

	// Setup our session
	session.Stdin = cmd.Stdin
	session.Stdout = cmd.Stdout
	session.Stderr = cmd.Stderr

	if c.config.Pty {
		// Request a PTY
		termModes := ssh.TerminalModes{
			ssh.ECHO:          0,     // do not echo
			ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
			ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
		}

		if err = session.RequestPty("xterm", 40, 80, termModes); err != nil {
			return
		}
	}

	log.Printf("[DEBUG] starting remote command: %s", cmd.Command)
	err = session.Start(cmd.Command + "\n")
	if err != nil {
		return
	}

	go func() {
		if c.config.KeepAliveInterval <= 0 {
			return
		}
		c := time.NewTicker(c.config.KeepAliveInterval)
		defer c.Stop()
		for range c.C {
			_, err := session.SendRequest("keepalive@packer.io", true, nil)
			if err != nil {
				return
			}
		}
	}()

	// Start a goroutine to wait for the session to end and set the
	// exit boolean and status.
	go func() {
		defer session.Close()

		err := session.Wait()
		exitStatus := 0
		if err != nil {
			switch err.(type) {
			case *ssh.ExitError:
				exitStatus = err.(*ssh.ExitError).ExitStatus()
				log.Printf("[ERROR] Remote command exited with '%d': %s", exitStatus, cmd.Command)
			case *ssh.ExitMissingError:
				log.Printf("[ERROR] Remote command exited without exit status or exit signal.")
				exitStatus = packersdk.CmdDisconnect
			default:
				log.Printf("[ERROR] Error occurred waiting for ssh session: %s", err.Error())
			}
		}
		cmd.SetExited(exitStatus)
	}()
	return
}

func (c *comm) Upload(path string, input io.Reader, fi *os.FileInfo) error {
	if c.config.UseSftp {
		return c.sftpUploadSession(path, input, fi)
	} else {
		return c.scpUploadSession(path, input, fi)
	}
}

func (c *comm) UploadDir(dst string, src string, excl []string) error {
	log.Printf("[DEBUG] Upload dir '%s' to '%s'", src, dst)
	if c.config.UseSftp {
		return c.sftpUploadDirSession(dst, src, excl)
	} else {
		return c.scpUploadDirSession(dst, src, excl)
	}
}

func (c *comm) DownloadDir(src string, dst string, excl []string) error {
	log.Printf("[DEBUG] Download dir '%s' to '%s'", src, dst)
	scpFunc := func(w io.Writer, stdoutR *bufio.Reader) error {
		dirStack := []string{dst}
		for {
			fmt.Fprint(w, "\x00")

			// read file info
			fi, err := stdoutR.ReadString('\n')
			if err != nil {
				return err
			}

			if len(fi) < 0 {
				return fmt.Errorf("empty response from server")
			}

			switch fi[0] {
			case '\x01', '\x02':
				return fmt.Errorf("%s", fi[1:])
			case 'C', 'D':
				break
			case 'E':
				dirStack = dirStack[:len(dirStack)-1]
				if len(dirStack) == 0 {
					fmt.Fprint(w, "\x00")
					return nil
				}
				continue
			default:
				return fmt.Errorf("unexpected server response (%x)", fi[0])
			}

			var mode int64
			var size int64
			var name string
			log.Printf("[DEBUG] Download dir str:%s", fi)
			n, err := fmt.Sscanf(fi[1:], "%o %d %s", &mode, &size, &name)
			if err != nil || n != 3 {
				return fmt.Errorf("can't parse server response (%s)", fi)
			}
			if size < 0 {
				return fmt.Errorf("negative file size")
			}

			log.Printf("[DEBUG] Download dir mode:%0o size:%d name:%s", mode, size, name)

			dst = filepath.Join(dirStack...)
			switch fi[0] {
			case 'D':
				err = os.MkdirAll(filepath.Join(dst, name), os.FileMode(mode))
				if err != nil {
					return err
				}
				dirStack = append(dirStack, name)
				continue
			case 'C':
				fmt.Fprint(w, "\x00")
				err = scpDownloadFile(filepath.Join(dst, name), stdoutR, size, os.FileMode(mode))
				if err != nil {
					return err
				}
			}

			if err := checkSCPStatus(stdoutR); err != nil {
				return err
			}
		}
	}
	return c.scpSession("scp -vrf "+src, scpFunc)
}

func (c *comm) Download(path string, output io.Writer) error {
	if c.config.UseSftp {
		return c.sftpDownloadSession(path, output)
	}
	return c.scpDownloadSession(path, output)
}

func (c *comm) newSession() (session *ssh.Session, err error) {
	log.Println("[DEBUG] Opening new ssh session")
	if c.client == nil {
		err = errors.New("client not available")
	} else {
		session, err = c.client.NewSession()
	}

	if err != nil {
		log.Printf("[ERROR] ssh session open error: '%s', attempting reconnect", err)
		if err := c.reconnect(); err != nil {
			return nil, err
		}

		if c.client == nil {
			return nil, errors.New("client not available")
		} else {
			return c.client.NewSession()
		}
	}

	return session, nil
}

func (c *comm) reconnect() (err error) {
	if c.conn != nil {
		// Ignore errors here because we don't care if it fails
		c.conn.Close()
	}

	// Set the conn and client to nil since we'll recreate it
	c.conn = nil
	c.client = nil

	log.Printf("[DEBUG] reconnecting to TCP connection for SSH")
	c.conn, err = c.config.Connection()
	if err != nil {
		// Explicitly set this to the REAL nil. Connection() can return
		// a nil implementation of net.Conn which will make the
		// "if c.conn == nil" check fail above. Read here for more information
		// on this psychotic language feature:
		//
		// http://golang.org/doc/faq#nil_error
		c.conn = nil

		log.Printf("[ERROR] reconnection error: %s", err)
		return
	}

	if c.config.Timeout > 0 {
		c.conn = &timeoutConn{c.conn, c.config.Timeout, c.config.Timeout}
	}

	log.Printf("[DEBUG] handshaking with SSH")

	// Default timeout to 1 minute if it wasn't specified (zero value). For
	// when you need to handshake from low orbit.
	var duration time.Duration
	if c.config.HandshakeTimeout == 0 {
		duration = 1 * time.Minute
	} else {
		duration = c.config.HandshakeTimeout
	}

	connectionEstablished := make(chan struct{}, 1)

	var sshConn ssh.Conn
	var sshChan <-chan ssh.NewChannel
	var req <-chan *ssh.Request

	go func() {
		sshConn, sshChan, req, err = ssh.NewClientConn(c.conn, c.address, c.config.SSHConfig)
		close(connectionEstablished)
	}()

	select {
	case <-connectionEstablished:
		// We don't need to do anything here. We just want select to block until
		// we connect or timeout.
	case <-time.After(duration):
		if c.conn != nil {
			c.conn.Close()
		}
		if sshConn != nil {
			sshConn.Close()
		}
		return ErrHandshakeTimeout
	}

	if err != nil {
		return
	}
	log.Printf("[DEBUG] handshake complete!")
	if sshConn != nil {
		c.client = ssh.NewClient(sshConn, sshChan, req)
	}
	c.connectToAgent()
	err = c.connectTunnels(sshConn)
	if err != nil {
		return
	}

	return
}

func (c *comm) connectTunnels(sshConn ssh.Conn) (err error) {
	if c.client == nil {
		return
	}

	if len(c.config.Tunnels) == 0 {
		// No Tunnels to configure
		return
	}

	// Start remote forwards of ports to ourselves.
	log.Printf("[DEBUG] Tunnel configuration: %v", c.config.Tunnels)
	for _, v := range c.config.Tunnels {
		done := make(chan struct{})
		var listener net.Listener
		switch v.Direction {
		case RemoteTunnel:
			// This requests the sshd Host to bind a port and send traffic back to us
			listener, err = c.client.Listen(v.ListenType, v.ListenAddr)
			if err != nil {
				err = fmt.Errorf("Tunnel: Failed to bind remote ('%v'): %s", v, err)
				return
			}
			log.Printf("[INFO] Tunnel: Remote bound on %s forwarding to %s", v.ListenAddr, v.ForwardAddr)
			connectFunc := ConnectFunc(v.ForwardType, v.ForwardAddr)
			go ProxyServe(listener, done, connectFunc)
			// Wait for our sshConn to be shutdown
			// FIXME: Is there a better "on-shutdown" we can wait on?
			go shutdownProxyTunnel(sshConn, done, listener)
		case LocalTunnel:
			// This binds locally and sends traffic back to the sshd host
			listener, err = net.Listen(v.ListenType, v.ListenAddr)
			if err != nil {
				err = fmt.Errorf("Tunnel: Failed to bind local ('%v'): %s", v, err)
				return
			}
			log.Printf("[INFO] Tunnel: Local bound on %s forwarding to %s", v.ListenAddr, v.ForwardAddr)
			connectFunc := func() (net.Conn, error) {
				// This Dial occurs on the SSH server's side
				return c.client.Dial(v.ForwardType, v.ForwardAddr)
			}
			go ProxyServe(listener, done, connectFunc)
			// FIXME: Is there a better "on-shutdown" we can wait on?
			go shutdownProxyTunnel(sshConn, done, listener)
		default:
			err = fmt.Errorf("Tunnel: Unknown tunnel direction ('%v'): %v", v, v.Direction)
			return
		}
	}

	return
}

// shutdownProxyTunnel waits for our sshConn to be shutdown and closes the listeners
func shutdownProxyTunnel(sshConn ssh.Conn, done chan struct{}, listener net.Listener) {
	sshConn.Wait()
	log.Printf("[INFO] Tunnel: Shutting down listener %v", listener)
	done <- struct{}{}
	close(done)
	listener.Close()
}

func (c *comm) connectToAgent() {
	if c.client == nil {
		return
	}

	if c.config.DisableAgentForwarding {
		log.Printf("[INFO] SSH agent forwarding is disabled.")
		return
	}

	// open connection to the local agent
	socketLocation := os.Getenv("SSH_AUTH_SOCK")
	if socketLocation == "" {
		log.Printf("[INFO] no local agent socket, will not connect agent")
		return
	}
	agentConn, err := net.Dial("unix", socketLocation)
	if err != nil {
		log.Printf("[ERROR] could not connect to local agent socket: %s", socketLocation)
		return
	}

	// create agent and add in auth
	forwardingAgent := agent.NewClient(agentConn)
	if forwardingAgent == nil {
		log.Printf("[ERROR] Could not create agent client")
		agentConn.Close()
		return
	}

	// add callback for forwarding agent to SSH config
	// XXX - might want to handle reconnects appending multiple callbacks
	auth := ssh.PublicKeysCallback(forwardingAgent.Signers)
	c.config.SSHConfig.Auth = append(c.config.SSHConfig.Auth, auth)
	agent.ForwardToAgent(c.client, forwardingAgent)

	// Setup a session to request agent forwarding
	session, err := c.newSession()
	if err != nil {
		return
	}
	defer session.Close()

	err = agent.RequestAgentForwarding(session)
	if err != nil {
		log.Printf("[ERROR] RequestAgentForwarding: %#v", err)
		return
	}

	log.Printf("[INFO] agent forwarding enabled")
	return
}

func (c *comm) sftpUploadSession(path string, input io.Reader, fi *os.FileInfo) error {
	sftpFunc := func(client *sftp.Client) error {
		return c.sftpUploadFile(path, input, client, fi)
	}

	return c.sftpSession(sftpFunc)
}

func (c *comm) sftpUploadFile(path string, input io.Reader, client *sftp.Client, fi *os.FileInfo) error {
	log.Printf("[DEBUG] sftp: uploading %s", path)
	f, err := client.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = io.Copy(f, input); err != nil {
		return err
	}

	if fi != nil && (*fi).Mode().IsRegular() {
		mode := (*fi).Mode().Perm()
		err = client.Chmod(path, mode)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *comm) sftpUploadDirSession(dst string, src string, excl []string) error {
	sftpFunc := func(client *sftp.Client) error {
		rootDst := dst
		if src[len(src)-1] != '/' {
			log.Printf("[DEBUG] No trailing slash, creating the source directory name")
			rootDst = filepath.Join(dst, filepath.Base(src))
		}
		walkFunc := func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// Calculate the final destination using the
			// base source and root destination
			relSrc, err := filepath.Rel(src, path)
			if err != nil {
				return err
			}
			finalDst := filepath.Join(rootDst, relSrc)

			// In Windows, Join uses backslashes which we don't want to get
			// to the sftp server
			finalDst = filepath.ToSlash(finalDst)

			// Skip the creation of the target destination directory since
			// it should exist and we might not even own it
			if finalDst == dst {
				return nil
			}

			return c.sftpVisitFile(finalDst, path, info, client)
		}

		return filepath.Walk(src, walkFunc)
	}

	return c.sftpSession(sftpFunc)
}

func (c *comm) sftpMkdir(path string, client *sftp.Client, fi os.FileInfo) error {
	log.Printf("[DEBUG] sftp: creating dir %s", path)

	if err := client.Mkdir(path); err != nil {
		// Do not consider it an error if the directory existed
		remoteFi, fiErr := client.Lstat(path)
		if fiErr != nil || !remoteFi.IsDir() {
			return err
		}
	}

	mode := fi.Mode().Perm()
	if err := client.Chmod(path, mode); err != nil {
		return err
	}
	return nil
}

func (c *comm) sftpVisitFile(dst string, src string, fi os.FileInfo, client *sftp.Client) error {
	if !fi.IsDir() {
		f, err := os.Open(src)
		if err != nil {
			return err
		}
		defer f.Close()
		return c.sftpUploadFile(dst, f, client, &fi)
	} else {
		err := c.sftpMkdir(dst, client, fi)
		return err
	}
}

func (c *comm) sftpDownloadSession(path string, output io.Writer) error {
	sftpFunc := func(client *sftp.Client) error {
		f, err := client.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err = io.Copy(output, f); err != nil {
			return err
		}

		return nil
	}

	return c.sftpSession(sftpFunc)
}

func (c *comm) sftpSession(f func(*sftp.Client) error) error {
	client, err := c.newSftpClient()
	if err != nil {
		return fmt.Errorf("sftpSession error: %s", err.Error())
	}
	defer client.Close()

	return f(client)
}

func (c *comm) newSftpClient() (*sftp.Client, error) {
	session, err := c.newSession()
	if err != nil {
		return nil, err
	}

	if err := session.RequestSubsystem("sftp"); err != nil {
		return nil, err
	}

	pw, err := session.StdinPipe()
	if err != nil {
		return nil, err
	}
	pr, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}

	// Capture stdout so we can return errors to the user
	var stdout bytes.Buffer
	tee := io.TeeReader(pr, &stdout)
	client, err := sftp.NewClientPipe(tee, pw)
	if err != nil && stdout.Len() > 0 {
		log.Printf("[ERROR] Upload failed: %s", stdout.Bytes())
	}

	return client, err
}

func (c *comm) scpUploadSession(path string, input io.Reader, fi *os.FileInfo) error {

	// The target directory and file for talking the SCP protocol
	target_dir := filepath.Dir(path)
	target_file := filepath.Base(path)

	// On windows, filepath.Dir uses backslash separators (ie. "\tmp").
	// This does not work when the target host is unix.  Switch to forward slash
	// which works for unix and windows
	target_dir = filepath.ToSlash(target_dir)

	// Escape spaces in remote directory
	target_dir = strings.Replace(target_dir, " ", "\\ ", -1)

	scpFunc := func(w io.Writer, stdoutR *bufio.Reader) error {
		return scpUploadFile(target_file, input, w, stdoutR, fi)
	}

	return c.scpSession("scp -vt "+target_dir, scpFunc)
}

func (c *comm) scpUploadDirSession(dst string, src string, excl []string) error {
	scpFunc := func(w io.Writer, r *bufio.Reader) error {
		uploadEntries := func() error {
			f, err := os.Open(src)
			if err != nil {
				return err
			}
			defer f.Close()

			entries, err := f.Readdir(-1)
			if err != nil {
				return err
			}

			return scpUploadDir(src, entries, w, r)
		}

		if src[len(src)-1] != '/' {
			log.Printf("[DEBUG] No trailing slash, creating the source directory name")
			fi, err := os.Stat(src)
			if err != nil {
				return err
			}
			return scpUploadDirProtocol(filepath.Base(src), w, r, uploadEntries, fi)
		} else {
			// Trailing slash, so only upload the contents
			return uploadEntries()
		}
	}

	return c.scpSession("scp -rvt "+dst, scpFunc)
}

func (c *comm) scpDownloadSession(path string, output io.Writer) error {
	scpFunc := func(w io.Writer, stdoutR *bufio.Reader) error {
		fmt.Fprint(w, "\x00")

		// read file info
		fi, err := stdoutR.ReadString('\n')
		if err != nil {
			return err
		}

		if len(fi) < 0 {
			return fmt.Errorf("empty response from server")
		}

		switch fi[0] {
		case '\x01', '\x02':
			return fmt.Errorf("%s", fi[1:])
		case 'C':
		case 'D':
			return fmt.Errorf("remote file is directory")
		default:
			return fmt.Errorf("unexpected server response (%x)", fi[0])
		}

		var mode string
		var size int64

		n, err := fmt.Sscanf(fi, "%6s %d ", &mode, &size)
		if err != nil || n != 2 {
			return fmt.Errorf("can't parse server response (%s)", fi)
		}
		if size < 0 {
			return fmt.Errorf("negative file size")
		}

		fmt.Fprint(w, "\x00")

		if _, err := io.CopyN(output, stdoutR, size); err != nil {
			return err
		}

		fmt.Fprint(w, "\x00")

		return checkSCPStatus(stdoutR)
	}

	if !strings.Contains(path, " ") {
		return c.scpSession("scp -vf "+path, scpFunc)
	}
	return c.scpSession("scp -vf "+strconv.Quote(path), scpFunc)
}

func (c *comm) scpSession(scpCommand string, f func(io.Writer, *bufio.Reader) error) error {
	session, err := c.newSession()
	if err != nil {
		return err
	}
	defer session.Close()

	// Get a pipe to stdin so that we can send data down
	stdinW, err := session.StdinPipe()
	if err != nil {
		return err
	}

	// We only want to close once, so we nil w after we close it,
	// and only close in the defer if it hasn't been closed already.
	defer func() {
		if stdinW != nil {
			stdinW.Close()
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

	// Start the sink mode on the other side
	// TODO(mitchellh): There are probably issues with shell escaping the path
	log.Println("[DEBUG] Starting remote scp process: ", scpCommand)
	if err := session.Start(scpCommand); err != nil {
		return err
	}

	// Call our callback that executes in the context of SCP. We ignore
	// EOF errors if they occur because it usually means that SCP prematurely
	// ended on the other side.
	log.Println("[DEBUG] Started SCP session, beginning transfers...")
	if err := f(stdinW, stdoutR); err != nil && err != io.EOF {
		return err
	}

	// Close the stdin, which sends an EOF, and then set w to nil so that
	// our defer func doesn't close it again since that is unsafe with
	// the Go SSH package.
	log.Println("[DEBUG] SCP session complete, closing stdin pipe.")
	stdinW.Close()
	stdinW = nil

	// Wait for the SCP connection to close, meaning it has consumed all
	// our data and has completed. Or has errored.
	log.Println("[DEBUG] Waiting for SSH session to complete.")
	err = session.Wait()
	log.Printf("[DEBUG] scp stderr (length %d): %s", stderr.Len(), stderr.String())
	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			// Otherwise, we have an ExitError, meaning we can just read the
			// exit status
			log.Printf("[DEBUG] non-zero exit status: %d, %v", exitErr.ExitStatus(), err)
			stdoutB, err := ioutil.ReadAll(stdoutR)
			if err != nil {
				return err
			}
			log.Printf("[DEBUG] scp output: %s", stdoutB)

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

	return nil
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

func scpDownloadFile(dst string, src io.Reader, size int64, mode os.FileMode) error {
	f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := io.CopyN(f, src, size); err != nil {
		return err
	}
	return nil
}

func scpUploadFile(dst string, src io.Reader, w io.Writer, r *bufio.Reader, fi *os.FileInfo) error {
	var mode os.FileMode
	var size int64

	if fi != nil && (*fi).Mode().IsRegular() {
		mode = (*fi).Mode().Perm()
		size = (*fi).Size()
	} else {
		// Create a temporary file where we can copy the contents of the src
		// so that we can determine the length, since SCP is length-prefixed.
		tf, err := tmp.File("packer-upload")
		if err != nil {
			return fmt.Errorf("Error creating temporary file for upload: %s", err)
		}
		defer os.Remove(tf.Name())
		defer tf.Close()

		mode = 0644

		log.Println("[DEBUG] Copying input data into temporary file so we can read the length")
		if _, err := io.Copy(tf, src); err != nil {
			return fmt.Errorf("Error copying input data into local temporary "+
				"file. Check that TEMPDIR has enough space. Please see "+
				"https://www.packer.io/docs/other/environment-variables#tmpdir"+
				"for more info. Error: %s", err)
		}

		// Sync the file so that the contents are definitely on disk, then
		// read the length of it.
		if err := tf.Sync(); err != nil {
			return fmt.Errorf("Error creating temporary file for upload: %s", err)
		}

		// Seek the file to the beginning so we can re-read all of it
		if _, err := tf.Seek(0, 0); err != nil {
			return fmt.Errorf("Error creating temporary file for upload: %s", err)
		}

		tfi, err := tf.Stat()
		if err != nil {
			return fmt.Errorf("Error creating temporary file for upload: %s", err)
		}

		size = tfi.Size()
		src = tf
	}

	// Start the protocol
	perms := fmt.Sprintf("C%04o", mode)
	log.Printf("[DEBUG] scp: Uploading %s: perms=%s size=%d", dst, perms, size)

	fmt.Fprintln(w, perms, size, dst)
	if err := checkSCPStatus(r); err != nil {
		return err
	}

	if _, err := io.CopyN(w, src, size); err != nil {
		return err
	}

	fmt.Fprint(w, "\x00")
	return checkSCPStatus(r)
}

func scpUploadDirProtocol(name string, w io.Writer, r *bufio.Reader, f func() error, fi os.FileInfo) error {
	log.Printf("[DEBUG] SCP: starting directory upload: %s", name)

	mode := fi.Mode().Perm()

	perms := fmt.Sprintf("D%04o 0", mode)

	fmt.Fprintln(w, perms, name)
	err := checkSCPStatus(r)
	if err != nil {
		return err
	}

	if err := f(); err != nil {
		return err
	}

	fmt.Fprintln(w, "E")
	return err
}

func scpUploadDir(root string, fs []os.FileInfo, w io.Writer, r *bufio.Reader) error {
	for _, fi := range fs {
		realPath := filepath.Join(root, fi.Name())

		// Track if this is actually a symlink to a directory. If it is
		// a symlink to a file we don't do any special behavior because uploading
		// a file just works. If it is a directory, we need to know so we
		// treat it as such.
		isSymlinkToDir := false
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			symPath, err := filepath.EvalSymlinks(realPath)
			if err != nil {
				return err
			}

			symFi, err := os.Lstat(symPath)
			if err != nil {
				return err
			}

			isSymlinkToDir = symFi.IsDir()
		}

		if !fi.IsDir() && !isSymlinkToDir {
			// It is a regular file (or symlink to a file), just upload it
			f, err := os.Open(realPath)
			if err != nil {
				return err
			}

			err = func() error {
				defer f.Close()
				return scpUploadFile(fi.Name(), f, w, r, &fi)
			}()

			if err != nil {
				return err
			}

			continue
		}

		// It is a directory, recursively upload
		err := scpUploadDirProtocol(fi.Name(), w, r, func() error {
			f, err := os.Open(realPath)
			if err != nil {
				return err
			}
			defer f.Close()

			entries, err := f.Readdir(-1)
			if err != nil {
				return err
			}

			return scpUploadDir(realPath, entries, w, r)
		}, fi)
		if err != nil {
			return err
		}
	}

	return nil
}
