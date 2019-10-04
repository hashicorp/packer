package communicator

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/packer/communicator/winrm"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	winrmcmd "github.com/masterzen/winrm"
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

func (s *StepConnectWinRM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	var comm packer.Communicator
	var err error

	subCtx, cancel := context.WithCancel(ctx)
	waitDone := make(chan bool, 1)
	go func() {
		ui.Say("Waiting for WinRM to become available...")
		comm, err = s.waitForWinRM(state, subCtx)
		cancel() // just to make 'possible context leak' analysis happy
		waitDone <- true
	}()

	log.Printf("Waiting for WinRM, up to timeout: %s", s.Config.WinRMTimeout)
	timeout := time.After(s.Config.WinRMTimeout)
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
			return multistep.ActionContinue
		case <-timeout:
			err := fmt.Errorf("Timeout waiting for WinRM.")
			state.Put("error", err)
			ui.Error(err.Error())
			cancel()
			return multistep.ActionHalt
		case <-ctx.Done():
			// The step sequence was cancelled, so cancel waiting for WinRM
			// and just start the halting process.
			cancel()
			log.Println("Interrupt detected, quitting waiting for WinRM.")
			return multistep.ActionHalt
		case <-time.After(1 * time.Second):
		}
	}
}

func (s *StepConnectWinRM) Cleanup(multistep.StateBag) {
}

func (s *StepConnectWinRM) waitForWinRM(state multistep.StateBag, ctx context.Context) (packer.Communicator, error) {
	var comm packer.Communicator
	first := true
	for {
		// Don't check for cancel or wait on first iteration
		if !first {
			select {
			case <-ctx.Done():
				log.Println("[INFO] WinRM wait cancelled. Exiting loop.")
				return nil, errors.New("WinRM wait cancelled")
			case <-time.After(5 * time.Second):
			}
		}
		first = false

		host, err := s.Host(state)
		if err != nil {
			log.Printf("[DEBUG] Error getting WinRM host: %s", err)
			continue
		}
		s.Config.WinRMHost = host

		port := s.Config.WinRMPort
		if s.WinRMPort != nil {
			port, err = s.WinRMPort(state)
			if err != nil {
				log.Printf("[DEBUG] Error getting WinRM port: %s", err)
				continue
			}
			s.Config.WinRMPort = port
		}

		state.Put("communicator_config", s.Config)

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
	// run an "echo" command to make sure winrm is actually connected before moving on.
	var connectCheckCommand = winrmcmd.Powershell(`if (Test-Path variable:global:ProgressPreference){$ProgressPreference='SilentlyContinue'}; echo "WinRM connected."`)
	var retryableSleep = 5 * time.Second
	// run an "echo" command to make sure that the winrm is connected
	for {
		cmd := &packer.RemoteCmd{Command: connectCheckCommand}
		var buf, buf2 bytes.Buffer
		cmd.Stdout = &buf
		cmd.Stdout = io.MultiWriter(cmd.Stdout, &buf2)
		select {
		case <-ctx.Done():
			log.Println("WinRM wait canceled, exiting loop")
			return comm, fmt.Errorf("WinRM wait canceled")
		case <-time.After(retryableSleep):
		}

		log.Printf("Checking that WinRM is connected with: '%s'", connectCheckCommand)
		ui := state.Get("ui").(packer.Ui)
		err := cmd.RunWithUi(ctx, comm, ui)

		if err != nil {
			log.Printf("Communication connection err: %s", err)
			continue
		}

		log.Printf("Connected to machine")
		stdoutToRead := buf2.String()
		if !strings.Contains(stdoutToRead, "WinRM connected.") {
			log.Printf("echo didn't succeed; retrying...")
			continue
		}
		break
	}

	return comm, nil
}
