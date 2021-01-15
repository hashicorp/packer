package yandexexport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/yandex"
)

type StepDump struct {
	ExtraSize bool
	SizeLimit int64
}

const (
	dumpCommand = "%sqemu-img convert -O qcow2 -o cluster_size=2M %s disk.qcow2 2>&1"
)

// Run reads the instance metadata and looks for the log entry
// indicating the cloud-init script finished.
func (s *StepDump) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	comm := state.Get("communicator").(packersdk.Communicator)

	device := "/dev/disk/by-id/virtio-doexport"
	cmdDumpCheckAccess := &packersdk.RemoteCmd{
		Command: fmt.Sprintf("qemu-img info %s", device),
	}
	if err := comm.Start(ctx, cmdDumpCheckAccess); err != nil {
		return yandex.StepHaltWithError(state, err)
	}
	sudo := ""
	if cmdDumpCheckAccess.Wait() != 0 {
		sudo = "sudo "
	}

	if s.ExtraSize && which(ctx, comm, "losetup") == nil {
		ui.Say("Map loop device...")
		buff := new(bytes.Buffer)
		cmd := &packersdk.RemoteCmd{
			Command: fmt.Sprintf("%slosetup --show -r --sizelimit %d -f %s", sudo, s.SizeLimit, device),
			Stdout:  buff,
		}
		if err := comm.Start(ctx, cmd); err != nil {
			return yandex.StepHaltWithError(state, err)
		}
		if cmd.Wait() != 0 {
			return yandex.StepHaltWithError(state, fmt.Errorf("Cannot losetup: %d", cmd.ExitStatus()))
		}
		device = strings.TrimSpace(buff.String())
		if device == "" {
			return yandex.StepHaltWithError(state, fmt.Errorf("Bad lo device"))
		}
	}
	wg := new(sync.WaitGroup)
	defer wg.Wait()
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()
	wg.Add(1)
	go func() {
		defer wg.Done()
		cmd := &packersdk.RemoteCmd{
			Command: "while true ; do sleep 3; sudo kill -s SIGUSR1 $(pidof qemu-img); done",
		}

		err := cmd.RunWithUi(ctxWithCancel, comm, ui)
		if err != nil && !errors.Is(err, context.Canceled) {
			ui.Error("qemu-img signal sender error: " + err.Error())
			return
		}
	}()

	cmdDump := &packersdk.RemoteCmd{
		Command: fmt.Sprintf(dumpCommand, sudo, device),
	}
	ui.Say("Dumping...")
	if err := cmdDump.RunWithUi(ctx, comm, ui); err != nil {
		return yandex.StepHaltWithError(state, err)
	}
	if cmdDump.ExitStatus() != 0 {
		return yandex.StepHaltWithError(state, fmt.Errorf("Cannot dump disk, exit code: %d", cmdDump.ExitStatus()))
	}

	return multistep.ActionContinue
}

// Cleanup nothing
func (s *StepDump) Cleanup(state multistep.StateBag) {}
