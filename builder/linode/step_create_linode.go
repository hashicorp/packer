package linode

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/linode/linodego"
)

type stepCreateLinode struct {
	client linodego.Client
}

func (s *stepCreateLinode) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Creating Linode...")

	createOpts := linodego.InstanceCreateOptions{
		RootPass:       c.Comm.Password(),
		AuthorizedKeys: []string{string(c.Comm.SSHPublicKey)},
		Region:         c.Region,
		Type:           c.InstanceType,
		Label:          c.Label,
		Image:          c.Image,
		SwapSize:       &c.SwapSize,
	}

	instance, err := s.client.CreateInstance(ctx, createOpts)
	if err != nil {
		err = errors.New("Error creating Linode: " + err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("instance", instance)

	// wait until instance is running
	for instance.Status != linodego.InstanceRunning {
		time.Sleep(2 * time.Second)
		if instance, err = s.client.GetInstance(ctx, instance.ID); err != nil {
			err = errors.New("Error creating Linode: " + err.Error())
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		state.Put("instance", instance)
		// instance_id is the generic term used so that users can have access to the
		// instance id inside of the provisioners, used in step_provision.
		state.Put("instance_id", instance.ID)
	}

	disk, err := s.findDisk(ctx, instance.ID)
	if err != nil {
		err = errors.New("Error creating Linode: " + err.Error())
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	} else if disk == nil {
		err := errors.New("Error creating Linode: no suitable disk was found")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("disk", disk)
	return multistep.ActionContinue
}

func (s *stepCreateLinode) findDisk(ctx context.Context, instanceID int) (*linodego.InstanceDisk, error) {
	disks, err := s.client.ListInstanceDisks(ctx, instanceID, nil)
	if err != nil {
		return nil, err
	}
	for _, disk := range disks {
		if disk.Filesystem != linodego.FilesystemSwap {
			return &disk, nil
		}
	}
	return nil, nil
}

func (s *stepCreateLinode) Cleanup(state multistep.StateBag) {
	instance, ok := state.GetOk("instance")
	if !ok {
		return
	}

	ui := state.Get("ui").(packersdk.Ui)

	if err := s.client.DeleteInstance(context.Background(), instance.(*linodego.Instance).ID); err != nil {
		ui.Error("Error cleaning up Linode: " + err.Error())
	}
}
