package chroot

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepResolvePlatformImageVersion_Run(t *testing.T) {

	pi := &StepResolvePlatformImageVersion{
		PlatformImage: &client.PlatformImage{
			Version: "latest",
		}}

	m := compute.NewVirtualMachineImagesClient("subscriptionId")
	m.Sender = autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
		if !strings.Contains(r.URL.String(), "%24orderby=name+desc") {
			t.Errorf("Expected url to use odata based sorting, but got %q", r.URL.String())
		}
		return &http.Response{
			Request: r,
			Body: ioutil.NopCloser(strings.NewReader(
				`[
					{"name":"1.2.3"},
					{"name":"4.5.6"}
				]`)),
			StatusCode: 200,
		}, nil
	})

	state := new(multistep.BasicStateBag)
	state.Put("azureclient", &client.AzureClientSetMock{
		VirtualMachineImagesClientMock: client.VirtualMachineImagesClient{
			VirtualMachineImagesClientAPI: m}})

	ui, getErrs := testUI()
	state.Put("ui", ui)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	got := pi.Run(ctx, state)
	if got != multistep.ActionContinue {
		t.Errorf("Expected 'continue', but got %q", got)
	}

	if pi.PlatformImage.Version != "1.2.3" {
		t.Errorf("Expected version '1.2.3', but got %q", pi.PlatformImage.Version)
	}

	_ = getErrs
}
