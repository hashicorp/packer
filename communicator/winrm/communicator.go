package winrm

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/masterzen/winrm/winrm"
	"github.com/mitchellh/packer/packer"
	"github.com/packer-community/winrmcp/winrmcp"

	// This import is a bit strange, but it's needed so `make updatedeps`
	// can see and download it
	_ "github.com/dylanmei/winrmtest"
)

// Communicator represents the WinRM communicator
type Communicator struct {
	config   *Config
	client   *winrm.Client
	endpoint *winrm.Endpoint
}

// New creates a new communicator implementation over WinRM.
func New(config *Config) (*Communicator, error) {
	endpoint := &winrm.Endpoint{
		Host:     config.Host,
		Port:     config.Port,
		HTTPS:    config.Https,
		Insecure: config.Insecure,

		/*
			TODO
			HTTPS:    connInfo.HTTPS,
			Insecure: connInfo.Insecure,
			CACert:   connInfo.CACert,
		*/
	}

	// Create the client
	params := winrm.DefaultParameters()

	if config.TransportDecorator != nil {
		params.TransportDecorator = config.TransportDecorator
	}

	params.Timeout = formatDuration(config.Timeout)
	client, err := winrm.NewClientWithParameters(
		endpoint, config.Username, config.Password, params)
	if err != nil {
		return nil, err
	}

	// Create the shell to verify the connection
	log.Printf("[DEBUG] connecting to remote shell using WinRM")
	shell, err := client.CreateShell()
	if err != nil {
		log.Printf("[ERROR] connection error: %s", err)
		return nil, err
	}

	if err := shell.Close(); err != nil {
		log.Printf("[ERROR] error closing connection: %s", err)
		return nil, err
	}

	return &Communicator{
		config:   config,
		client:   client,
		endpoint: endpoint,
	}, nil
}

// Start implementation of communicator.Communicator interface
func (c *Communicator) Start(rc *packer.RemoteCmd) error {
	shell, err := c.client.CreateShell()
	if err != nil {
		return err
	}

	log.Printf("[INFO] starting remote command: %s", rc.Command)
	cmd, err := shell.Execute(rc.Command)
	if err != nil {
		return err
	}

	go runCommand(shell, cmd, rc)
	return nil
}

func runCommand(shell *winrm.Shell, cmd *winrm.Command, rc *packer.RemoteCmd) {
	defer shell.Close()
	var wg sync.WaitGroup

	copyFunc := func(w io.Writer, r io.Reader) {
		defer wg.Done()
		io.Copy(w, r)
	}

	if rc.Stdout != nil && cmd.Stdout != nil {
		wg.Add(1)
		go copyFunc(rc.Stdout, cmd.Stdout)
	} else {
		log.Printf("[WARN] Failed to read stdout for command '%s'", rc.Command)
	}

	if rc.Stderr != nil && cmd.Stderr != nil {
		wg.Add(1)
		go copyFunc(rc.Stderr, cmd.Stderr)
	} else {
		log.Printf("[WARN] Failed to read stderr for command '%s'", rc.Command)
	}

	cmd.Wait()
	wg.Wait()

	code := cmd.ExitCode()
	log.Printf("[INFO] command '%s' exited with code: %d", rc.Command, code)
	rc.SetExited(code)
}

// Upload implementation of communicator.Communicator interface
func (c *Communicator) Upload(path string, input io.Reader, _ *os.FileInfo) error {
	wcp, err := c.newCopyClient()
	if err != nil {
		return err
	}
	log.Printf("Uploading file to '%s'", path)
	return wcp.Write(path, input)
}

// UploadDir implementation of communicator.Communicator interface
func (c *Communicator) UploadDir(dst string, src string, exclude []string) error {
	log.Printf("Uploading dir '%s' to '%s'", src, dst)
	wcp, err := c.newCopyClient()
	if err != nil {
		return err
	}
	return wcp.Copy(src, dst)
}

func (c *Communicator) Download(src string, dst io.Writer) error {
	return fmt.Errorf("WinRM doesn't support download.")
}

func (c *Communicator) DownloadDir(src string, dst string, exclude []string) error {
	return fmt.Errorf("WinRM doesn't support download dir.")
}

func (c *Communicator) newCopyClient() (*winrmcp.Winrmcp, error) {
	addr := fmt.Sprintf("%s:%d", c.endpoint.Host, c.endpoint.Port)
	return winrmcp.New(addr, &winrmcp.Config{
		Auth: winrmcp.Auth{
			User:     c.config.Username,
			Password: c.config.Password,
		},
		Https:                 c.config.Https,
		Insecure:              c.config.Insecure,
		OperationTimeout:      c.config.Timeout,
		MaxOperationsPerShell: 15, // lowest common denominator
		TransportDecorator:    c.config.TransportDecorator,
	})
}
