package packer

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	pluginsdk "github.com/hashicorp/packer-plugin-sdk/plugin"
	packerrpc "github.com/hashicorp/packer-plugin-sdk/rpc"
)

// If this is true, then the "unexpected EOF" panic will not be
// raised throughout the clients.
var Killed = false

// This is a slice of the "managed" clients which are cleaned up when
// calling Cleanup
var managedClients = make([]*PluginClient, 0, 5)

// Client handles the lifecycle of a plugin application, determining its
// RPC address, and returning various types of packer interface implementations
// across the multi-process communication layer.
type PluginClient struct {
	config      *PluginClientConfig
	exited      bool
	doneLogging chan struct{}
	l           sync.Mutex
	address     net.Addr
}

// PluginClientConfig is the configuration used to initialize a new
// plugin client. After being used to initialize a plugin client,
// that configuration must not be modified again.
type PluginClientConfig struct {
	// The unstarted subprocess for starting the plugin.
	Cmd *exec.Cmd

	// Managed represents if the client should be managed by the
	// plugin package or not. If true, then by calling CleanupClients,
	// it will automatically be cleaned up. Otherwise, the client
	// user is fully responsible for making sure to Kill all plugin
	// clients. By default the client is _not_ managed.
	Managed bool

	// The minimum and maximum port to use for communicating with
	// the subprocess. If not set, this defaults to 10,000 and 25,000
	// respectively.
	MinPort, MaxPort int

	// StartTimeout is the timeout to wait for the plugin to say it
	// has started successfully.
	StartTimeout time.Duration

	// If non-nil, then the stderr of the client will be written to here
	// (as well as the log).
	Stderr io.Writer
}

// This makes sure all the managed subprocesses are killed and properly
// logged. This should be called before the parent process running the
// plugins exits.
//
// This must only be called _once_.
func CleanupClients() {
	// Set the killed to true so that we don't get unexpected panics
	Killed = true

	// Kill all the managed clients in parallel and use a WaitGroup
	// to wait for them all to finish up.
	var wg sync.WaitGroup
	for _, client := range managedClients {
		wg.Add(1)

		go func(client *PluginClient) {
			client.Kill()
			wg.Done()
		}(client)
	}

	log.Println("waiting for all plugin processes to complete...")
	wg.Wait()
}

// Creates a new plugin client which manages the lifecycle of an external
// plugin and gets the address for the RPC connection.
//
// The client must be cleaned up at some point by calling Kill(). If
// the client is a managed client (created with NewManagedClient) you
// can just call CleanupClients at the end of your program and they will
// be properly cleaned.
func NewClient(config *PluginClientConfig) (c *PluginClient) {
	if config.MinPort == 0 && config.MaxPort == 0 {
		config.MinPort = 10000
		config.MaxPort = 25000
	}

	if config.StartTimeout == 0 {
		config.StartTimeout = 1 * time.Minute
	}

	if config.Stderr == nil {
		config.Stderr = ioutil.Discard
	}

	c = &PluginClient{config: config}
	if config.Managed {
		managedClients = append(managedClients, c)
	}

	return
}

// Tells whether or not the underlying process has exited.
func (c *PluginClient) Exited() bool {
	c.l.Lock()
	defer c.l.Unlock()
	return c.exited
}

// Returns a builder implementation that is communicating over this
// client. If the client hasn't been started, this will start it.
func (c *PluginClient) Builder() (packersdk.Builder, error) {
	client, err := c.Client()
	if err != nil {
		return nil, err
	}

	return &cmdBuilder{client.Builder(), c}, nil
}

// Returns a hook implementation that is communicating over this
// client. If the client hasn't been started, this will start it.
func (c *PluginClient) Hook() (packersdk.Hook, error) {
	client, err := c.Client()
	if err != nil {
		return nil, err
	}

	return &cmdHook{client.Hook(), c}, nil
}

// Returns a post-processor implementation that is communicating over
// this client. If the client hasn't been started, this will start it.
func (c *PluginClient) PostProcessor() (packersdk.PostProcessor, error) {
	client, err := c.Client()
	if err != nil {
		return nil, err
	}

	return &cmdPostProcessor{client.PostProcessor(), c}, nil
}

// Returns a provisioner implementation that is communicating over this
// client. If the client hasn't been started, this will start it.
func (c *PluginClient) Provisioner() (packersdk.Provisioner, error) {
	client, err := c.Client()
	if err != nil {
		return nil, err
	}

	return &cmdProvisioner{client.Provisioner(), c}, nil
}

// Returns a data source implementation that is communicating over this
// client. If the client hasn't been started, this will start it.
func (c *PluginClient) Datasource() (packersdk.Datasource, error) {
	client, err := c.Client()
	if err != nil {
		return nil, err
	}
	return &cmdDatasource{client.Datasource(), c}, nil
}

// End the executing subprocess (if it is running) and perform any cleanup
// tasks necessary such as capturing any remaining logs and so on.
//
// This method blocks until the process successfully exits.
//
// This method can safely be called multiple times.
func (c *PluginClient) Kill() {
	cmd := c.config.Cmd

	if cmd.Process == nil {
		return
	}

	cmd.Process.Kill()

	// Wait for the client to finish logging so we have a complete log
	<-c.doneLogging
}

// Starts the underlying subprocess, communicating with it to negotiate
// a port for RPC connections, and returning the address to connect via RPC.
//
// This method is safe to call multiple times. Subsequent calls have no effect.
// Once a client has been started once, it cannot be started again, even if
// it was killed.
func (c *PluginClient) Start() (net.Addr, error) {
	c.l.Lock()
	defer c.l.Unlock()

	if c.address != nil {
		return c.address, nil
	}

	c.doneLogging = make(chan struct{})

	env := []string{
		fmt.Sprintf("%s=%s", pluginsdk.MagicCookieKey, pluginsdk.MagicCookieValue),
		fmt.Sprintf("PACKER_PLUGIN_MIN_PORT=%d", c.config.MinPort),
		fmt.Sprintf("PACKER_PLUGIN_MAX_PORT=%d", c.config.MaxPort),
	}

	stdout_r, stdout_w := io.Pipe()
	stderr_r, stderr_w := io.Pipe()

	cmd := c.config.Cmd
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, env...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = stderr_w
	cmd.Stdout = stdout_w

	log.Printf("Starting plugin: %s %#v", cmd.Path, cmd.Args)
	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	// Make sure the command is properly cleaned up if there is an error
	defer func() {
		r := recover()

		if err != nil || r != nil {
			cmd.Process.Kill()
		}

		if r != nil {
			panic(r)
		}
	}()

	// Start goroutine to wait for process to exit
	exitCh := make(chan struct{})
	go func() {
		// Make sure we close the write end of our stderr/stdout so
		// that the readers send EOF properly.
		defer stderr_w.Close()
		defer stdout_w.Close()

		// Wait for the command to end.
		cmd.Wait()

		// Log and make sure to flush the logs write away
		log.Printf("%s: plugin process exited\n", cmd.Path)
		os.Stderr.Sync()

		// Mark that we exited
		close(exitCh)

		// Set that we exited, which takes a lock
		c.l.Lock()
		defer c.l.Unlock()
		c.exited = true
	}()

	// Start goroutine that logs the stderr
	go c.logStderr(stderr_r)

	// Start a goroutine that is going to be reading the lines
	// out of stdout
	linesCh := make(chan []byte)
	go func() {
		defer close(linesCh)

		buf := bufio.NewReader(stdout_r)
		for {
			line, err := buf.ReadBytes('\n')
			if line != nil {
				linesCh <- line
			}

			if err == io.EOF {
				return
			}
		}
	}()

	// Make sure after we exit we read the lines from stdout forever
	// so they dont' block since it is an io.Pipe
	defer func() {
		go func() {
			for range linesCh {
			}
		}()
	}()

	// Some channels for the next step
	timeout := time.After(c.config.StartTimeout)

	// Start looking for the address
	log.Printf("Waiting for RPC address for: %s", cmd.Path)
	select {
	case <-timeout:
		err = errors.New("timeout while waiting for plugin to start")
	case <-exitCh:
		err = errors.New("plugin exited before we could connect")
	case lineBytes := <-linesCh:
		// Trim the line and split by "|" in order to get the parts of
		// the output.
		line := strings.TrimSpace(string(lineBytes))
		parts := strings.SplitN(line, "|", 4)
		if len(parts) == 3 {
			// In protocol version 4 and before, the protocol only had a Major
			// version
			err = fmt.Errorf("The protocol of this plugin (protocol version 4 " +
				"and lower) was deprecated, please use a newer version of this plugin." +
				"Or use an older version of Packer (pre 1.7) with this plugin.")
			return nil, err
		}
		if len(parts) < 4 {
			err = fmt.Errorf("Unrecognized remote plugin message: %s", line)
			return nil, err
		}
		pluginMajorAPIVersion, pluginMinorAPIVersion, network, netAddr := parts[0], parts[1], parts[2], parts[3]

		// Test the API versions
		if pluginMajorAPIVersion != pluginsdk.APIVersionMajor {
			err = fmt.Errorf("Incompatible API MAJOR version with plugin. "+
				"plugin MINOR API version: %s, Ours: %s", pluginMajorAPIVersion, pluginsdk.APIVersionMajor)
			return nil, err
		}
		if pluginMinorAPIVersion > pluginsdk.APIVersionMinor {
			err = fmt.Errorf("Incompatible API MINOR version with plugin. "+
				"plugin MINOR API version: %s, Ours: %s. Please upgrade Packer.", pluginMinorAPIVersion, pluginsdk.APIVersionMinor)
			return nil, err
		}

		switch network {
		case "tcp":
			c.address, err = net.ResolveTCPAddr("tcp", netAddr)
		case "unix":
			c.address, err = net.ResolveUnixAddr("unix", netAddr)
		default:
			return nil, fmt.Errorf("Unknown address type: %s", network)
		}
		log.Printf("Received %s RPC address for %s: addr is %s", network, cmd.Path, c.address)
	}

	return c.address, err
}

func (c *PluginClient) logStderr(r io.Reader) {
	logPrefix := filepath.Base(c.config.Cmd.Path)
	if logPrefix == "packer" {
		// we just called the normal packer binary with the plugin arg.
		// grab the last arg from the list which will match the plugin name.
		logPrefix = c.config.Cmd.Args[len(c.config.Cmd.Args)-1]
	}

	bufR := bufio.NewReader(r)
	for {
		line, err := bufR.ReadString('\n')
		if line != "" {
			c.config.Stderr.Write([]byte(line))

			line = strings.TrimRightFunc(line, unicode.IsSpace)

			log.Printf("%s plugin: %s", logPrefix, line)
		}

		if err == io.EOF {
			break
		}
	}

	// Flag that we've completed logging for others
	close(c.doneLogging)
}

func (c *PluginClient) Client() (*packerrpc.Client, error) {
	addr, err := c.Start()
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial(addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		// Make sure to set keep alive so that the connection doesn't die
		tcpConn.SetKeepAlive(true)
	}

	client, err := packerrpc.NewClient(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return client, nil
}
