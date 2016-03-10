package communicator

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/communicator/winrm"
	"github.com/mitchellh/packer/packer"
)

// StepConnectWinRM is a multistep Step implementation that waits for WinRM
// to become available. It gets the connection information from a single
// configuration when creating the step.
//
// Uses:
//   ui packer.Ui
//
// Produces:
//   communicator packer.Communicator
type StepConnectWinRM struct {
	// All the fields below are documented on StepConnect
	Config      *Config
	Host        func(multistep.StateBag) (string, error)
	WinRMConfig func(multistep.StateBag) (*WinRMConfig, error)
	WinRMPort   func(multistep.StateBag) (int, error)
}

func (s *StepConnectWinRM) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	var comm packer.Communicator
	var err error

	cancel := make(chan struct{})
	waitDone := make(chan bool, 1)
	go func() {
		ui.Say("Waiting for WinRM to become available...")
		comm, err = s.waitForWinRM(state, cancel)
		waitDone <- true
	}()

	log.Printf("Waiting for WinRM, up to timeout: %s", s.Config.WinRMTimeout)
	timeout := time.After(s.Config.WinRMTimeout)
WaitLoop:
	for {
		// Wait for either WinRM to become available, a timeout to occur,
		// or an interrupt to come through.
		select {
		case <-waitDone:
			if err != nil {
				ui.Error(fmt.Sprintf("Error waiting for WinRM: %s", err))
				return multistep.ActionHalt
			}

			ui.Say("Connected to WinRM!")
			state.Put("communicator", comm)
			break WaitLoop
		case <-timeout:
			err := fmt.Errorf("Timeout waiting for WinRM.")
			state.Put("error", err)
			ui.Error(err.Error())
			close(cancel)
			return multistep.ActionHalt
		case <-time.After(1 * time.Second):
			if _, ok := state.GetOk(multistep.StateCancelled); ok {
				// The step sequence was cancelled, so cancel waiting for WinRM
				// and just start the halting process.
				close(cancel)
				log.Println("Interrupt detected, quitting waiting for WinRM.")
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (s *StepConnectWinRM) Cleanup(multistep.StateBag) {
}

func (s *StepConnectWinRM) waitForWinRM(state multistep.StateBag, cancel <-chan struct{}) (packer.Communicator, error) {
	var comm packer.Communicator
	for {
		select {
		case <-cancel:
			log.Println("[INFO] WinRM wait cancelled. Exiting loop.")
			return nil, errors.New("WinRM wait cancelled")
		case <-time.After(5 * time.Second):
		}

		host, err := s.Host(state)
		if err != nil {
			log.Printf("[DEBUG] Error getting WinRM host: %s", err)
			continue
		}
		port := s.Config.WinRMPort
		if s.WinRMPort != nil {
			port, err = s.WinRMPort(state)
			if err != nil {
				log.Printf("[DEBUG] Error getting WinRM port: %s", err)
				continue
			}
		}

		user := s.Config.WinRMUser
		password := s.Config.WinRMPassword
		if s.WinRMConfig != nil {
			config, err := s.WinRMConfig(state)
			if err != nil {
				log.Printf("[DEBUG] Error getting WinRM config: %s", err)
				continue
			}

			if config.Username != "" {
				user = config.Username
			}
			if config.Password != "" {
				password = config.Password
			}
		}

		log.Println("[INFO] Attempting WinRM connection...")
		comm, err = winrm.New(&winrm.Config{
			Host:               host,
			Port:               port,
			Username:           user,
			Password:           password,
			Timeout:            s.Config.WinRMTimeout,
			Https:              s.Config.WinRMUseSSL,
			Insecure:           s.Config.WinRMInsecure,
			TransportDecorator: s.Config.WinRMTransportDecorator,
		})
		if err != nil {
			log.Printf("[ERROR] WinRM connection err: %s", err)
			continue
		}

		break
	}

	return comm, nil
}
