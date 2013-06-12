package virtualbox

import (
	gossh "code.google.com/p/go.crypto/ssh"
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/communicator/ssh"
	"github.com/mitchellh/packer/packer"
	"log"
	"net"
	"time"
)

// This step waits for SSH to become available and establishes an SSH
// connection.
//
// Uses:
//   config *config
//   sshHostPort uint
//   ui     packer.Ui
//
// Produces:
//   communicator packer.Communicator
type stepWaitForSSH struct {
	cancel bool
	conn   net.Conn
}

func (s *stepWaitForSSH) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	ui := state["ui"].(packer.Ui)

	var comm packer.Communicator
	var err error

	waitDone := make(chan bool, 1)
	go func() {
		comm, err = s.waitForSSH(state)
		waitDone <- true
	}()

	log.Printf("Waiting for SSH, up to timeout: %s", config.SSHWaitTimeout.String())

	timeout := time.After(config.SSHWaitTimeout)
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

func (s *stepWaitForSSH) Cleanup(map[string]interface{}) {
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}
}

// This blocks until SSH becomes available, and sends the communicator
// on the given channel.
func (s *stepWaitForSSH) waitForSSH(state map[string]interface{}) (packer.Communicator, error) {
	config := state["config"].(*config)
	ui := state["ui"].(packer.Ui)
	sshHostPort := state["sshHostPort"].(uint)

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
		nc, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", sshHostPort))
		if err != nil {
			log.Printf("TCP connection to SSH ip/port failed: %s", err)
			continue
		}

		// Then we attempt to connect via SSH
		sshConfig := &gossh.ClientConfig{
			User: config.SSHUser,
			Auth: []gossh.ClientAuth{
				gossh.ClientAuthPassword(ssh.Password(config.SSHPassword)),
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
