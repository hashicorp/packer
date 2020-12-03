package chroot

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

// StepResolvePlatformImageVersion resolves the exact PIR version when the version is 'latest'
type StepResolvePlatformImageVersion struct {
	*client.PlatformImage
	Location string
}

// Run retrieves all available versions of a PIR image and stores the latest in the PlatformImage
func (pi *StepResolvePlatformImageVersion) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	if strings.EqualFold(pi.Version, "latest") {
		azcli := state.Get("azureclient").(client.AzureClientSet)

		vmi, err := azcli.VirtualMachineImagesClient().GetLatest(ctx, pi.Publisher, pi.Offer, pi.Sku, pi.Location)
		if err != nil {
			log.Printf("StepResolvePlatformImageVersion.Run: error: %+v", err)
			err := fmt.Errorf("error retieving latest version of %q: %v", pi.URN(), err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
		pi.Version = to.String(vmi.Name)
		ui.Say("Resolved latest version of source image: " + pi.Version)
	} else {
		ui.Say("Nothing to do, version is not 'latest'")
	}

	return multistep.ActionContinue
}

func (*StepResolvePlatformImageVersion) Cleanup(multistep.StateBag) {}
