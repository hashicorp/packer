package yandexexport

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/yandex"
)

type StepPrepareTools struct{}

// Run reads the instance metadata and looks for the log entry
// indicating the cloud-init script finished.
func (s *StepPrepareTools) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	comm := state.Get("communicator").(packersdk.Communicator)
	pkgManager, errPkgManager := detectPkgManager(ctx, comm)

	if which(ctx, comm, "qemu-img") != nil {
		if errPkgManager != nil {
			return yandex.StepHaltWithError(state, errPkgManager)
		}
		ui.Message("Install qemu-img...")
		if err := pkgManager.InstallQemuIMG(ctx, comm); err != nil {
			return yandex.StepHaltWithError(state, err)
		}
	}
	if which(ctx, comm, "aws") != nil {
		if errPkgManager != nil {
			return yandex.StepHaltWithError(state, errPkgManager)
		}
		ui.Message("Install aws...")
		if err := pkgManager.InstallAWS(ctx, comm); err != nil {
			return yandex.StepHaltWithError(state, err)
		}
	}

	return multistep.ActionContinue
}

// Cleanup nothing
func (s *StepPrepareTools) Cleanup(state multistep.StateBag) {}

func detectPkgManager(ctx context.Context, comm packersdk.Communicator) (pkgManager, error) {
	if err := which(ctx, comm, "apt"); err == nil {
		return &apt{}, nil
	}
	if err := which(ctx, comm, "yum"); err == nil {
		return &yum{}, nil
	}

	return nil, fmt.Errorf("Cannot detect package manager")
}

func which(ctx context.Context, comm packersdk.Communicator, what string) error {
	cmdCheckAPT := &packersdk.RemoteCmd{
		Command: fmt.Sprintf("which %s", what),
	}
	if err := comm.Start(ctx, cmdCheckAPT); err != nil {
		return err
	}
	if cmdCheckAPT.Wait() == 0 {
		return nil
	}
	return fmt.Errorf("Not found: %s", what)
}

type pkgManager interface {
	InstallQemuIMG(ctx context.Context, comm packersdk.Communicator) error
	InstallAWS(ctx context.Context, comm packersdk.Communicator) error
}

type apt struct {
	updated bool
}

func (p *apt) InstallAWS(ctx context.Context, comm packersdk.Communicator) error {
	if err := p.Update(ctx, comm); err != nil {
		return err
	}
	if err := execCMDWithSudo(ctx, comm, "apt install -y awscli"); err != nil {
		return fmt.Errorf("Cannot install awscli")
	}
	return nil
}

func (p *apt) InstallQemuIMG(ctx context.Context, comm packersdk.Communicator) error {
	if err := p.Update(ctx, comm); err != nil {
		return err
	}
	if err := execCMDWithSudo(ctx, comm, "apt install -y qemu-utils"); err != nil {
		return fmt.Errorf("Cannot install qemu-utils")
	}
	return nil
}
func (p *apt) Update(ctx context.Context, comm packersdk.Communicator) error {
	if p.updated {
		return nil
	}
	if err := execCMDWithSudo(ctx, comm, "apt update"); err != nil {
		return fmt.Errorf("Cannot update: %s", err)
	}
	p.updated = true
	return nil
}

type yum struct{}

func (p *yum) InstallAWS(ctx context.Context, comm packersdk.Communicator) error {
	if which(ctx, comm, "pip3") != nil {
		if err := execCMDWithSudo(ctx, comm, "yum install -y python3-pip"); err != nil {
			return fmt.Errorf("Cannot install qemu-img: %s", err)
		}
	}

	if err := execCMDWithSudo(ctx, comm, "pip3 install awscli"); err != nil {
		return fmt.Errorf("Install awscli: %s", err)
	}
	return nil
}

func (p *yum) InstallQemuIMG(ctx context.Context, comm packersdk.Communicator) error {
	if err := execCMDWithSudo(ctx, comm, "yum install -y libgcrypt qemu-img"); err != nil {
		return fmt.Errorf("Cannot install qemu-img: %s", err)
	}
	return nil
}

func execCMDWithSudo(ctx context.Context, comm packersdk.Communicator, cmdStr string) error {
	cmd := &packersdk.RemoteCmd{
		Command: cmdStr,
	}
	if err := comm.Start(ctx, cmd); err != nil {
		return err
	}
	if cmd.Wait() != 0 {
		cmd := &packersdk.RemoteCmd{
			Command: fmt.Sprintf("sudo %s", cmdStr),
		}
		if err := comm.Start(ctx, cmd); err != nil {
			return err
		}
		if cmd.Wait() != 0 {
			return fmt.Errorf("Bad exit code: %d", cmd.ExitStatus())
		}
	}
	return nil
}
