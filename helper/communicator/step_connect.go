package communicator

import (
	"fmt"
	"log"

	"github.com/mitchellh/multistep"
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

	substep multistep.Step
}

func (s *StepConnect) Run(state multistep.StateBag) multistep.StepAction {
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
		},
	}

	step, ok := typeMap[s.Config.Type]
	if !ok {
		state.Put("error", fmt.Errorf("unknown communicator type: %s", s.Config.Type))
		return multistep.ActionHalt
	}

	if step == nil {
		log.Printf("[INFO] communicator disabled, will not connect")
		return multistep.ActionContinue
	}

	s.substep = step
	return s.substep.Run(state)
}

func (s *StepConnect) Cleanup(state multistep.StateBag) {
	if s.substep != nil {
		s.substep.Cleanup(state)
	}
}
