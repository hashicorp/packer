package communicator

import (
	"github.com/mitchellh/multistep"
	gossh "golang.org/x/crypto/ssh"
)

// StepConnect is a multistep Step implementation that connects to
// the proper communicator and stores it in the "communicator" key in the
// state bag.
type StepConnect struct {
	// Config is the communicator config struct
	Config *Config

	// The fields below are callbacks to assist with connecting to SSH.
	//
	// SSHAddress should return the default host to connect to for SSH.
	// This is only called if ssh_host isn't specified in the config.
	//
	// SSHConfig should return the default configuration for
	// connecting via SSH.
	SSHAddress func(multistep.StateBag) (string, error)
	SSHConfig  func(multistep.StateBag) (*gossh.ClientConfig, error)

	substep multistep.Step
}

func (s *StepConnect) Run(state multistep.StateBag) multistep.StepAction {
	// Eventually we might switch between multiple of these depending
	// on the communicator type.
	s.substep = &StepConnectSSH{
		Config:     s.Config,
		SSHAddress: s.SSHAddress,
		SSHConfig:  s.SSHConfig,
	}

	return s.substep.Run(state)
}

func (s *StepConnect) Cleanup(state multistep.StateBag) {
	if s.substep != nil {
		s.substep.Cleanup(state)
	}
}
