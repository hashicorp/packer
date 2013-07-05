package amazonebs

import (
	gossh "code.google.com/p/go.crypto/ssh"
	"errors"
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/communicator/ssh"
	"github.com/mitchellh/packer/packer"
	"log"
	"net"
	"time"
)

type stepConnectSSH struct {
	conn net.Conn
}

func (s *stepConnectSSH) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(config)
	instance := state["instance"].(*ec2.Instance)
	privateKey := state["privateKey"].(string)
	ui := state["ui"].(packer.Ui)

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

	// Build the actual SSH client configuration
	sshConfig := &gossh.ClientConfig{
		User: config.SSHUsername,
		Auth: []gossh.ClientAuth{
			gossh.ClientAuthKeyring(keyring),
		},
	}

	// Start trying to connect to SSH
	connected := make(chan bool, 1)
	connectQuit := make(chan bool, 1)
	defer func() {
		connectQuit <- true
	}()

	var comm packer.Communicator

	go func() {
		var err error

		ui.Say("Connecting to the instance via SSH...")
		attempts := 0
		for {
			select {
			case <-connectQuit:
				return
			default:
			}

			attempts += 1
			log.Printf(
				"Opening TCP conn for SSH to %s:%d (attempt %d)",
				instance.DNSName, config.SSHPort, attempts)
			s.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", instance.DNSName, config.SSHPort))

			if err == nil {
				comm, err = ssh.New(s.conn, sshConfig)
			}
			if err == nil {
				break
			}

			log.Printf("TCP/SSH connection failed: %s", err)

			// A brief sleep so we're not being overly zealous attempting
			// to connect to the instance.
			time.Sleep(500 * time.Millisecond)
		}

		connected <- true
	}()

	log.Printf("Waiting up to %s for SSH connection", config.SSHTimeout)
	timeout := time.After(config.SSHTimeout)

ConnectWaitLoop:
	for {
		select {
		case <-connected:
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

	if err != nil {
		err := fmt.Errorf("Error connecting to SSH: %s", err)
		state["error"] = err
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Set the communicator on the state bag so it can be used later
	state["communicator"] = comm

	return multistep.ActionContinue
}

func (s *stepConnectSSH) Cleanup(map[string]interface{}) {
	if s.conn != nil {
		s.conn.Close()
	}
}
