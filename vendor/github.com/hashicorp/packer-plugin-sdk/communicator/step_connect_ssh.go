package communicator

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/terminal"

	helperssh "github.com/hashicorp/packer-plugin-sdk/communicator/ssh"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/pathing"
	"github.com/hashicorp/packer-plugin-sdk/sdk-internals/communicator/ssh"
	gossh "golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/net/proxy"
)

// StepConnectSSH is a step that only connects to SSH.
//
// In general, you should use StepConnect.
type StepConnectSSH struct {
	// All the fields below are documented on StepConnect
	Config    *Config
	Host      func(multistep.StateBag) (string, error)
	SSHConfig func(multistep.StateBag) (*gossh.ClientConfig, error)
	SSHPort   func(multistep.StateBag) (int, error)
}

func (s *StepConnectSSH) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	var comm packersdk.Communicator
	var err error

	subCtx, cancel := context.WithCancel(ctx)
	waitDone := make(chan bool, 1)
	go func() {
		ui.Say("Waiting for SSH to become available...")
		comm, err = s.waitForSSH(state, subCtx)
		cancel() // just to make 'possible context leak' analysis happy
		waitDone <- true
	}()

	log.Printf("[INFO] Waiting for SSH, up to timeout: %s", s.Config.SSHTimeout)
	timeout := time.After(s.Config.SSHTimeout)
	for {
		// Wait for either SSH to become available, a timeout to occur,
		// or an interrupt to come through.
		select {
		case <-waitDone:
			if err != nil {
				ui.Error(fmt.Sprintf("Error waiting for SSH: %s", err))
				state.Put("error", err)
				return multistep.ActionHalt
			}

			ui.Say("Connected to SSH!")
			state.Put("communicator", comm)
			return multistep.ActionContinue
		case <-timeout:
			err := fmt.Errorf("Timeout waiting for SSH.")
			state.Put("error", err)
			ui.Error(err.Error())
			cancel()
			return multistep.ActionHalt
		case <-ctx.Done():
			// The step sequence was cancelled, so cancel waiting for SSH
			// and just start the halting process.
			cancel()
			log.Println("[WARN] Interrupt detected, quitting waiting for SSH.")
			return multistep.ActionHalt
		case <-time.After(1 * time.Second):
		}
	}
}

func (s *StepConnectSSH) Cleanup(multistep.StateBag) {
}

func (s *StepConnectSSH) waitForSSH(state multistep.StateBag, ctx context.Context) (packersdk.Communicator, error) {
	// Determine if we're using a bastion host, and if so, retrieve
	// that configuration. This configuration doesn't change so we
	// do this one before entering the retry loop.
	var bProto, bAddr string
	var bConf *gossh.ClientConfig
	var pAddr string
	var pAuth *proxy.Auth
	if s.Config.SSHBastionHost != "" {
		// The protocol is hardcoded for now, but may be configurable one day
		bProto = "tcp"
		bAddr = fmt.Sprintf(
			"%s:%d", s.Config.SSHBastionHost, s.Config.SSHBastionPort)

		conf, err := sshBastionConfig(s.Config)
		if err != nil {
			return nil, fmt.Errorf("Error configuring bastion: %s", err)
		}
		bConf = conf
	}

	if s.Config.SSHProxyHost != "" {
		pAddr = fmt.Sprintf("%s:%d", s.Config.SSHProxyHost, s.Config.SSHProxyPort)
		if s.Config.SSHProxyUsername != "" {
			pAuth = new(proxy.Auth)
			pAuth.User = s.Config.SSHProxyUsername
			pAuth.Password = s.Config.SSHProxyPassword
		}

	}

	handshakeAttempts := 0

	var comm packersdk.Communicator
	first := true
	for {
		// Don't check for cancel or wait on first iteration
		if !first {
			select {
			case <-ctx.Done():
				log.Println("[DEBUG] SSH wait cancelled. Exiting loop.")
				return nil, errors.New("SSH wait cancelled")
			case <-time.After(5 * time.Second):
			}
		}
		first = false

		// First we request the TCP connection information
		host, err := s.Host(state)
		if err != nil {
			log.Printf("[DEBUG] Error getting SSH address: %s", err)
			continue
		}
		// store host and port in config so we can access them from provisioners
		s.Config.SSHHost = host
		port := s.Config.SSHPort
		if s.SSHPort != nil {
			port, err = s.SSHPort(state)
			if err != nil {
				log.Printf("[DEBUG] Error getting SSH port: %s", err)
				continue
			}
			s.Config.SSHPort = port
		}
		state.Put("communicator_config", s.Config)

		// Retrieve the SSH configuration
		sshConfig, err := s.SSHConfig(state)
		if err != nil {
			log.Printf("[DEBUG] Error getting SSH config: %s", err)
			continue
		}

		// Attempt to connect to SSH port
		var connFunc func() (net.Conn, error)
		address := fmt.Sprintf("%s:%d", host, port)
		if bAddr != "" {
			// We're using a bastion host, so use the bastion connfunc
			connFunc = ssh.BastionConnectFunc(
				bProto, bAddr, bConf, "tcp", address)
		} else if pAddr != "" {
			// Connect via SOCKS5 proxy
			connFunc = ssh.ProxyConnectFunc(pAddr, pAuth, "tcp", address)
		} else {
			// No bastion host, connect directly
			connFunc = ssh.ConnectFunc("tcp", address)
		}

		nc, err := connFunc()
		if err != nil {
			log.Printf("[DEBUG] TCP connection to SSH ip/port failed: %s", err)
			continue
		}
		nc.Close()

		// Parse out all the requested Port Tunnels that will go over our SSH connection
		var tunnels []ssh.TunnelSpec
		for _, v := range s.Config.SSHLocalTunnels {
			t, err := helperssh.ParseTunnelArgument(v, ssh.LocalTunnel)
			if err != nil {
				return nil, fmt.Errorf(
					"Error parsing port forwarding: %s", err)
			}
			tunnels = append(tunnels, t)
		}
		for _, v := range s.Config.SSHRemoteTunnels {
			t, err := helperssh.ParseTunnelArgument(v, ssh.RemoteTunnel)
			if err != nil {
				return nil, fmt.Errorf(
					"Error parsing port forwarding: %s", err)
			}
			tunnels = append(tunnels, t)
		}

		// Then we attempt to connect via SSH
		config := &ssh.Config{
			Connection:             connFunc,
			SSHConfig:              sshConfig,
			Pty:                    s.Config.SSHPty,
			DisableAgentForwarding: s.Config.SSHDisableAgentForwarding,
			UseSftp:                s.Config.SSHFileTransferMethod == "sftp",
			KeepAliveInterval:      s.Config.SSHKeepAliveInterval,
			Timeout:                s.Config.SSHReadWriteTimeout,
			Tunnels:                tunnels,
		}

		log.Printf("[INFO] Attempting SSH connection to %s...", address)
		comm, err = ssh.New(address, config)
		if err != nil {
			log.Printf("[DEBUG] SSH handshake err: %s", err)

			// Only count this as an attempt if we were able to attempt
			// to authenticate. Note this is very brittle since it depends
			// on the string of the error... but I don't see any other way.
			if strings.Contains(err.Error(), "authenticate") {
				log.Printf(
					"[DEBUG] Detected authentication error. Increasing handshake attempts.")
				err = fmt.Errorf("Packer experienced an authentication error "+
					"when trying to connect via SSH. This can happen if your "+
					"username/password are wrong. You may want to double-check"+
					" your credentials as part of your debugging process. "+
					"original error: %s",
					err)
				handshakeAttempts += 1
			}

			if handshakeAttempts < s.Config.SSHHandshakeAttempts {
				// Try to connect via SSH a handful of times. We sleep here
				// so we don't get a ton of authentication errors back to back.
				time.Sleep(2 * time.Second)
				continue
			}

			return nil, err
		}

		break
	}

	return comm, nil
}

func sshBastionConfig(config *Config) (*gossh.ClientConfig, error) {
	auth := make([]gossh.AuthMethod, 0, 2)

	if config.SSHBastionInteractive {
		var c io.ReadWriteCloser
		if terminal.IsTerminal(int(os.Stdin.Fd())) {
			c = os.Stdin
		} else {
			tty, err := os.Open("/dev/tty")
			if err != nil {
				return nil, err
			}
			defer tty.Close()
			c = tty
		}
		auth = append(auth, gossh.KeyboardInteractive(ssh.KeyboardInteractive(c)))
	}

	if config.SSHBastionPassword != "" {
		auth = append(auth,
			gossh.Password(config.SSHBastionPassword),
			gossh.KeyboardInteractive(
				ssh.PasswordKeyboardInteractive(config.SSHBastionPassword)))
	}

	if config.SSHBastionPrivateKeyFile != "" {
		path, err := pathing.ExpandUser(config.SSHBastionPrivateKeyFile)
		if err != nil {
			return nil, fmt.Errorf(
				"Error expanding path for SSH bastion private key: %s", err)
		}

		if config.SSHBastionCertificateFile != "" {
			identityPath, err := pathing.ExpandUser(config.SSHBastionCertificateFile)
			if err != nil {
				return nil, fmt.Errorf("Error expanding path for SSH bastion identity certificate: %s", err)
			}
			signer, err := helperssh.FileSignerWithCert(path, identityPath)
			if err != nil {
				return nil, err
			}
			auth = append(auth, gossh.PublicKeys(signer))
		} else {
			signer, err := helperssh.FileSigner(path)
			if err != nil {
				return nil, err
			}
			auth = append(auth, gossh.PublicKeys(signer))
		}
	}

	if config.SSHBastionAgentAuth {
		authSock := os.Getenv("SSH_AUTH_SOCK")
		if authSock == "" {
			return nil, fmt.Errorf("SSH_AUTH_SOCK is not set")
		}

		sshAgent, err := net.Dial("unix", authSock)
		if err != nil {
			return nil, fmt.Errorf("Cannot connect to SSH Agent socket %q: %s", authSock, err)
		}

		auth = append(auth, gossh.PublicKeysCallback(agent.NewClient(sshAgent).Signers))
	}

	return &gossh.ClientConfig{
		User:            config.SSHBastionUsername,
		Auth:            auth,
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	}, nil
}
