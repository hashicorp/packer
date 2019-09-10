package chroot

import (
	"context"
	"io/ioutil"
	"net/http"
	"regexp"
	"testing"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
		"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

func Test_StepCreateNewDisk_FromDisk(t *testing.T) {
	sut := StepCreateNewDisk{
		SubscriptionID:         "SubscriptionID",
		ResourceGroup:          "ResourceGroupName",
		DiskName:               "TemporaryOSDiskName",
		DiskSizeGB:             42,
		DiskStorageAccountType: string(compute.PremiumLRS),
		HyperVGeneration:       string(compute.V1),
		Location:               "westus",
		SourceDiskResourceID:   "SourceDisk",
	}

	expected := regexp.MustCompile(`[\s\n]`).ReplaceAllString(`
{
	"location": "westus",
	"properties": {
		"osType": "Linux",
		"hyperVGeneration": "V1",
		"creationData": {
			"createOption": "Copy",
			"sourceResourceId": "SourceDisk"
		},
		"diskSizeGB": 42
	},
	"sku": {
		"name": "Premium_LRS"
	}
}`, "")

	m := compute.NewDisksClient("subscriptionId")
	m.Sender = autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
		b, _ := ioutil.ReadAll(r.Body)
		if string(b) != expected {
			t.Fatalf("expected body to be %q, but got %q", expected, string(b))
		}
		return &http.Response{
			Request:    r,
			StatusCode: 200,
		}, nil
	})

	state := new(multistep.BasicStateBag)
	state.Put("azureclient", &client.AzureClientSetMock{
		DisksClientMock: m,
	})
	state.Put("ui", packer.TestUi(t))

	r := sut.Run(context.TODO(), state)

	if r != multistep.ActionContinue {
		t.Fatal("Run failed")
	}
}
