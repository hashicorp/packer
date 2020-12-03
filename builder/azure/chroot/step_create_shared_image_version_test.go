package chroot

import (
	"context"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/go-autorest/autorest"
)

func TestStepCreateSharedImageVersion_Run(t *testing.T) {
	type fields struct {
		Destination       SharedImageGalleryDestination
		OSDiskCacheType   string
		DataDiskCacheType string
		Location          string
	}
	tests := []struct {
		name            string
		fields          fields
		snapshotset     Diskset
		want            multistep.StepAction
		expectedPutBody string
	}{
		{
			name: "happy path",
			fields: fields{
				Destination: SharedImageGalleryDestination{
					ResourceGroup: "ResourceGroup",
					GalleryName:   "GalleryName",
					ImageName:     "ImageName",
					ImageVersion:  "0.1.2",
					TargetRegions: []TargetRegion{
						{
							Name:               "region1",
							ReplicaCount:       5,
							StorageAccountType: "Standard_ZRS",
						},
					},
					ExcludeFromLatest: true,
				},
				OSDiskCacheType:   "ReadWrite",
				DataDiskCacheType: "None",
				Location:          "region2",
			},
			snapshotset: diskset(
				"/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/snapshots/osdisksnapshot",
				"/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/snapshots/datadisksnapshot0",
				"/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/snapshots/datadisksnapshot1",
				"/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/snapshots/datadisksnapshot2"),
			expectedPutBody: `{
				"location": "region2",
				"properties": {
					"publishingProfile": {
						"targetRegions": [
							{
								"name": "region1",
								"regionalReplicaCount": 5,
								"storageAccountType": "Standard_ZRS"
							}
						],
						"excludeFromLatest": true
					},
					"storageProfile": {
						"osDiskImage": {
							"hostCaching": "ReadWrite",
							"source": {
								"id": "/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/snapshots/osdisksnapshot"
							}
						},
						"dataDiskImages": [
							{
								"lun": 0,
								"hostCaching": "None",
								"source": {
									"id": "/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/snapshots/datadisksnapshot0"
								}
							},
							{
								"lun": 1,
								"hostCaching": "None",
								"source": {
									"id": "/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/snapshots/datadisksnapshot1"
								}
							},
							{
								"lun": 2,
								"hostCaching": "None",
								"source": {
									"id": "/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/snapshots/datadisksnapshot2"
								}
							}
						]
					}
				}
			}`,
		},
	}
	for _, tt := range tests {
		expectedPutBody := regexp.MustCompile(`[\s\n]`).ReplaceAllString(tt.expectedPutBody, "")

		m := compute.NewGalleryImageVersionsClient("subscriptionId")
		m.Sender = autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != "PUT" {
				t.Fatal("Expected only a PUT call")
			}
			b, _ := ioutil.ReadAll(r.Body)
			if string(b) != expectedPutBody {
				t.Errorf("expected body to be %v, but got %v", expectedPutBody, string(b))
			}
			return &http.Response{
				Request:    r,
				StatusCode: 200,
			}, nil
		})

		state := new(multistep.BasicStateBag)
		state.Put("azureclient", &client.AzureClientSetMock{
			GalleryImageVersionsClientMock: m,
		})
		state.Put("ui", packersdk.TestUi(t))
		state.Put(stateBagKey_Snapshotset, tt.snapshotset)

		t.Run(tt.name, func(t *testing.T) {
			s := &StepCreateSharedImageVersion{
				Destination:       tt.fields.Destination,
				OSDiskCacheType:   tt.fields.OSDiskCacheType,
				DataDiskCacheType: tt.fields.DataDiskCacheType,
				Location:          tt.fields.Location,
			}
			if got := s.Run(context.TODO(), state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StepCreateSharedImageVersion.Run() = %v, want %v", got, tt.want)
			}
		})
	}
}
