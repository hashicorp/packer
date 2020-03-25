package chroot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

var _ multistep.Step = &StepCreateSnapshot{}

type StepCreateSnapshot struct {
	ResourceID string
	Location   string

	SkipCleanup bool

	subscriptionID, resourceGroup, snapshotName string
}

func parseSnapshotResourceID(resourceID string) (subscriptionID, resourceGroup, snapshotName string, err error) {
	r, err := azure.ParseResourceID(resourceID)
	if err != nil {
		return "", "", "", err
	}

	if !strings.EqualFold(r.Provider, "Microsoft.Compute") ||
		!strings.EqualFold(r.ResourceType, "snapshots") {
		return "", "", "", fmt.Errorf("Resource %q is not of type Microsoft.Compute/snapshots", resourceID)
	}

	return r.SubscriptionID, r.ResourceGroup, r.ResourceName, nil
}

const (
	stateBagKey_OSDiskSnapshotResourceID = "os_disk_snapshot_resource_id"
)

func (s *StepCreateSnapshot) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	azcli := state.Get("azureclient").(client.AzureClientSet)
	ui := state.Get("ui").(packer.Ui)
	osDiskResourceID := state.Get(stateBagKey_OSDiskResourceID).(string)

	state.Put(stateBagKey_OSDiskSnapshotResourceID, s.ResourceID)
	ui.Say(fmt.Sprintf("Creating snapshot '%s'", s.ResourceID))

	var err error
	s.subscriptionID, s.resourceGroup, s.snapshotName, err = parseSnapshotResourceID(s.ResourceID)
	if err != nil {
		log.Printf("StepCreateSnapshot.Run: error: %+v", err)
		err := fmt.Errorf(
			"error parsing resource id '%s': %v", s.ResourceID, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	snapshot := compute.Snapshot{
		Location: to.StringPtr(s.Location),
		SnapshotProperties: &compute.SnapshotProperties{
			CreationData: &compute.CreationData{
				CreateOption:     compute.Copy,
				SourceResourceID: to.StringPtr(osDiskResourceID),
			},
			Incremental: to.BoolPtr(false),
		},
	}

	f, err := azcli.SnapshotsClient().CreateOrUpdate(ctx, s.resourceGroup, s.snapshotName, snapshot)
	if err == nil {
		err = f.WaitForCompletionRef(ctx, azcli.PollClient())
	}
	if err != nil {
		log.Printf("StepCreateSnapshot.Run: error: %+v", err)
		err := fmt.Errorf(
			"error creating snapshot '%s': %v", s.ResourceID, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepCreateSnapshot) Cleanup(state multistep.StateBag) {
	if !s.SkipCleanup {
		azcli := state.Get("azureclient").(client.AzureClientSet)
		ui := state.Get("ui").(packer.Ui)

		ui.Say(fmt.Sprintf("Removing any active SAS for snapshot %q", s.ResourceID))
		{
			f, err := azcli.SnapshotsClient().RevokeAccess(context.TODO(), s.resourceGroup, s.snapshotName)
			if err == nil {
				log.Printf("StepCreateSnapshot.Cleanup: removing SAS...")
				err = f.WaitForCompletionRef(context.TODO(), azcli.PollClient())
			}
			if err != nil {
				log.Printf("StepCreateSnapshot.Cleanup: error: %+v", err)
				ui.Error(fmt.Sprintf("error deleting snapshot '%s': %v.", s.ResourceID, err))
			}
		}

		ui.Say(fmt.Sprintf("Deleting snapshot %q", s.ResourceID))
		{
			f, err := azcli.SnapshotsClient().Delete(context.TODO(), s.resourceGroup, s.snapshotName)
			if err == nil {
				log.Printf("StepCreateSnapshot.Cleanup: deleting snapshot...")
				err = f.WaitForCompletionRef(context.TODO(), azcli.PollClient())
			}
			if err != nil {
				log.Printf("StepCreateSnapshot.Cleanup: error: %+v", err)
				ui.Error(fmt.Sprintf("error deleting snapshot '%s': %v.", s.ResourceID, err))
			}
		}
	}
}
