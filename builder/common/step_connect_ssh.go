package common

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

// StepConnectSSH is a multistep Step implementation that waits for SSH
// to become available. It gets the connection information from a single
// configuration when creating the step.
//
// Uses:
//   ui packer.Ui
//
// Produces:
//   communicator packer.Communicator
type StepConnectSSH struct {
	// SSHAddress is a function that returns the TCP address to connect to
	// for SSH. This is a function so that you can query information
	// if necessary for this address.
	SSHAddress func() (string, error)

	// SSHConfig is a function that returns the proper client configuration
	// for SSH access.
	SSHConfig func() (*gossh.ClientConfig, error)

	// SSHWaitTimeout is the total timeout to wait for SSH to become available.
	SSHWaitTimeout time.Duration

	cancel bool
	comm   packer.Communicator
}

func (s *StepConnectSSH) Run(state map[string]interface{}) multistep.StepAction {
	ui := state["ui"].(packer.Ui)

	var comm packer.Communicator
	var err error

	waitDone := make(chan bool, 1)
	go func() {
		ui.Say("Waiting for SSH to become available...")
		comm, err = s.waitForSSH()
		waitDone <- true
	}()

	log.Printf("Waiting for SSH, up to timeout: %s", s.SSHWaitTimeout)
	timeout := time.After(s.SSHWaitTimeout)
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

			ui.Say("Connected to SSH!")
			s.comm = comm
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

func (s *StepConnectSSH) Cleanup(map[string]interface{}) {
}

func (s *StepConnectSSH) waitForSSH() (packer.Communicator, error) {
	handshakeAttempts := 0

	var comm packer.Communicator
	for {
		time.Sleep(5 * time.Second)

		if s.cancel {
			log.Println("SSH wait cancelled. Exiting loop.")
			return nil, errors.New("SSH wait cancelled")
		}

		// First we request the TCP connection information
		address, err := s.SSHAddress()
		if err != nil {
			log.Printf("Error getting SSH address: %s", err)
			continue
		}

		// Retrieve the SSH configuration
		sshConfig, err := s.SSHConfig()
		if err != nil {
			log.Printf("Error getting SSH config: %s", err)
			continue
		}

		// Attempt to connect to SSH port
		connFunc := ssh.ConnectFunc("tcp", address, 5*time.Minute)
		nc, err := connFunc()
		if err != nil {
			log.Printf("TCP connection to SSH ip/port failed: %s", err)
			continue
		}
		nc.Close()

		// Then we attempt to connect via SSH
		config := &ssh.Config{
			Connection: connFunc,
			SSHConfig:  sshConfig,
		}

		comm, err = ssh.New(config)
		if err != nil {
			log.Printf("SSH handshake err: %s", err)

			handshakeAttempts += 1
			if handshakeAttempts < 10 {
				// Try to connect via SSH a handful of times
				continue
			}

			return nil, err
		}

		break
	}

	return comm, nil
}
