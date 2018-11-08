package vsphere_template

import (
	"context"
	"strings"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/post-processor/vsphere"
	"github.com/vmware/govmomi"
)

type stepCreateSnapshot struct {
	VMName              string
	RemoteFolder        string
	SnapshotName        string
	SnapshotDescription string
	SnapshotMemory      bool
	SnapshotQuiesce     bool
	SnapshotEnable      bool
}

func NewStepCreateSnapshot(artifact packer.Artifact, p *PostProcessor) *stepCreateSnapshot {
	remoteFolder := "Discovered virtual machine"
	vmname := artifact.Id()

	if artifact.BuilderId() == vsphere.BuilderId {
		id := strings.Split(artifact.Id(), "::")
		remoteFolder = id[1]
		vmname = id[2]
	}

	return &stepCreateSnapshot{
		VMName:              vmname,
		RemoteFolder:        remoteFolder,
		SnapshotEnable:      p.config.SnapshotEnable,
		SnapshotName:        p.config.SnapshotName,
		SnapshotDescription: p.config.SnapshotDescription,
		SnapshotMemory:      p.config.SnapshotMemory,
		SnapshotQuiesce:     p.config.SnapshotQuiesce,
	}
}

func (s *stepCreateSnapshot) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	cli := state.Get("client").(*govmomi.Client)
	dcPath := state.Get("dcPath").(string)

	if !s.SnapshotEnable {
		ui.Message("Snapshot Not Enabled, Continue...")
		return multistep.ActionContinue
	}

	ui.Message("Creating a Snapshot...")

	vm, err := findRuntimeVM(cli, dcPath, s.VMName, s.RemoteFolder)

	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	task, err := vm.CreateSnapshot(context.Background(), s.SnapshotName, s.SnapshotDescription, s.SnapshotMemory, s.SnapshotQuiesce)

	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if err = task.Wait(context.Background()); err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepCreateSnapshot) Cleanup(multistep.StateBag) {}
