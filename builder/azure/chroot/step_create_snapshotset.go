package chroot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
)

var _ multistep.Step = &StepCreateSnapshotset{}

type StepCreateSnapshotset struct {
	OSDiskSnapshotID         string
	DataDiskSnapshotIDPrefix string
	Location                 string

	SkipCleanup bool

	snapshots Diskset
}

func (s *StepCreateSnapshotset) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	azcli := state.Get("azureclient").(client.AzureClientSet)
	ui := state.Get("ui").(packersdk.Ui)
	diskset := state.Get(stateBagKey_Diskset).(Diskset)

	s.snapshots = make(Diskset)

	errorMessage := func(format string, params ...interface{}) multistep.StepAction {
		err := fmt.Errorf("StepCreateSnapshotset.Run: error: "+format, params...)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	for lun, resource := range diskset {
		snapshotID := fmt.Sprintf("%s%d", s.DataDiskSnapshotIDPrefix, lun)
		if lun == -1 {
			snapshotID = s.OSDiskSnapshotID
		}
		ssr, err := client.ParseResourceID(snapshotID)
		if err != nil {
			errorMessage("Could not create a valid resource id, tried %q: %v", snapshotID, err)
		}
		if !strings.EqualFold(ssr.Provider, "Microsoft.Compute") ||
			!strings.EqualFold(ssr.ResourceType.String(), "snapshots") {
			return errorMessage("Resource %q is not of type Microsoft.Compute/snapshots", snapshotID)
		}
		s.snapshots[lun] = ssr
		state.Put(stateBagKey_Snapshotset, s.snapshots)

		ui.Say(fmt.Sprintf("Creating snapshot %q", ssr))

		snapshot := compute.Snapshot{
			Location: to.StringPtr(s.Location),
			SnapshotProperties: &compute.SnapshotProperties{
				CreationData: &compute.CreationData{
					CreateOption:     compute.Copy,
					SourceResourceID: to.StringPtr(resource.String()),
				},
				Incremental: to.BoolPtr(false),
			},
		}

		f, err := azcli.SnapshotsClient().CreateOrUpdate(ctx, ssr.ResourceGroup, ssr.ResourceName.String(), snapshot)
		if err != nil {
			return errorMessage("error initiating snapshot %q: %v", ssr, err)
		}

		pollClient := azcli.PollClient()
		pollClient.PollingDelay = 2 * time.Second
		ctx, cancel := context.WithTimeout(ctx, time.Hour*12)
		defer cancel()
		err = f.WaitForCompletionRef(ctx, pollClient)

		if err != nil {
			return errorMessage("error creating snapshot '%s': %v", s.OSDiskSnapshotID, err)
		}
	}

	return multistep.ActionContinue
}

func (s *StepCreateSnapshotset) Cleanup(state multistep.StateBag) {
	if !s.SkipCleanup {
		azcli := state.Get("azureclient").(client.AzureClientSet)
		ui := state.Get("ui").(packersdk.Ui)

		for _, resource := range s.snapshots {

			ui.Say(fmt.Sprintf("Removing any active SAS for snapshot %q", resource))
			{
				f, err := azcli.SnapshotsClient().RevokeAccess(context.TODO(), resource.ResourceGroup, resource.ResourceName.String())
				if err == nil {
					log.Printf("StepCreateSnapshotset.Cleanup: removing SAS...")
					err = f.WaitForCompletionRef(context.TODO(), azcli.PollClient())
				}
				if err != nil {
					log.Printf("StepCreateSnapshotset.Cleanup: error: %+v", err)
					ui.Error(fmt.Sprintf("error deleting snapshot %q: %v.", resource, err))
				}
			}

			ui.Say(fmt.Sprintf("Deleting snapshot %q", resource))
			{
				f, err := azcli.SnapshotsClient().Delete(context.TODO(), resource.ResourceGroup, resource.ResourceName.String())
				if err == nil {
					log.Printf("StepCreateSnapshotset.Cleanup: deleting snapshot...")
					err = f.WaitForCompletionRef(context.TODO(), azcli.PollClient())
				}
				if err != nil {
					log.Printf("StepCreateSnapshotset.Cleanup: error: %+v", err)
					ui.Error(fmt.Sprintf("error deleting snapshot %q: %v.", resource, err))
				}
			}
		}
	}
}
