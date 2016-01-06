package restart

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/masterzen/winrm/winrm"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

var DefaultRestartCommand = "shutdown /r /f /t 0 /c \"packer restart\""
var DefaultRestartCheckCommand = winrm.Powershell(`echo "${env:COMPUTERNAME} restarted."`)
var retryableSleep = 5 * time.Second
var TryCheckReboot = "shutdown.exe -f -r -t 60"
var AbortReboot = "shutdown.exe -a"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// The command used to restart the guest machine
	RestartCommand string `mapstructure:"restart_command"`

	// The command used to check if the guest machine has restarted
	// The output of this command will be displayed to the user
	RestartCheckCommand string `mapstructure:"restart_check_command"`

	// The timeout for waiting for the machine to restart
	RestartTimeout time.Duration `mapstructure:"restart_timeout"`

	ctx interpolate.Context
}

type Provisioner struct {
	config     Config
	comm       packer.Communicator
	ui         packer.Ui
	cancel     chan struct{}
	cancelLock sync.Mutex
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"execute_command",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	if p.config.RestartCommand == "" {
		p.config.RestartCommand = DefaultRestartCommand
	}

	if p.config.RestartCheckCommand == "" {
		p.config.RestartCheckCommand = DefaultRestartCheckCommand
	}

	if p.config.RestartTimeout == 0 {
		p.config.RestartTimeout = 5 * time.Minute
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	p.cancelLock.Lock()
	p.cancel = make(chan struct{})
	p.cancelLock.Unlock()

	ui.Say("Restarting Machine")
	p.comm = comm
	p.ui = ui

	var cmd *packer.RemoteCmd
	command := p.config.RestartCommand
	err := p.retryable(func() error {
		cmd = &packer.RemoteCmd{Command: command}
		return cmd.StartWithUi(comm, ui)
	})

	if err != nil {
		return err
	}

	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Restart script exited with non-zero exit status: %d", cmd.ExitStatus)
	}

	return waitForRestart(p, comm)
}

var waitForRestart = func(p *Provisioner, comm packer.Communicator) error {
	ui := p.ui
	ui.Say("Waiting for machine to restart...")
	waitDone := make(chan bool, 1)
	timeout := time.After(p.config.RestartTimeout)
	var err error

	p.comm = comm
	var cmd *packer.RemoteCmd
	trycommand := TryCheckReboot
	abortcommand := AbortReboot
	// Stolen from Vagrant reboot checker
	for {
		log.Printf("Check if machine is rebooting...")
		cmd = &packer.RemoteCmd{Command: trycommand}
		err = cmd.StartWithUi(comm, ui)
		if err != nil {
			// Couldnt execute, we asume machine is rebooting already
			break
		}
		if cmd.ExitStatus == 1115 || cmd.ExitStatus == 1190 {
			// Reboot already in progress but not completed
			log.Printf("Reboot already in progress, waiting...")
			time.Sleep(10 * time.Second)
		}
		if cmd.ExitStatus == 0 {
			// Cancel reboot we created to test if machine was already rebooting
			cmd = &packer.RemoteCmd{Command: abortcommand}
			cmd.StartWithUi(comm, ui)
			break
		}
	}

	go func() {
		log.Printf("Waiting for machine to become available...")
		err = waitForCommunicator(p)
		waitDone <- true
	}()

	log.Printf("Waiting for machine to reboot with timeout: %s", p.config.RestartTimeout)

WaitLoop:
	for {
		// Wait for either WinRM to become available, a timeout to occur,
		// or an interrupt to come through.
		select {
		case <-waitDone:
			if err != nil {
				ui.Error(fmt.Sprintf("Error waiting for WinRM: %s", err))
				return err
			}

			ui.Say("Machine successfully restarted, moving on")
			close(p.cancel)
			break WaitLoop
		case <-timeout:
			err := fmt.Errorf("Timeout waiting for WinRM.")
			ui.Error(err.Error())
			close(p.cancel)
			return err
		case <-p.cancel:
			close(waitDone)
			return fmt.Errorf("Interrupt detected, quitting waiting for machine to restart")
		}
	}

	return nil

}

var waitForCommunicator = func(p *Provisioner) error {
	cmd := &packer.RemoteCmd{Command: p.config.RestartCheckCommand}

	for {
		select {
		case <-p.cancel:
			log.Println("Communicator wait cancelled, exiting loop")
			return fmt.Errorf("Communicator wait cancelled")
		case <-time.After(retryableSleep):
		}

		log.Printf("Attempting to communicator to machine with: '%s'", cmd.Command)

		err := cmd.StartWithUi(p.comm, p.ui)
		if err != nil {
			log.Printf("Communication connection err: %s", err)
			continue
		}

		log.Printf("Connected to machine")
		break
	}

	return nil
}

func (p *Provisioner) Cancel() {
	log.Printf("Received interrupt Cancel()")

	p.cancelLock.Lock()
	defer p.cancelLock.Unlock()
	if p.cancel != nil {
		close(p.cancel)
	}
}

// retryable will retry the given function over and over until a
// non-error is returned.
func (p *Provisioner) retryable(f func() error) error {
	startTimeout := time.After(p.config.RestartTimeout)
	for {
		var err error
		if err = f(); err == nil {
			return nil
		}

		// Create an error and log it
		err = fmt.Errorf("Retryable error: %s", err)
		log.Printf(err.Error())

		// Check if we timed out, otherwise we retry. It is safe to
		// retry since the only error case above is if the command
		// failed to START.
		select {
		case <-startTimeout:
			return err
		default:
			time.Sleep(retryableSleep)
		}
	}
}
