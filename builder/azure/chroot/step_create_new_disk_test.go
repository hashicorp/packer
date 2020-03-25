package chroot

import (
	"context"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

func TestStepCreateNewDisk_Run(t *testing.T) {
	type fields struct {
		ResourceID             string
		DiskSizeGB             int32
		DiskStorageAccountType string
		HyperVGeneration       string
		Location               string
		PlatformImage          *client.PlatformImage
		SourceDiskResourceID   string

		expectedPutDiskBody string
	}
	tests := []struct {
		name   string
		fields fields
		want   multistep.StepAction
	}{
		{
			name: "HappyPathDiskSource",
			fields: fields{
				ResourceID:             "/subscriptions/SubscriptionID/resourcegroups/ResourceGroupName/providers/Microsoft.Compute/disks/TemporaryOSDiskName",
				DiskSizeGB:             42,
				DiskStorageAccountType: string(compute.PremiumLRS),
				HyperVGeneration:       string(compute.V1),
				Location:               "westus",
				SourceDiskResourceID:   "SourceDisk",

				expectedPutDiskBody: `
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
				}`,
			},
			want: multistep.ActionContinue,
		},
		{
			name: "HappyPathDiskSource",
			fields: fields{
				ResourceID:             "/subscriptions/SubscriptionID/resourcegroups/ResourceGroupName/providers/Microsoft.Compute/disks/TemporaryOSDiskName",
				DiskStorageAccountType: string(compute.StandardLRS),
				HyperVGeneration:       string(compute.V1),
				Location:               "westus",
				PlatformImage: &client.PlatformImage{
					Publisher: "Microsoft",
					Offer:     "Windows",
					Sku:       "2016-DataCenter",
					Version:   "2016.1.4",
				},

				expectedPutDiskBody: `
				{
					"location": "westus",
					"properties": {
						"osType": "Linux",
						"hyperVGeneration": "V1",
						"creationData": {
							"createOption":"FromImage",
							"imageReference": {
								"id":"/subscriptions/SubscriptionID/providers/Microsoft.Compute/locations/westus/publishers/Microsoft/artifacttypes/vmimage/offers/Windows/skus/2016-DataCenter/versions/2016.1.4"
							}
						}
					},
					"sku": {
						"name": "Standard_LRS"
					}
				}`,
			},
			want: multistep.ActionContinue,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := StepCreateNewDisk{
				ResourceID:             tt.fields.ResourceID,
				DiskSizeGB:             tt.fields.DiskSizeGB,
				DiskStorageAccountType: tt.fields.DiskStorageAccountType,
				HyperVGeneration:       tt.fields.HyperVGeneration,
				Location:               tt.fields.Location,
				PlatformImage:          tt.fields.PlatformImage,
				SourceDiskResourceID:   tt.fields.SourceDiskResourceID,
			}

			expectedPutDiskBody := regexp.MustCompile(`[\s\n]`).ReplaceAllString(tt.fields.expectedPutDiskBody, "")

			m := compute.NewDisksClient("subscriptionId")
			m.Sender = autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
				if r.Method != "PUT" {
					t.Fatal("Expected only a PUT disk call")
				}
				b, _ := ioutil.ReadAll(r.Body)
				if string(b) != expectedPutDiskBody {
					t.Fatalf("expected body to be %q, but got %q", expectedPutDiskBody, string(b))
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

			if got := s.Run(context.TODO(), state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StepCreateNewDisk.Run() = %v, want %v", got, tt.want)
			}
		})
	}
}
