package communicator

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/sdk-internals/communicator/none"
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

func (s *StepConnect) pause(pauseLen time.Duration, ctx context.Context) bool {
	// Use a select to determine if we get cancelled during the wait
	select {
	case <-ctx.Done():
		return true
	case <-time.After(pauseLen):
	}
	log.Printf("Pause over; connecting...")
	return false
}

func (s *StepConnect) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

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
	action := s.substep.Run(ctx, state)
	if action == multistep.ActionHalt {
		return action
	}

	if s.Config.PauseBeforeConnect > 0 {
		ui.Say(fmt.Sprintf("Pausing %s before connecting...",
			s.Config.PauseBeforeConnect.String()))
		cancelled := s.pause(s.Config.PauseBeforeConnect, ctx)
		if cancelled {
			return multistep.ActionHalt
		}
		// After pause is complete, re-run the connect substep to make sure
		// you've connected properly
		action := s.substep.Run(ctx, state)
		if action == multistep.ActionHalt {
			return action
		}
	}

	// Put communicator config into state so we can pass it to provisioners
	// for specialized interpolation later
	state.Put("communicator_config", s.Config)

	return multistep.ActionContinue
}

func (s *StepConnect) Cleanup(state multistep.StateBag) {
	if s.substep != nil {
		s.substep.Cleanup(state)
	}
}
