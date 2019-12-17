//go:generate mapstructure-to-hcl2 -type Config

package restart

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"log"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/masterzen/winrm"
)

var DefaultRestartCommand = `shutdown /r /f /t 0 /c "packer restart"`
var DefaultRestartCheckCommand = winrm.Powershell(`echo "${env:COMPUTERNAME} restarted."`)
var retryableSleep = 5 * time.Second
var TryCheckReboot = `shutdown /r /f /t 60 /c "packer restart test"`
var AbortReboot = `shutdown /a`

var DefaultRegistryKeys = []string{
	"HKLM:SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Component Based Servicing\\RebootPending",
	"HKLM:SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Component Based Servicing\\PackagesPending",
	"HKLM:Software\\Microsoft\\Windows\\CurrentVersion\\Component Based Servicing\\RebootInProgress",
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// The command used to restart the guest machine
	RestartCommand string `mapstructure:"restart_command"`

	// The command used to check if the guest machine has restarted
	// The output of this command will be displayed to the user
	RestartCheckCommand string `mapstructure:"restart_check_command"`

	// The timeout for waiting for the machine to restart
	RestartTimeout time.Duration `mapstructure:"restart_timeout"`

	// Whether to check the registry (see RegistryKeys) for pending reboots
	CheckKey bool `mapstructure:"check_registry"`

	// custom keys to check for
	RegistryKeys []string `mapstructure:"registry_keys"`

	ctx interpolate.Context
}

type Provisioner struct {
	config     Config
	comm       packer.Communicator
	ui         packer.Ui
	cancel     chan struct{}
	cancelLock sync.Mutex
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

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

	if len(p.config.RegistryKeys) == 0 {
		p.config.RegistryKeys = DefaultRegistryKeys
	}

	return nil
}

func (p *Provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator) error {
	p.cancelLock.Lock()
	p.cancel = make(chan struct{})
	p.cancelLock.Unlock()

	ui.Say("Restarting Machine")
	p.comm = comm
	p.ui = ui

	var cmd *packer.RemoteCmd
	command := p.config.RestartCommand
	err := retry.Config{StartTimeout: p.config.RestartTimeout}.Run(ctx, func(context.Context) error {
		cmd = &packer.RemoteCmd{Command: command}
		return cmd.RunWithUi(ctx, comm, ui)
	})

	if err != nil {
		return err
	}

	if cmd.ExitStatus() != 0 && cmd.ExitStatus() != 1115 && cmd.ExitStatus() != 1190 {
		return fmt.Errorf("Restart script exited with non-zero exit status: %d", cmd.ExitStatus())
	}

	return waitForRestart(ctx, p, comm)
}

var waitForRestart = func(ctx context.Context, p *Provisioner, comm packer.Communicator) error {
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
		err = cmd.RunWithUi(ctx, comm, ui)
		if err != nil {
			// Couldn't execute, we assume machine is rebooting already
			break
		}
		if cmd.ExitStatus() == 1 {
			// SSH provisioner, and we're already rebooting. SSH can reconnect
			// without our help; exit this wait loop.
			break
		}
		if cmd.ExitStatus() == 1115 || cmd.ExitStatus() == 1190 || cmd.ExitStatus() == 1717 {
			// Reboot already in progress but not completed
			log.Printf("Reboot already in progress, waiting...")
			time.Sleep(10 * time.Second)
		}
		if cmd.ExitStatus() == 0 {
			// Cancel reboot we created to test if machine was already rebooting
			cmd = &packer.RemoteCmd{Command: abortcommand}
			cmd.RunWithUi(ctx, comm, ui)
			break
		}
	}

	go func() {
		log.Printf("Waiting for machine to become available...")
		err = waitForCommunicator(ctx, p)
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
				ui.Error(fmt.Sprintf("Error waiting for machine to restart: %s", err))
				return err
			}

			ui.Say("Machine successfully restarted, moving on")
			close(p.cancel)
			break WaitLoop
		case <-timeout:
			err := fmt.Errorf("Timeout waiting for machine to restart.")
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

var waitForCommunicator = func(ctx context.Context, p *Provisioner) error {
	runCustomRestartCheck := true
	if p.config.RestartCheckCommand == DefaultRestartCheckCommand {
		runCustomRestartCheck = false
	}
	// This command is configurable by the user to make sure that the
	// vm has met their necessary criteria for having restarted. If the
	// user doesn't set a special restart command, we just run the
	// default as cmdModuleLoad below.
	cmdRestartCheck := &packer.RemoteCmd{Command: p.config.RestartCheckCommand}
	log.Printf("Checking that communicator is connected with: '%s'",
		cmdRestartCheck.Command)
	for {
		select {
		case <-ctx.Done():
			log.Println("Communicator wait canceled, exiting loop")
			return fmt.Errorf("Communicator wait canceled")
		case <-time.After(retryableSleep):
		}
		if runCustomRestartCheck {
			// run user-configured restart check
			err := cmdRestartCheck.RunWithUi(ctx, p.comm, p.ui)
			if err != nil {
				log.Printf("Communication connection err: %s", err)
				continue
			}
			log.Printf("Connected to machine")
			runCustomRestartCheck = false
		}
		// This is the non-user-configurable check that powershell
		// modules have loaded.

		// If we catch the restart in just the right place, we will be able
		// to run the restart check but the output will be an error message
		// about how it needs powershell modules to load, and we will start
		// provisioning before powershell is actually ready.
		// In this next check, we parse stdout to make sure that the command is
		// actually running as expected.
		cmdModuleLoad := &packer.RemoteCmd{Command: DefaultRestartCheckCommand}
		var buf, buf2 bytes.Buffer
		cmdModuleLoad.Stdout = &buf
		cmdModuleLoad.Stdout = io.MultiWriter(cmdModuleLoad.Stdout, &buf2)

		cmdModuleLoad.RunWithUi(ctx, p.comm, p.ui)
		stdoutToRead := buf2.String()

		if !strings.Contains(stdoutToRead, "restarted.") {
			log.Printf("echo didn't succeed; retrying...")
			continue
		}

		if p.config.CheckKey {
			log.Printf("Connected to machine")
			shouldContinue := false
			for _, RegKey := range p.config.RegistryKeys {
				KeyTestCommand := winrm.Powershell(fmt.Sprintf(`Test-Path "%s"`, RegKey))
				cmdKeyCheck := &packer.RemoteCmd{Command: KeyTestCommand}
				log.Printf("Checking registry for pending reboots")
				var buf, buf2 bytes.Buffer
				cmdKeyCheck.Stdout = &buf
				cmdKeyCheck.Stdout = io.MultiWriter(cmdKeyCheck.Stdout, &buf2)

				err := p.comm.Start(ctx, cmdKeyCheck)
				if err != nil {
					log.Printf("Communication connection err: %s", err)
					shouldContinue = true
				}
				cmdKeyCheck.Wait()

				stdoutToRead := buf2.String()
				if strings.Contains(stdoutToRead, "True") {
					log.Printf("RegistryKey %s exists; waiting...", KeyTestCommand)
					shouldContinue = true
				} else {
					log.Printf("No Registry keys found; exiting wait loop")
				}
			}
			if shouldContinue {
				continue
			}
		}
		break
	}

	return nil
}
