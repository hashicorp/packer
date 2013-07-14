package vmware

import (
	gossh "code.google.com/p/go.crypto/ssh"
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/communicator/ssh"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"log"
	"os"
	"time"
)

// This step waits for SSH to become available and establishes an SSH
// connection.
//
// Uses:
//   config *config
//   ui     packer.Ui
//   vmx_path string
//
// Produces:
//   communicator packer.Communicator
type stepWaitForSSH struct {
	cancel bool
	comm   packer.Communicator
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

	log.Printf("Waiting for SSH, up to timeout: %s", config.sshWaitTimeout.String())
	timeout := time.After(config.sshWaitTimeout)
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

func (s *stepWaitForSSH) Cleanup(map[string]interface{}) {
	if s.comm != nil {
		// TODO: close
		s.comm = nil
	}
}

// Reads the network information for lookup via DHCP.
func (s *stepWaitForSSH) dhcpLeaseLookup(vmxPath string) (GuestIPFinder, error) {
	f, err := os.Open(vmxPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	vmxBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	vmxData := ParseVMX(string(vmxBytes))

	var ok bool
	macAddress := ""
	if macAddress, ok = vmxData["ethernet0.address"]; !ok || macAddress == "" {
		if macAddress, ok = vmxData["ethernet0.generatedAddress"]; !ok || macAddress == "" {
			return nil, errors.New("couldn't find MAC address in VMX")
		}
	}

	return &DHCPLeaseGuestLookup{"vmnet8", macAddress}, nil
}

// This blocks until SSH becomes available, and sends the communicator
// on the given channel.
func (s *stepWaitForSSH) waitForSSH(state map[string]interface{}) (packer.Communicator, error) {
	config := state["config"].(*config)
	ui := state["ui"].(packer.Ui)
	vmxPath := state["vmx_path"].(string)

	handshakeAttempts := 0

	ui.Say("Waiting for SSH to become available...")
	var comm packer.Communicator
	for {
		time.Sleep(5 * time.Second)

		if s.cancel {
			log.Println("SSH wait cancelled. Exiting loop.")
			return nil, errors.New("SSH wait cancelled")
		}

		// First we wait for the IP to become available...
		log.Println("Lookup up IP information...")
		ipLookup, err := s.dhcpLeaseLookup(vmxPath)
		if err != nil {
			log.Printf("Can't lookup via DHCP lease: %s", err)
		}

		ip, err := ipLookup.GuestIP()
		if err != nil {
			log.Printf("IP lookup failed: %s", err)
			continue
		}

		log.Printf("Detected IP: %s", ip)

		// Attempt to connect to SSH port
		connFunc := ssh.ConnectFunc(
			"tcp", fmt.Sprintf("%s:%d", ip, config.SSHPort), 5*time.Minute)
		nc, err := connFunc()
		if err != nil {
			log.Printf("TCP connection to SSH ip/port failed: %s", err)
			continue
		}
		nc.Close()

		// Then we attempt to connect via SSH
		config := &ssh.Config{
			Connection: connFunc,
			SSHConfig: &gossh.ClientConfig{
				User: config.SSHUser,
				Auth: []gossh.ClientAuth{
					gossh.ClientAuthPassword(ssh.Password(config.SSHPassword)),
					gossh.ClientAuthKeyboardInteractive(
						ssh.PasswordKeyboardInteractive(config.SSHPassword)),
				},
			},
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

		ui.Say("Connected via SSH!")
		break
	}

	return comm, nil
}
