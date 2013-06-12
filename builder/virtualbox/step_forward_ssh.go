package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"math/rand"
	"net"
	"time"
)

// This step adds a NAT port forwarding definition so that SSH is available
// on the guest machine.
//
// Uses:
//
// Produces:
type stepForwardSSH struct{}

func (s *stepForwardSSH) Run(state map[string]interface{}) multistep.StepAction {
	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)
	vmName := state["vmName"].(string)

	log.Printf("Looking for available SSH port between %d and %d", config.SSHHostPortMin, config.SSHHostPortMax)
	var sshHostPort uint
	portRange := int(config.SSHHostPortMax - config.SSHHostPortMin)
	for {
		sshHostPort = uint(rand.Intn(portRange)) + config.SSHHostPortMin
		log.Printf("Trying port: %d", sshHostPort)
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", sshHostPort))
		if err == nil {
			defer l.Close()
			break
		}
	}

	// Attach the disk to the controller
	ui.Say(fmt.Sprintf("Creating forwarded port mapping for SSH (host port %d)", sshHostPort))
	command := []string{
		"modifyvm", vmName,
		"--natpf1",
		fmt.Sprintf("packerssh,tcp,127.0.0.1,%d,,%d", sshHostPort, config.SSHPort),
	}
	if err := driver.VBoxManage(command...); err != nil {
		ui.Error(fmt.Sprintf("Error creating port forwarding rule: %s", err))
		return multistep.ActionHalt
	}

	time.Sleep(15 * time.Second)
	return multistep.ActionContinue
}

func (s *stepForwardSSH) Cleanup(state map[string]interface{}) {}
