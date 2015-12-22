package lxc

import (
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"strings"
	"time"
)

type StepWaitInit struct {
	WaitTimeout time.Duration
}

func (s *StepWaitInit) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	var err error

	cancel := make(chan struct{})
	waitDone := make(chan bool, 1)
	go func() {
		ui.Say("Waiting for container to finish init...")
		err = s.waitForInit(state, cancel)
		waitDone <- true
	}()

	log.Printf("Waiting for container to finish init, up to timeout: %s", s.WaitTimeout)
	timeout := time.After(s.WaitTimeout)
WaitLoop:
	for {
		select {
		case <-waitDone:
			if err != nil {
				ui.Error(fmt.Sprintf("Error waiting for container to finish init: %s", err))
				return multistep.ActionHalt
			}

			ui.Say("Container finished init!")
			break WaitLoop
		case <-timeout:
			err := fmt.Errorf("Timeout waiting for container to finish init.")
			state.Put("error", err)
			ui.Error(err.Error())
			close(cancel)
			return multistep.ActionHalt
		case <-time.After(1 * time.Second):
			if _, ok := state.GetOk(multistep.StateCancelled); ok {
				close(cancel)
				log.Println("Interrupt detected, quitting waiting for container to finish init.")
				return multistep.ActionHalt
			}
		}
	}

	return multistep.ActionContinue
}

func (s *StepWaitInit) Cleanup(multistep.StateBag) {
}

func (s *StepWaitInit) waitForInit(state multistep.StateBag, cancel <-chan struct{}) error {
	config := state.Get("config").(*Config)
	mountPath := state.Get("mount_path").(string)
	wrappedCommand := state.Get("wrappedCommand").(CommandWrapper)

	for {
		select {
		case <-cancel:
			log.Println("Cancelled. Exiting loop.")
			return errors.New("Wait cancelled")
		case <-time.After(1 * time.Second):
		}

		comm := &LxcAttachCommunicator{
			ContainerName: config.ContainerName,
			RootFs:        mountPath,
			CmdWrapper:    wrappedCommand,
		}

		runlevel, _ := comm.CheckInit()
		currentRunlevel := "unknown"
		if arr := strings.Split(runlevel, " "); len(arr) >= 2 {
			currentRunlevel = arr[1]
		}

		log.Printf("Current runlevel in container: '%s'", runlevel)

		targetRunlevel := fmt.Sprintf("%d", config.TargetRunlevel)
		if currentRunlevel == targetRunlevel {
			log.Printf("Container finished init.")
			break
		}

		/*log.Println("Attempting SSH connection...")
		comm, err = ssh.New(config)
		if err != nil {
			log.Printf("SSH handshake err: %s", err)

			// Only count this as an attempt if we were able to attempt
			// to authenticate. Note this is very brittle since it depends
			// on the string of the error... but I don't see any other way.
			if strings.Contains(err.Error(), "authenticate") {
				log.Printf("Detected authentication error. Increasing handshake attempts.")
				handshakeAttempts += 1
			}

			if handshakeAttempts < 10 {
				// Try to connect via SSH a handful of times
				continue
			}

			return nil, err
		}

		break
		*/
	}

	return nil
}
