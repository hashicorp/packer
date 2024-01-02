// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

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
	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/retry"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/masterzen/winrm"
)

var DefaultRestartCommand = `shutdown /r /f /t 0 /c "packer restart"`
var DefaultRestartCheckCommand = winrm.Powershell(`echo ("{0} restarted." -f [System.Net.Dns]::GetHostName())`)
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

	// The command to run after executing `restart_command` to check if the guest machine has restarted.
	// This command will retry until the connection to the guest machine has been restored or `restart_timeout` has exceeded.
	// The output of this command will be displayed to the user.
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
	comm       packersdk.Communicator
	ui         packersdk.Ui
	cancel     chan struct{}
	cancelLock sync.Mutex
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         "windows-restart",
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

func (p *Provisioner) Provision(ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator, _ map[string]interface{}) error {
	p.cancelLock.Lock()
	p.cancel = make(chan struct{})
	p.cancelLock.Unlock()

	ui.Say("Restarting Machine")
	p.comm = comm
	p.ui = ui

	var cmd *packersdk.RemoteCmd
	command := p.config.RestartCommand
	err := retry.Config{StartTimeout: p.config.RestartTimeout}.Run(ctx, func(context.Context) error {
		cmd = &packersdk.RemoteCmd{Command: command}
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

var waitForRestart = func(ctx context.Context, p *Provisioner, comm packersdk.Communicator) error {
	ui := p.ui
	ui.Say("Waiting for machine to restart...")
	waitDone := make(chan bool, 1)
	timeout := time.After(p.config.RestartTimeout)
	var err error

	p.comm = comm
	var cmd *packersdk.RemoteCmd
	trycommand := TryCheckReboot
	abortcommand := AbortReboot

	// Stolen from Vagrant reboot checker
	for {
		log.Printf("Check if machine is rebooting...")
		cmd = &packersdk.RemoteCmd{Command: trycommand}
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
			cmd = &packersdk.RemoteCmd{Command: abortcommand}
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
	cmdRestartCheck := &packersdk.RemoteCmd{Command: p.config.RestartCheckCommand}
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
		cmdModuleLoad := &packersdk.RemoteCmd{Command: DefaultRestartCheckCommand}
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
				cmdKeyCheck := &packersdk.RemoteCmd{Command: KeyTestCommand}
				log.Printf("Checking registry for pending reboots")
				var buf, buf2 bytes.Buffer
				cmdKeyCheck.Stdout = &buf
				cmdKeyCheck.Stdout = io.MultiWriter(cmdKeyCheck.Stdout, &buf2)

				err := cmdKeyCheck.RunWithUi(ctx, p.comm, p.ui)
				if err != nil {
					log.Printf("Communication connection err: %s", err)
					shouldContinue = true
				}

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
