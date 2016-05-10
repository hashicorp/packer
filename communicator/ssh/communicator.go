package ssh

import (
	"bufio"
	"bytes"
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
	"sync"
	"time"

	"github.com/mitchellh/packer/packer"
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

	// DisableAgent, if true, will not forward the SSH agent.
	DisableAgent bool

	// HandshakeTimeout limits the amount of time we'll wait to handshake before
	// saying the connection failed.
	HandshakeTimeout time.Duration

	// UseSftp, if true, sftp will be used instead of scp for file transfers
	UseSftp bool
}

// Creates a new packer.Communicator implementation over SSH. This takes
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

func (c *comm) Start(cmd *packer.RemoteCmd) (err error) {
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

	log.Printf("starting remote command: %s", cmd.Command)
	err = session.Start(cmd.Command + "\n")
	if err != nil {
		return
	}

	// A channel to keep track of our done state
	doneCh := make(chan struct{})
	sessionLock := new(sync.Mutex)
	timedOut := false

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

		sessionLock.Lock()
		defer sessionLock.Unlock()

		if timedOut {
			// We timed out, so set the exit status to -1
			exitStatus = -1
		}

		log.Printf("remote command exited with '%d': %s", exitStatus, cmd.Command)
		cmd.SetExited(exitStatus)
		close(doneCh)
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
	log.Printf("Upload dir '%s' to '%s'", src, dst)
	if c.config.UseSftp {
		return c.sftpUploadDirSession(dst, src, excl)
	} else {
		return c.scpUploadDirSession(dst, src, excl)
	}
}

func (c *comm) DownloadDir(src string, dst string, excl []string) error {
	log.Printf("Download dir '%s' to '%s'", src, dst)
	scpFunc := func(w io.Writer, stdoutR *bufio.Reader) error {
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
				return fmt.Errorf("%s", fi[1:len(fi)])
			case 'C', 'D':
				break
			default:
				return fmt.Errorf("unexpected server response (%x)", fi[0])
			}

			var mode string
			var size int64
			var name string
			log.Printf("Download dir str:%s", fi)
			n, err := fmt.Sscanf(fi, "%6s %d %s", &mode, &size, &name)
			if err != nil || n != 3 {
				return fmt.Errorf("can't parse server response (%s)", fi)
			}
			if size < 0 {
				return fmt.Errorf("negative file size")
			}

			log.Printf("Download dir mode:%s size:%d name:%s", mode, size, name)
			switch fi[0] {
			case 'D':
				err = os.MkdirAll(filepath.Join(dst, name), os.FileMode(0755))
				if err != nil {
					return err
				}
				fmt.Fprint(w, "\x00")
				return nil
			case 'C':
				fmt.Fprint(w, "\x00")
				err = scpDownloadFile(filepath.Join(dst, name), stdoutR, size, os.FileMode(0644))
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
		// Explicitly set this to the REAL nil. Connection() can return
		// a nil implementation of net.Conn which will make the
		// "if c.conn == nil" check fail above. Read here for more information
		// on this psychotic language feature:
		//
		// http://golang.org/doc/faq#nil_error
		c.conn = nil

		log.Printf("reconnection error: %s", err)
		return
	}

	log.Printf("handshaking with SSH")

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
		log.Printf("handshake error: %s", err)
		return
	}
	log.Printf("handshake complete!")
	if sshConn != nil {
		c.client = ssh.NewClient(sshConn, sshChan, req)
	}
	c.connectToAgent()

	return
}

func (c *comm) connectToAgent() {
	if c.client == nil {
		return
	}

	if c.config.DisableAgent {
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
		return sftpUploadFile(path, input, client, fi)
	}

	return c.sftpSession(sftpFunc)
}

func sftpUploadFile(path string, input io.Reader, client *sftp.Client, fi *os.FileInfo) error {
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
			log.Printf("No trailing slash, creating the source directory name")
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

			return sftpVisitFile(finalDst, path, info, client)
		}

		return filepath.Walk(src, walkFunc)
	}

	return c.sftpSession(sftpFunc)
}

func sftpMkdir(path string, client *sftp.Client, fi os.FileInfo) error {
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

func sftpVisitFile(dst string, src string, fi os.FileInfo, client *sftp.Client) error {
	if !fi.IsDir() {
		f, err := os.Open(src)
		if err != nil {
			return err
		}
		defer f.Close()
		return sftpUploadFile(dst, f, client, &fi)
	} else {
		err := sftpMkdir(dst, client, fi)
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
		return err
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

	return sftp.NewClientPipe(pr, pw)
}

func (c *comm) scpUploadSession(path string, input io.Reader, fi *os.FileInfo) error {

	// The target directory and file for talking the SCP protocol
	target_dir := filepath.Dir(path)
	target_file := filepath.Base(path)

	// On windows, filepath.Dir uses backslash seperators (ie. "\tmp").
	// This does not work when the target host is unix.  Switch to forward slash
	// which works for unix and windows
	target_dir = filepath.ToSlash(target_dir)

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
			log.Printf("No trailing slash, creating the source directory name")
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
			return fmt.Errorf("%s", fi[1:len(fi)])
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

		if err := checkSCPStatus(stdoutR); err != nil {
			return err
		}

		return nil
	}

	if strings.Index(path, " ") == -1 {
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
	log.Println("Starting remote scp process: ", scpCommand)
	if err := session.Start(scpCommand); err != nil {
		return err
	}

	// Call our callback that executes in the context of SCP. We ignore
	// EOF errors if they occur because it usually means that SCP prematurely
	// ended on the other side.
	log.Println("Started SCP session, beginning transfers...")
	if err := f(stdinW, stdoutR); err != nil && err != io.EOF {
		return err
	}

	// Close the stdin, which sends an EOF, and then set w to nil so that
	// our defer func doesn't close it again since that is unsafe with
	// the Go SSH package.
	log.Println("SCP session complete, closing stdin pipe.")
	stdinW.Close()
	stdinW = nil

	// Wait for the SCP connection to close, meaning it has consumed all
	// our data and has completed. Or has errored.
	log.Println("Waiting for SSH session to complete.")
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
		tf, err := ioutil.TempFile("", "packer-upload")
		if err != nil {
			return fmt.Errorf("Error creating temporary file for upload: %s", err)
		}
		defer os.Remove(tf.Name())
		defer tf.Close()

		mode = 0644

		log.Println("Copying input data into temporary file so we can read the length")
		if _, err := io.Copy(tf, src); err != nil {
			return err
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
	if err := checkSCPStatus(r); err != nil {
		return err
	}

	return nil
}

func scpUploadDirProtocol(name string, w io.Writer, r *bufio.Reader, f func() error, fi os.FileInfo) error {
	log.Printf("SCP: starting directory upload: %s", name)

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
	if err != nil {
		return err
	}

	return nil
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
