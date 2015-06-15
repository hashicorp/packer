package communicator

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/communicator/ssh"
	"github.com/mitchellh/packer/packer"
	gossh "golang.org/x/crypto/ssh"
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

func (s *StepConnectSSH) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	var comm packer.Communicator
	var err error

	cancel := make(chan struct{})
	waitDone := make(chan bool, 1)
	go func() {
		ui.Say("Waiting for SSH to become available...")
		comm, err = s.waitForSSH(state, cancel)
		waitDone <- true
	}()

	log.Printf("[INFO] Waiting for SSH, up to timeout: %s", s.Config.SSHTimeout)
	timeout := time.After(s.Config.SSHTimeout)
WaitLoop:
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
			break WaitLoop
		case <-timeout:
			err := fmt.Errorf("Timeout waiting for SSH.")
			state.Put("error", err)
			ui.Error(err.Error())
			close(cancel)
			return multistep.ActionHalt
		case <-time.After(1 * time.Second):
			if _, ok := state.GetOk(multistep.StateCancelled); ok {
				// The step sequence was cancelled, so cancel waiting for SSH
				// and just start the halting process.
				close(cancel)
				log.Println("[WARN] Interrupt detected, quitting waiting for SSH.")
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (s *StepConnectSSH) Cleanup(multistep.StateBag) {
}

func (s *StepConnectSSH) waitForSSH(state multistep.StateBag, cancel <-chan struct{}) (packer.Communicator, error) {
	handshakeAttempts := 0

	var comm packer.Communicator
	first := true
	for {
		// Don't check for cancel or wait on first iteration
		if !first {
			select {
			case <-cancel:
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
		port := s.Config.SSHPort
		if s.SSHPort != nil {
			port, err = s.SSHPort(state)
			if err != nil {
				log.Printf("[DEBUG] Error getting SSH port: %s", err)
				continue
			}
		}

		// Retrieve the SSH configuration
		sshConfig, err := s.SSHConfig(state)
		if err != nil {
			log.Printf("[DEBUG] Error getting SSH config: %s", err)
			continue
		}

		address := fmt.Sprintf("%s:%d", host, port)

		// Attempt to connect to SSH port
		connFunc := ssh.ConnectFunc("tcp", address)
		nc, err := connFunc()
		if err != nil {
			log.Printf("[DEBUG] TCP connection to SSH ip/port failed: %s", err)
			continue
		}
		nc.Close()

		// Then we attempt to connect via SSH
		config := &ssh.Config{
			Connection: connFunc,
			SSHConfig:  sshConfig,
			Pty:        s.Config.SSHPty,
		}

		log.Println("[INFO] Attempting SSH connection...")
		comm, err = ssh.New(address, config)
		if err != nil {
			log.Printf("[DEBUG] SSH handshake err: %s", err)

			// Only count this as an attempt if we were able to attempt
			// to authenticate. Note this is very brittle since it depends
			// on the string of the error... but I don't see any other way.
			if strings.Contains(err.Error(), "authenticate") {
				log.Printf(
					"[DEBUG] Detected authentication error. Increasing handshake attempts.")
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
