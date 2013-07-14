package digitalocean

import (
	gossh "code.google.com/p/go.crypto/ssh"
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/communicator/ssh"
	"github.com/mitchellh/packer/packer"
	"log"
	"time"
)

type stepConnectSSH struct {
	comm packer.Communicator
}

func (s *stepConnectSSH) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(config)
	privateKey := state["privateKey"].(string)
	ui := state["ui"].(packer.Ui)
	ipAddress := state["droplet_ip"]

	// Build the keyring for authentication. This stores the private key
	// we'll use to authenticate.
	keyring := &ssh.SimpleKeychain{}
	err := keyring.AddPEMKey(privateKey)
	if err != nil {
		err := fmt.Errorf("Error setting up SSH config: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	connFunc := ssh.ConnectFunc(
		"tcp",
		fmt.Sprintf("%s:%d", ipAddress, config.SSHPort),
		5*time.Minute)

	// Build the actual SSH client configuration
	sshConfig := &ssh.Config{
		Connection: connFunc,
		SSHConfig: &gossh.ClientConfig{
			User: config.SSHUsername,
			Auth: []gossh.ClientAuth{
				gossh.ClientAuthKeyring(keyring),
			},
		},
	}

	// Start trying to connect to SSH
	connected := make(chan error, 1)
	connectQuit := make(chan bool, 1)
	defer func() {
		connectQuit <- true
	}()

	var comm packer.Communicator
	go func() {
		ui.Say("Connecting to the droplet via SSH...")
		attempts := 0
		handshakeAttempts := 0
		for {
			select {
			case <-connectQuit:
				return
			default:
			}

			// A brief sleep so we're not being overly zealous attempting
			// to connect to the instance.
			time.Sleep(500 * time.Millisecond)

			attempts += 1
			nc, err := connFunc()
			if err != nil {
				continue
			}
			nc.Close()

			log.Println("TCP connection made. Attempting SSH handshake.")
			comm, err = ssh.New(sshConfig)
			if err == nil {
				log.Println("Connected to SSH!")
				break
			}

			handshakeAttempts += 1
			log.Printf("SSH handshake error: %s", err)

			if handshakeAttempts > 5 {
				connected <- err
				return
			}
		}

		connected <- nil
	}()

	log.Printf("Waiting up to %s for SSH connection", config.sshTimeout)
	timeout := time.After(config.sshTimeout)

ConnectWaitLoop:
	for {
		select {
		case err := <-connected:
			if err != nil {
				err := fmt.Errorf("Error connecting to SSH: %s", err)
				state["error"] = err
				ui.Error(err.Error())
				return multistep.ActionHalt
			}

			// We connected. Just break the loop.
			break ConnectWaitLoop
		case <-timeout:
			err := errors.New("Timeout waiting for SSH to become available.")
			state["error"] = err
			ui.Error(err.Error())
			return multistep.ActionHalt
		case <-time.After(1 * time.Second):
			if _, ok := state[multistep.StateCancelled]; ok {
				log.Println("Interrupt detected, quitting waiting for SSH.")
				return multistep.ActionHalt
			}
		}
	}

	// Set the communicator on the state bag so it can be used later
	s.comm = comm
	state["communicator"] = comm

	return multistep.ActionContinue
}

func (s *stepConnectSSH) Cleanup(map[string]interface{}) {
	if s.comm != nil {
		// TODO: close
		s.comm = nil
	}
}
