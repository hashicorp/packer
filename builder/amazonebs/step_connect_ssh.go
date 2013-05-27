package amazonebs

import (
	gossh "code.google.com/p/go.crypto/ssh"
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/packer/communicator/ssh"
	"github.com/mitchellh/packer/packer"
	"log"
	"net"
	"time"
)

type stepConnectSSH struct {
	conn net.Conn
}

func (s *stepConnectSSH) Run(state map[string]interface{}) StepAction {
	config := state["config"].(config)
	instance := state["instance"].(*ec2.Instance)
	privateKey := state["privateKey"].(string)
	ui := state["ui"].(packer.Ui)

	// Build the keyring for authentication. This stores the private key
	// we'll use to authenticate.
	keyring := &ssh.SimpleKeychain{}
	err := keyring.AddPEMKey(privateKey)
	if err != nil {
		ui.Say(fmt.Sprintf("Error setting up SSH config: %s", err))
		return StepHalt
	}

	// Build the actual SSH client configuration
	sshConfig := &gossh.ClientConfig{
		User: config.SSHUsername,
		Auth: []gossh.ClientAuth{
			gossh.ClientAuthKeyring(keyring),
		},
	}

	// Try to connect for SSH a few times
	ui.Say("Connecting to the instance via SSH...")
	for i := 0; i < 5; i++ {
		time.Sleep(time.Duration(i) * time.Second)

		log.Printf(
			"Opening TCP conn for SSH to %s:%d (attempt %d)",
			instance.DNSName, config.SSHPort, i+1)
		s.conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", instance.DNSName, config.SSHPort))
		if err != nil {
			continue
		}
	}

	var comm packer.Communicator
	if err == nil {
		comm, err = ssh.New(s.conn, sshConfig)
	}

	if err != nil {
		ui.Error(fmt.Sprintf("Error connecting to SSH: %s", err))
		return StepHalt
	}

	// Set the communicator on the state bag so it can be used later
	state["communicator"] = comm

	return StepContinue
}

func (s *stepConnectSSH) Cleanup(map[string]interface{}) {
	if s.conn != nil {
		s.conn.Close()
	}
}
