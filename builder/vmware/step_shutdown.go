package vmware

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// This step shuts down the machine. It first attempts to do so gracefully,
// but ultimately forcefully shuts it down if that fails.
//
// Uses:
//   communicator packer.Communicator
//   config *config
//   driver Driver
//   ui     packer.Ui
//   vmx_path string
//
// Produces:
//   <nothing>
type stepShutdown struct{}

func (s *stepShutdown) Run(state map[string]interface{}) multistep.StepAction {
	comm := state["communicator"].(packer.Communicator)
	config := state["config"].(*config)
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)
	vmxPath := state["vmx_path"].(string)

	if config.ShutdownCommand != "" {
		ui.Say("Gracefully halting virtual machine...")
		log.Printf("Executing shutdown command: %s", config.ShutdownCommand)

		var stdout, stderr bytes.Buffer
		cmd := &packer.RemoteCmd{
			Command: config.ShutdownCommand,
			Stdout:  &stdout,
			Stderr:  &stderr,
		}
		if err := comm.Start(cmd); err != nil {
			err := fmt.Errorf("Failed to send shutdown command: %s", err)
			state["error"] = err
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		// Wait for the command to run
		cmd.Wait()

		// If the command failed to run, notify the user in some way.
		if cmd.ExitStatus != 0 {
			state["error"] = fmt.Errorf(
				"Shutdown command has non-zero exit status.\n\nStdout: %s\n\nStderr: %s",
				stdout.String(), stderr.String())
			return multistep.ActionHalt
		}

		log.Printf("Shutdown stdout: %s", stdout.String())
		log.Printf("Shutdown stderr: %s", stderr.String())

		// Wait for the machine to actually shut down
		log.Printf("Waiting max %s for shutdown to complete", config.shutdownTimeout)
		shutdownTimer := time.After(config.shutdownTimeout)
		for {
			running, _ := driver.IsRunning(vmxPath)
			if !running {
				break
			}

			select {
			case <-shutdownTimer:
				err := errors.New("Timeout while waiting for machine to shut down.")
				state["error"] = err
				ui.Error(err.Error())
				return multistep.ActionHalt
			default:
				time.Sleep(1 * time.Second)
			}
		}
	} else {
		if err := driver.Stop(vmxPath); err != nil {
			err := fmt.Errorf("Error stopping VM: %s", err)
			state["error"] = err
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	ui.Message("Waiting for VMware to clean up after itself...")
	lockPattern := filepath.Join(config.OutputDir, "*.lck")
	timer := time.After(15 * time.Second)
LockWaitLoop:
	for {
		locks, err := filepath.Glob(lockPattern)
		if err == nil {
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
		case <-time.After(1 * time.Second):
		}
	}

	if runtime.GOOS == "windows" {
		// Windows takes a while to yield control of the files when the
		// process is exiting. We just sleep here. In the future, it'd be
		// nice to find a better solution to this.
		time.Sleep(5 * time.Second)
	}

	log.Println("VM shut down.")
	return multistep.ActionContinue
}

func (s *stepShutdown) Cleanup(state map[string]interface{}) {}
