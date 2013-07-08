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
	cancel bool
	conn   net.Conn
}

func (s *stepConnectSSH) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(config)
	ui := state["ui"].(packer.Ui)

	var comm packer.Communicator
	var err error

	waitDone := make(chan bool, 1)
	go func() {
		comm, err = s.waitForSSH(state)
		waitDone <- true
	}()

	log.Printf("Waiting for SSH, up to timeout: %s", config.SSHTimeout.String())

	timeout := time.After(config.SSHTimeout)
WaitLoop:
	for {
		// Wait for either SSH to become available, a timeout to occur,
		// or an interrupt to come through.
		select {
		case <-waitDone:
			if err != nil {
				ui.Error(fmt.Sprintf("Error waiting for SSH: %s", err))
				return multistep.ActionHalt
			}

			state["communicator"] = comm
			break WaitLoop
		case <-timeout:
			ui.Error("Timeout waiting for SSH.")
			s.cancel = true
			return multistep.ActionHalt
		case <-time.After(1 * time.Second):
			if _, ok := state[multistep.StateCancelled]; ok {
				log.Println("Interrupt detected, quitting waiting for SSH.")
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (s *stepConnectSSH) Cleanup(map[string]interface{}) {
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
}

// This blocks until SSH becomes available, and sends the communicator
// on the given channel.
func (s *stepConnectSSH) waitForSSH(state map[string]interface{}) (packer.Communicator, error) {
	config := state["config"].(config)
	instance := state["instance"].(*ec2.Instance)
	privateKey := state["privateKey"].(string)
	ui := state["ui"].(packer.Ui)

	// Build the keyring for authentication. This stores the private key
	// we'll use to authenticate.
	keyring := &ssh.SimpleKeychain{}
	err := keyring.AddPEMKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("Error setting up SSH config: %s", err)
	}

	ui.Say("Waiting for SSH to become available...")
	var comm packer.Communicator
	var nc net.Conn
	for {
		if nc != nil {
			nc.Close()
		}

		time.Sleep(5 * time.Second)

		if s.cancel {
			log.Println("SSH wait cancelled. Exiting loop.")
			return nil, errors.New("SSH wait cancelled")
		}

		// Attempt to connect to SSH port
		log.Printf(
			"Opening TCP conn for SSH to %s:%d",
			instance.DNSName, config.SSHPort)
		nc, err := net.Dial("tcp",
			fmt.Sprintf("%s:%d", instance.DNSName, config.SSHPort))
		if err != nil {
			log.Printf("TCP connection to SSH ip/port failed: %s", err)
			continue
		}

		// Build the actual SSH client configuration
		sshConfig := &gossh.ClientConfig{
			User: config.SSHUsername,
			Auth: []gossh.ClientAuth{
				gossh.ClientAuthKeyring(keyring),
			},
		}

		sshConnectSuccess := make(chan bool, 1)
		go func() {
			comm, err = ssh.New(nc, sshConfig)
			if err != nil {
				log.Printf("SSH connection fail: %s", err)
				sshConnectSuccess <- false
				return
			}

			sshConnectSuccess <- true
		}()

		select {
		case success := <-sshConnectSuccess:
			if !success {
				continue
			}
		case <-time.After(5 * time.Second):
			log.Printf("SSH handshake timeout. Trying again.")
			continue
		}

		ui.Say("Connected via SSH!")
		break
	}

	// Store the connection so we can close it later
	s.conn = nc
	return comm, nil
}
