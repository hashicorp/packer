package common

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// This step shuts down the machine. It first attempts to do so gracefully,
// but ultimately forcefully shuts it down if that fails.
//
// Uses:
//   communicator packer.Communicator
//   dir OutputDir
//   driver Driver
//   ui     packer.Ui
//   vmx_path string
//
// Produces:
//   <nothing>
type StepShutdown struct {
	Command string
	Timeout time.Duration

	// Set this to true if we're testing
	Testing bool
}

func (s *StepShutdown) Run(state multistep.StateBag) multistep.StepAction {
	comm := state.Get("communicator").(packer.Communicator)
	dir := state.Get("dir").(OutputDir)
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	vmxPath := state.Get("vmx_path").(string)

	if s.Command != "" {
		ui.Say("Gracefully halting virtual machine...")
		log.Printf("Executing shutdown command: %s", s.Command)

		var stdout, stderr bytes.Buffer
		cmd := &packer.RemoteCmd{
			Command: s.Command,
			Stdout:  &stdout,
			Stderr:  &stderr,
		}
		if err := comm.Start(cmd); err != nil {
			err := fmt.Errorf("Failed to send shutdown command: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Wait for the command to run
		cmd.Wait()

		// If the command failed to run, notify the user in some way.
		if cmd.ExitStatus != 0 {
			state.Put("error", fmt.Errorf(
				"Shutdown command has non-zero exit status.\n\nStdout: %s\n\nStderr: %s",
				stdout.String(), stderr.String()))
			return multistep.ActionHalt
		}

		log.Printf("Shutdown stdout: %s", stdout.String())
		log.Printf("Shutdown stderr: %s", stderr.String())

		// Wait for the machine to actually shut down
		log.Printf("Waiting max %s for shutdown to complete", s.Timeout)
		shutdownTimer := time.After(s.Timeout)
		for {
			running, _ := driver.IsRunning(vmxPath)
			if !running {
				break
			}

			select {
			case <-shutdownTimer:
				err := errors.New("Timeout while waiting for machine to shut down.")
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			default:
				time.Sleep(150 * time.Millisecond)
			}
		}
	} else {
		ui.Say("Forcibly halting virtual machine...")
		if err := driver.Stop(vmxPath); err != nil {
			err := fmt.Errorf("Error stopping VM: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	ui.Message("Waiting for VMware to clean up after itself...")
	lockRegex := regexp.MustCompile(`(?i)\.lck$`)
	timer := time.After(120 * time.Second)
LockWaitLoop:
	for {
		files, err := dir.ListFiles()
		if err != nil {
			log.Printf("Error listing files in outputdir: %s", err)
		} else {
			var locks []string
			for _, file := range files {
				if lockRegex.MatchString(file) {
					locks = append(locks, file)
				}
			}

			if len(locks) == 0 {
				log.Println("No more lock files found. VMware is clean.")
				break
			}

			if len(locks) == 1 && strings.HasSuffix(locks[0], ".vmx.lck") {
				log.Println("Only waiting on VMX lock. VMware is clean.")
				break
			}

			log.Printf("Waiting on lock files: %#v", locks)
		}

		select {
		case <-timer:
			log.Println("Reached timeout on waiting for clean VMware. Assuming clean.")
			break LockWaitLoop
		case <-time.After(150 * time.Millisecond):
		}
	}

	if runtime.GOOS != "darwin" && !s.Testing {
		// Windows takes a while to yield control of the files when the
		// process is exiting. Ubuntu will yield control of the files but
		// VMWare may overwrite the VMX cleanup steps that run after this,
		// so we wait to ensure VMWare has exited and flushed the VMX.

		// We just sleep here. In the future, it'd be nice to find a better
		// solution to this.
		time.Sleep(5 * time.Second)
	}

	log.Println("VM shut down.")
	return multistep.ActionContinue
}

func (s *StepShutdown) Cleanup(state multistep.StateBag) {}
