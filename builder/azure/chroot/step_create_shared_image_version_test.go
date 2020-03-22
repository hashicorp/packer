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

func TestStepCreateSharedImageVersion_Run(t *testing.T) {
	type fields struct {
		Destination     SharedImageGalleryDestination
		OSDiskCacheType string
		Location        string
	}
	tests := []struct {
		name            string
		fields          fields
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
						TargetRegion{
							Name:               "region1",
							ReplicaCount:       5,
							StorageAccountType: "Standard_ZRS",
						},
					},
					ExcludeFromLatest: true,
				},
				Location: "region2",
			},
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
							"source": {
								"id": "osdisksnapshotresourceid"
							}
						}
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
		state.Put("ui", packer.TestUi(t))
		state.Put(stateBagKey_OSDiskSnapshotResourceID, "osdisksnapshotresourceid")

		t.Run(tt.name, func(t *testing.T) {
			s := &StepCreateSharedImageVersion{
				Destination:     tt.fields.Destination,
				OSDiskCacheType: tt.fields.OSDiskCacheType,
				Location:        tt.fields.Location,
			}
			if got := s.Run(context.TODO(), state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StepCreateSharedImageVersion.Run() = %v, want %v", got, tt.want)
			}
		})
	}
}
