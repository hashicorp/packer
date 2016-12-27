// This package implements a provisioner for Packer that executes
// Converge to provision a remote machine

package converge

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"

	"strings"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

// Config for Converge provisioner
type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Bootstrapping
	NoBootstrap bool `mapstructure:"no_bootstrap"` // TODO: add a way to specify bootstrap version

	// Modules
	ModuleDirs []ModuleDir `mapstructure:"module_dirs"`

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
		},
		raws...,
	)

	return err
}

// Provision node somehow. TODO: actual docs
func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	ui.Say("Provisioning with Converge")

	// bootstrapping
	if err := p.maybeBootstrap(ui, comm); err != nil {
		return err // error messages are already user-friendly
	}

	// check version (really, this make sure that Converge is installed before we try to run it)
	if err := p.checkVersion(ui, comm); err != nil {
		return err // error messages are already user-friendly
	}

	// send module directories to the remote host
	if err := p.sendModuleDirectories(ui, comm); err != nil {
		return err // error messages are already user-friendly
	}

	return nil
}

func (p *Provisioner) maybeBootstrap(ui packer.Ui, comm packer.Communicator) error {
	if p.config.NoBootstrap {
		return nil
	}
	ui.Message("bootstrapping converge")

	bootstrap, err := http.Get("https://get.converge.sh")
	defer bootstrap.Body.Close()
	if err != nil {
		return fmt.Errorf("Error downloading bootstrap script: %s", err) // TODO: is github.com/pkg/error allowed?
	}
	if err := comm.Upload("/tmp/install-converge.sh", bootstrap.Body, nil); err != nil {
		return fmt.Errorf("Error uploading script: %s", err)
	}

	var out bytes.Buffer
	cmd := &packer.RemoteCmd{
		Command: "/bin/sh /tmp/install-converge.sh",
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

func (p *Provisioner) checkVersion(ui packer.Ui, comm packer.Communicator) error {
	var versionOut bytes.Buffer
	cmd := &packer.RemoteCmd{
		Command: "converge version",
		Stdin:   nil,
		Stdout:  &versionOut,
		Stderr:  &versionOut,
	}
	if err := comm.Start(cmd); err != nil || cmd.ExitStatus != 0 {
		return fmt.Errorf("Error running `converge version`: %s", err)
	}

	cmd.Wait()
	if cmd.ExitStatus == 127 {
		ui.Error("Could not determine Converge version. Is it installed and in PATH?")
		if p.config.NoBootstrap {
			ui.Error("Bootstrapping was disabled for this run. That might be why Converge isn't present.")
		}

		return errors.New("could not determine Converge version")

	} else if cmd.ExitStatus != 0 {
		ui.Error(versionOut.String())
		ui.Error(fmt.Sprintf("exited with error code %d", cmd.ExitStatus))
		return errors.New("Error running `converge version`")
	}

	ui.Say(fmt.Sprintf("Provisioning with %s", strings.TrimSpace(versionOut.String())))

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

// Cancel the provisioning process
func (p *Provisioner) Cancel() {
	log.Println("cancel called in Converge provisioner")
}
