// This package implements a provisioner for Packer that executes
// Converge to provision a remote machine

package converge

import (
	"bytes"
	"errors"
	"fmt"

	"strings"

	"encoding/json"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

// Config for Converge provisioner
type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Bootstrapping
	SkipBootstrap    bool   `mapstructure:"skip_bootstrap"`
	Version          string `mapstructure:"version"`
	BootstrapCommand string `mapstructure:"bootstrap_command"`

	// Modules
	ModuleDirs []ModuleDir `mapstructure:"module_dirs"`

	// Execution
	Module           string            `mapstructure:"module"`
	WorkingDirectory string            `mapstructure:"working_directory"`
	Params           map[string]string `mapstucture:"params"`
	ExecuteCommand   string            `mapstructure:"execute_command"`
	PreventSudo      bool              `mapstructure:"prevent_sudo"`

	ctx interpolate.Context
}

// ModuleDir is a directory to transfer to the remote system
type ModuleDir struct {
	Source      string   `mapstructure:"source"`
	Destination string   `mapstructure:"destination"`
	Exclude     []string `mapstructure:"exclude"`
}

// Provisioner for Converge
type Provisioner struct {
	config Config
}

// Prepare provisioner somehow. TODO: actual docs
func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(
		&p.config,
		&config.DecodeOpts{
			Interpolate:        true,
			InterpolateContext: &p.config.ctx,
			InterpolateFilter: &interpolate.RenderFilter{
				Exclude: []string{
					"execute_command",
					"bootstrap_command",
				},
			},
		},
		raws...,
	)
	if err != nil {
		return err
	}

	// require a single module
	if p.config.Module == "" {
		return errors.New("Converge requires a module to provision the system")
	}

	// set defaults
	if p.config.WorkingDirectory == "" {
		p.config.WorkingDirectory = "/tmp"
	}

	if p.config.ExecuteCommand == "" {
		p.config.ExecuteCommand = "cd {{.WorkingDirectory}} && {{if .Sudo}}sudo {{end}}converge apply --local --log-level=WARNING --paramsJSON '{{.ParamsJSON}}' {{.Module}}"
	}

	if p.config.BootstrapCommand == "" {
		p.config.BootstrapCommand = "curl -s https://get.converge.sh | sh {{if ne .Version \"\"}}-s -- -v {{.Version}}{{end}}"
	}

	// validate sources and destinations
	for i, dir := range p.config.ModuleDirs {
		if dir.Source == "" {
			return fmt.Errorf("Source (\"source\" key) is required in Converge module dir #%d", i)
		}
		if dir.Destination == "" {
			return fmt.Errorf("Destination (\"destination\" key) is required in Converge module dir #%d", i)
		}
	}

	return err
}

// Provision node somehow. TODO: actual docs
func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	ui.Say("Provisioning with Converge")

	// bootstrapping
	if err := p.maybeBootstrap(ui, comm); err != nil {
		return err // error messages are already user-friendly
	}

	// send module directories to the remote host
	if err := p.sendModuleDirectories(ui, comm); err != nil {
		return err // error messages are already user-friendly
	}

	// apply all the modules
	if err := p.applyModules(ui, comm); err != nil {
		return err // error messages are already user-friendly
	}

	return nil
}

func (p *Provisioner) maybeBootstrap(ui packer.Ui, comm packer.Communicator) error {
	if p.config.SkipBootstrap {
		return nil
	}
	ui.Message("bootstrapping converge")

	p.config.ctx.Data = struct {
		Version string
	}{
		Version: p.config.Version,
	}
	command, err := interpolate.Render(p.config.BootstrapCommand, &p.config.ctx)
	if err != nil {
		return fmt.Errorf("Could not interpolate bootstrap command: %s", err)
	}

	var out bytes.Buffer
	cmd := &packer.RemoteCmd{
		Command: command,
		Stdin:   nil,
		Stdout:  &out,
		Stderr:  &out,
	}

	if err = comm.Start(cmd); err != nil {
		return fmt.Errorf("Error bootstrapping converge: %s", err)
	}

	cmd.Wait()
	if cmd.ExitStatus != 0 {
		ui.Error(out.String())
		return errors.New("Error bootstrapping converge")
	}

	ui.Message(strings.TrimSpace(out.String()))
	return nil
}

func (p *Provisioner) sendModuleDirectories(ui packer.Ui, comm packer.Communicator) error {
	for _, dir := range p.config.ModuleDirs {
		if err := comm.UploadDir(dir.Destination, dir.Source, dir.Exclude); err != nil {
			return fmt.Errorf("Could not upload %q: %s", dir.Source, err)
		}
		ui.Message(fmt.Sprintf("transferred %q to %q", dir.Source, dir.Destination))
	}

	return nil
}

func (p *Provisioner) applyModules(ui packer.Ui, comm packer.Communicator) error {
	// create params JSON file
	params, err := json.Marshal(p.config.Params)
	if err != nil {
		return fmt.Errorf("Could not marshal parameters as JSON: %s", err)
	}

	p.config.ctx.Data = struct {
		ParamsJSON, WorkingDirectory, Module string
		Sudo                                 bool
	}{
		ParamsJSON:       string(params),
		WorkingDirectory: p.config.WorkingDirectory,
		Module:           p.config.Module,
		Sudo:             !p.config.PreventSudo,
	}
	command, err := interpolate.Render(p.config.ExecuteCommand, &p.config.ctx)
	if err != nil {
		return fmt.Errorf("Could not interpolate execute command: %s", err)
	}

	// run Converge in the specified directory
	var runOut bytes.Buffer
	cmd := &packer.RemoteCmd{
		Command: command,
		Stdin:   nil,
		Stdout:  &runOut,
		Stderr:  &runOut,
	}
	if err := comm.Start(cmd); err != nil {
		return fmt.Errorf("Error applying %q: %s", p.config.Module, err)
	}

	cmd.Wait()
	if cmd.ExitStatus == 127 {
		ui.Error("Could not find Converge. Is it installed and in PATH?")
		if p.config.SkipBootstrap {
			ui.Error("Bootstrapping was disabled for this run. That might be why Converge isn't present.")
		}

		return errors.New("Could not find Converge")

	} else if cmd.ExitStatus != 0 {
		ui.Error(strings.TrimSpace(runOut.String()))
		ui.Error(fmt.Sprintf("exited with error code %d", cmd.ExitStatus))
		return fmt.Errorf("Error applying %q", p.config.Module)
	}

	ui.Message(strings.TrimSpace(runOut.String()))

	return nil
}

// Cancel the provisioning process
func (p *Provisioner) Cancel() {
	// there's not an awful lot we can do to cancel Converge at the moment.
	// The default semantics are fine.
}
