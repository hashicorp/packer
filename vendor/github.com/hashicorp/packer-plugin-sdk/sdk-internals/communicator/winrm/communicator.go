// Package winrm implements the WinRM communicator. Plugin maintainers should not
// import this package directly, instead using the tooling in the
// "packer-plugin-sdk/communicator" module.
package winrm

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/masterzen/winrm"
	"github.com/packer-community/winrmcp/winrmcp"
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
	params := *winrm.DefaultParameters

	if config.TransportDecorator != nil {
		params.TransportDecorator = config.TransportDecorator
	}

	params.Timeout = formatDuration(config.Timeout)
	client, err := winrm.NewClientWithParameters(
		endpoint, config.Username, config.Password, &params)
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
func (c *Communicator) Start(ctx context.Context, rc *packersdk.RemoteCmd) error {
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

func runCommand(shell *winrm.Shell, cmd *winrm.Command, rc *packersdk.RemoteCmd) {
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
func (c *Communicator) Upload(path string, input io.Reader, fi *os.FileInfo) error {
	wcp, err := c.newCopyClient()
	if err != nil {
		return fmt.Errorf("Was unable to create winrm client: %s", err)
	}
	if strings.HasSuffix(path, `\`) {
		// path is a directory
		if fi != nil {
			path += filepath.Base((*fi).Name())
		} else {
			return fmt.Errorf("Was unable to infer file basename for upload.")
		}
	}
	log.Printf("Uploading file to '%s'", path)
	return wcp.Write(path, input)
}

// UploadDir implementation of communicator.Communicator interface
func (c *Communicator) UploadDir(dst string, src string, exclude []string) error {
	if !strings.HasSuffix(src, "/") {
		dst = fmt.Sprintf("%s\\%s", dst, filepath.Base(src))
	}
	log.Printf("Uploading dir '%s' to '%s'", src, dst)
	wcp, err := c.newCopyClient()
	if err != nil {
		return err
	}
	return wcp.Copy(src, dst)
}

func (c *Communicator) Download(src string, dst io.Writer) error {
	client, err := c.newWinRMClient()
	if err != nil {
		return err
	}

	encodeScript := `$file=[System.IO.File]::ReadAllBytes("%s"); Write-Output $([System.Convert]::ToBase64String($file))`

	base64DecodePipe := &Base64Pipe{w: dst}

	cmd := winrm.Powershell(fmt.Sprintf(encodeScript, src))
	_, err = client.Run(cmd, base64DecodePipe, ioutil.Discard)

	return err
}

func (c *Communicator) DownloadDir(src string, dst string, exclude []string) error {
	return fmt.Errorf("WinRM doesn't support download dir.")
}

func (c *Communicator) getClientConfig() *winrmcp.Config {
	return &winrmcp.Config{
		Auth: winrmcp.Auth{
			User:     c.config.Username,
			Password: c.config.Password,
		},
		Https:                 c.config.Https,
		Insecure:              c.config.Insecure,
		OperationTimeout:      c.config.Timeout,
		MaxOperationsPerShell: 15, // lowest common denominator
		TransportDecorator:    c.config.TransportDecorator,
	}
}

func (c *Communicator) newCopyClient() (*winrmcp.Winrmcp, error) {
	addr := fmt.Sprintf("%s:%d", c.endpoint.Host, c.endpoint.Port)
	clientConfig := c.getClientConfig()
	return winrmcp.New(addr, clientConfig)
}

func (c *Communicator) newWinRMClient() (*winrm.Client, error) {
	conf := c.getClientConfig()

	// Shamelessly borrowed from the winrmcp client to ensure
	// that the client is configured using the same defaulting behaviors that
	// winrmcp uses even we we aren't using winrmcp. This ensures similar
	// behavior between upload, download, and copy functions. We can't use the
	// one generated by winrmcp because it isn't exported.
	var endpoint *winrm.Endpoint
	endpoint = &winrm.Endpoint{
		Host:          c.endpoint.Host,
		Port:          c.endpoint.Port,
		HTTPS:         conf.Https,
		Insecure:      conf.Insecure,
		TLSServerName: conf.TLSServerName,
		CACert:        conf.CACertBytes,
		Timeout:       conf.ConnectTimeout,
	}
	params := winrm.NewParameters(
		winrm.DefaultParameters.Timeout,
		winrm.DefaultParameters.Locale,
		winrm.DefaultParameters.EnvelopeSize,
	)

	params.TransportDecorator = conf.TransportDecorator
	params.Timeout = "PT3M"

	client, err := winrm.NewClientWithParameters(
		endpoint, conf.Auth.User, conf.Auth.Password, params)
	return client, err
}

type Base64Pipe struct {
	w io.Writer // underlying writer (file, buffer)
}

func (d *Base64Pipe) ReadFrom(r io.Reader) (int64, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return 0, err
	}

	var i int
	i, err = d.Write(b)

	if err != nil {
		return 0, err
	}

	return int64(i), err
}

func (d *Base64Pipe) Write(p []byte) (int, error) {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(p)))

	decodedBytes, err := base64.StdEncoding.Decode(dst, p)
	if err != nil {
		return 0, err
	}

	return d.w.Write(dst[0:decodedBytes])
}
