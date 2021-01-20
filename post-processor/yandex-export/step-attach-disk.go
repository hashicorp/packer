package yandexexport

import (
	"context"
	"fmt"

	"github.com/c2h5oh/datasize"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/yandex"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

type StepAttachDisk struct {
	yandex.CommonConfig
	ImageID   string
	ExtraSize int
}

func (c *StepAttachDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(yandex.Driver)
	ui := state.Get("ui").(packersdk.Ui)
	instanceID := state.Get("instance_id").(string)

	ui.Say("Create secondary disk from image for export...")

	imageDesc, err := driver.SDK().Compute().Image().Get(ctx, &compute.GetImageRequest{
		ImageId: c.ImageID,
	})
	if err != nil {
		return yandex.StepHaltWithError(state, err)
	}

	op, err := driver.SDK().WrapOperation(driver.SDK().Compute().Disk().Create(ctx, &compute.CreateDiskRequest{
		Source: &compute.CreateDiskRequest_ImageId{
			ImageId: c.ImageID,
		},
		Name:        fmt.Sprintf("export-%s-disk", instanceID),
		Size:        int64(datasize.ByteSize(c.ExtraSize)*datasize.GB) + imageDesc.GetMinDiskSize(),
		ZoneId:      c.Zone,
		FolderId:    c.FolderID,
		TypeId:      c.DiskType,
		Description: "Temporary disk for exporting",
	}))
	if op == nil {
		return yandex.StepHaltWithError(state, err)
	}
	protoMD, err := op.Metadata()
	if err != nil {
		return yandex.StepHaltWithError(state, err)
	}
	md, ok := protoMD.(*compute.CreateDiskMetadata)
	if !ok {
		return yandex.StepHaltWithError(state, fmt.Errorf("could not get Disk ID from create operation metadata"))
	}
	state.Put("secondary_disk_id", md.GetDiskId())

	if err := op.Wait(ctx); err != nil {
		return yandex.StepHaltWithError(state, err)
	}

	ui.Say("Attach secondary disk to instance...")

	op, err = driver.SDK().WrapOperation(driver.SDK().Compute().Instance().AttachDisk(ctx, &compute.AttachInstanceDiskRequest{
		InstanceId: instanceID,
		AttachedDiskSpec: &compute.AttachedDiskSpec{
			AutoDelete: true,
			DeviceName: "doexport",
			Disk: &compute.AttachedDiskSpec_DiskId{
				DiskId: md.GetDiskId(),
			},
		},
	}))
	if err != nil {
		return yandex.StepHaltWithError(state, err)
	}
	ui.Message("Wait attached disk...")
	if err := op.Wait(ctx); err != nil {
		return yandex.StepHaltWithError(state, err)
	}

	state.Remove("secondary_disk_id")
	return multistep.ActionContinue
}

func (s *StepAttachDisk) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packersdk.Ui)
	driver := state.Get("driver").(yandex.Driver)
	if diskID, ok := state.GetOk("secondary_disk_id"); ok {
		ui.Say("Remove the secondary disk...")
		op, err := driver.SDK().WrapOperation(driver.SDK().Compute().Disk().Delete(context.Background(), &compute.DeleteDiskRequest{
			DiskId: diskID.(string),
		}))
		if err != nil {
			ui.Error(err.Error())
			return
		}
		if err := op.Wait(context.Background()); err != nil {
			ui.Error(err.Error())
		}
	}
}
