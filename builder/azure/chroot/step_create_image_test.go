package chroot

import (
	"context"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/go-autorest/autorest"
)

func TestStepCreateImage_Run(t *testing.T) {
	type fields struct {
		ImageResourceID            string
		ImageOSState               string
		OSDiskStorageAccountType   string
		OSDiskCacheType            string
		DataDiskStorageAccountType string
		DataDiskCacheType          string
		Location                   string
	}
	tests := []struct {
		name        string
		fields      fields
		diskset     Diskset
		want        multistep.StepAction
		wantPutBody string
	}{
		{
			name: "happy path",
			fields: fields{
				ImageResourceID:            "/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/images/myImage",
				Location:                   "location1",
				OSDiskStorageAccountType:   "Standard_LRS",
				OSDiskCacheType:            "ReadWrite",
				DataDiskStorageAccountType: "Premium_LRS",
				DataDiskCacheType:          "ReadOnly",
			},
			diskset: diskset(
				"/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/disks/osdisk",
				"/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/disks/datadisk0",
				"/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/disks/datadisk1",
				"/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/disks/datadisk2"),
			want: multistep.ActionContinue,
			wantPutBody: `{
				"location": "location1",
				"properties": {
					"storageProfile": {
						"osDisk": {
							"osType": "Linux",
							"managedDisk": {
								"id": "/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/disks/osdisk"
							},
							"caching": "ReadWrite",
							"storageAccountType": "Standard_LRS"
						},
						"dataDisks": [
							{
								"lun": 0,
								"managedDisk": {
									"id": "/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/disks/datadisk0"
								},
								"caching": "ReadOnly",
								"storageAccountType": "Premium_LRS"
							},
							{
								"lun": 1,
								"managedDisk": {
									"id": "/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/disks/datadisk1"
								},
								"caching": "ReadOnly",
								"storageAccountType": "Premium_LRS"
							},
							{
								"lun": 2,
								"managedDisk": {
									"id": "/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/disks/datadisk2"
								},
								"caching": "ReadOnly",
								"storageAccountType": "Premium_LRS"
							}
						]
					}
				}
			}`,
		},
	}
	for _, tt := range tests {

		ic := compute.NewImagesClient("subscriptionID")
		ic.Sender = autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != "PUT" {
				t.Fatal("Expected only a PUT call")
			}
			if tt.wantPutBody != "" {
				b, _ := ioutil.ReadAll(r.Body)
				expectedPutBody := regexp.MustCompile(`[\s\n]`).ReplaceAllString(tt.wantPutBody, "")
				if string(b) != expectedPutBody {
					t.Errorf("expected body to be %v, but got %v", expectedPutBody, string(b))
				}
			}
			return &http.Response{
				Request:    r,
				StatusCode: 200,
			}, nil
		})

		state := new(multistep.BasicStateBag)
		state.Put("azureclient", &client.AzureClientSetMock{
			ImagesClientMock: ic,
		})
		state.Put("ui", packer.TestUi(t))
		state.Put(stateBagKey_Diskset, tt.diskset)

		t.Run(tt.name, func(t *testing.T) {
			s := &StepCreateImage{
				ImageResourceID:            tt.fields.ImageResourceID,
				ImageOSState:               tt.fields.ImageOSState,
				OSDiskStorageAccountType:   tt.fields.OSDiskStorageAccountType,
				OSDiskCacheType:            tt.fields.OSDiskCacheType,
				DataDiskStorageAccountType: tt.fields.DataDiskStorageAccountType,
				DataDiskCacheType:          tt.fields.DataDiskCacheType,
				Location:                   tt.fields.Location,
			}
			if got := s.Run(context.TODO(), state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StepCreateImage.Run() = %v, want %v", got, tt.want)
			}
		})
	}
}
