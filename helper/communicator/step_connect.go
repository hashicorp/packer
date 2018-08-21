package communicator

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/communicator/none"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	gossh "golang.org/x/crypto/ssh"
)

// StepConnect is a multistep Step implementation that connects to
// the proper communicator and stores it in the "communicator" key in the
// state bag.
type StepConnect struct {
	// Config is the communicator config struct
	Config *Config

	// Host should return a host that can be connected to for communicator
	// connections.
	Host func(multistep.StateBag) (string, error)

	// The fields below are callbacks to assist with connecting to SSH.
	//
	// SSHConfig should return the default configuration for
	// connecting via SSH.
	SSHConfig func(multistep.StateBag) (*gossh.ClientConfig, error)
	SSHPort   func(multistep.StateBag) (int, error)

	// The fields below are callbacks to assist with connecting to WinRM.
	//
	// WinRMConfig should return the default configuration for
	// connecting via WinRM.
	WinRMConfig func(multistep.StateBag) (*WinRMConfig, error)
	WinRMPort   func(multistep.StateBag) (int, error)

	// CustomConnect can be set to have custom connectors for specific
	// types. These take highest precedence so you can also override
	// existing types.
	CustomConnect map[string]multistep.Step

	substep multistep.Step
}

func (s *StepConnect) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	typeMap := map[string]multistep.Step{
		"none": nil,
		"ssh": &StepConnectSSH{
			Config:    s.Config,
			Host:      s.Host,
			SSHConfig: s.SSHConfig,
			SSHPort:   s.SSHPort,
		},
		"winrm": &StepConnectWinRM{
			Config:      s.Config,
			Host:        s.Host,
			WinRMConfig: s.WinRMConfig,
			WinRMPort:   s.WinRMPort,
		},
	}
	for k, v := range s.CustomConnect {
		typeMap[k] = v
	}

	step, ok := typeMap[s.Config.Type]
	if !ok {
		state.Put("error", fmt.Errorf("unknown communicator type: %s", s.Config.Type))
		return multistep.ActionHalt
	}

	if step == nil {
		if comm, err := none.New("none"); err != nil {
			err := fmt.Errorf("Failed to set communicator 'none': %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt

		} else {
			state.Put("communicator", comm)
			log.Printf("[INFO] communicator disabled, will not connect")
		}
		return multistep.ActionContinue
	}

	if host, err := s.Host(state); err == nil {
		ui.Say(fmt.Sprintf("Using %s communicator to connect: %s", s.Config.Type, host))

	} else {
		log.Printf("[DEBUG] Unable to get address during connection step: %s", err)
	}

	s.substep = step
	return s.substep.Run(ctx, state)
}

func (s *StepConnect) Cleanup(state multistep.StateBag) {
	if s.substep != nil {
		s.substep.Cleanup(state)
	}
}
