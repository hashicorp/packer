package chroot

import (
	"context"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/go-autorest/autorest"
)

func TestStepCreateNewDisk_Run(t *testing.T) {
	tests := []struct {
		name                  string
		fields                StepCreateNewDiskset
		expectedPutDiskBodies []string
		want                  multistep.StepAction
		verifyDiskset         *Diskset
	}{
		{
			name: "from disk",
			fields: StepCreateNewDiskset{
				OSDiskID:                 "/subscriptions/SubscriptionID/resourcegroups/ResourceGroupName/providers/Microsoft.Compute/disks/TemporaryOSDiskName",
				OSDiskSizeGB:             42,
				OSDiskStorageAccountType: string(compute.PremiumLRS),
				HyperVGeneration:         string(compute.V1),
				Location:                 "westus",
				SourceOSDiskResourceID:   "SourceDisk",
			},
			expectedPutDiskBodies: []string{`
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
				}`},
			want:          multistep.ActionContinue,
			verifyDiskset: &Diskset{-1: resource("/subscriptions/SubscriptionID/resourceGroups/ResourceGroupName/providers/Microsoft.Compute/disks/TemporaryOSDiskName")},
		},
		{
			name: "from platform image",
			fields: StepCreateNewDiskset{
				OSDiskID:                 "/subscriptions/SubscriptionID/resourcegroups/ResourceGroupName/providers/Microsoft.Compute/disks/TemporaryOSDiskName",
				OSDiskStorageAccountType: string(compute.StandardLRS),
				HyperVGeneration:         string(compute.V1),
				Location:                 "westus",
				SourcePlatformImage: &client.PlatformImage{
					Publisher: "Microsoft",
					Offer:     "Windows",
					Sku:       "2016-DataCenter",
					Version:   "2016.1.4",
				},
			},
			expectedPutDiskBodies: []string{`
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
				}`},
			want:          multistep.ActionContinue,
			verifyDiskset: &Diskset{-1: resource("/subscriptions/SubscriptionID/resourceGroups/ResourceGroupName/providers/Microsoft.Compute/disks/TemporaryOSDiskName")},
		},
		{
			name: "from shared image",
			fields: StepCreateNewDiskset{
				OSDiskID:                   "/subscriptions/SubscriptionID/resourcegroups/ResourceGroupName/providers/Microsoft.Compute/disks/TemporaryOSDiskName",
				OSDiskStorageAccountType:   string(compute.StandardLRS),
				DataDiskStorageAccountType: string(compute.PremiumLRS),
				DataDiskIDPrefix:           "/subscriptions/SubscriptionID/resourcegroups/ResourceGroupName/providers/Microsoft.Compute/disks/TemporaryDataDisk-",
				HyperVGeneration:           string(compute.V1),
				Location:                   "westus",
				SourceImageResourceID:      "/subscriptions/SubscriptionID/resourcegroups/imagegroup/providers/Microsoft.Compute/galleries/MyGallery/images/MyImage/versions/1.2.3",
			},

			expectedPutDiskBodies: []string{`
				{
					"location": "westus",
					"properties": {
						"osType": "Linux",
						"hyperVGeneration": "V1",
						"creationData": {
							"createOption":"FromImage",
							"galleryImageReference": {
								"id":"/subscriptions/SubscriptionID/resourcegroups/imagegroup/providers/Microsoft.Compute/galleries/MyGallery/images/MyImage/versions/1.2.3"
							}
						}
					},
					"sku": {
						"name": "Standard_LRS"
					}
				}`, `
				{
					"location": "westus",
					"properties": {
						"creationData": {
							"createOption":"FromImage",
							"galleryImageReference": {
								"id": "/subscriptions/SubscriptionID/resourcegroups/imagegroup/providers/Microsoft.Compute/galleries/MyGallery/images/MyImage/versions/1.2.3",
								"lun": 5
							}
						}
					},
					"sku": {
						"name": "Premium_LRS"
					}
				}`, `
				{
					"location": "westus",
					"properties": {
						"creationData": {
							"createOption":"FromImage",
							"galleryImageReference": {
								"id": "/subscriptions/SubscriptionID/resourcegroups/imagegroup/providers/Microsoft.Compute/galleries/MyGallery/images/MyImage/versions/1.2.3",
								"lun": 9
							}
						}
					},
					"sku": {
						"name": "Premium_LRS"
					}
				}`, `
				{
					"location": "westus",
					"properties": {
						"creationData": {
							"createOption":"FromImage",
							"galleryImageReference": {
								"id": "/subscriptions/SubscriptionID/resourcegroups/imagegroup/providers/Microsoft.Compute/galleries/MyGallery/images/MyImage/versions/1.2.3",
								"lun": 3
							}
						}
					},
					"sku": {
						"name": "Premium_LRS"
					}
				}`},
			want: multistep.ActionContinue,
			verifyDiskset: &Diskset{
				-1: resource("/subscriptions/SubscriptionID/resourceGroups/ResourceGroupName/providers/Microsoft.Compute/disks/TemporaryOSDiskName"),
				3:  resource("/subscriptions/SubscriptionID/resourceGroups/ResourceGroupName/providers/Microsoft.Compute/disks/TemporaryDataDisk-3"),
				5:  resource("/subscriptions/SubscriptionID/resourceGroups/ResourceGroupName/providers/Microsoft.Compute/disks/TemporaryDataDisk-5"),
				9:  resource("/subscriptions/SubscriptionID/resourceGroups/ResourceGroupName/providers/Microsoft.Compute/disks/TemporaryDataDisk-9"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.fields

			bodyCount := 0
			m := compute.NewDisksClient("SubscriptionID")
			m.Sender = autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
				if r.Method != "PUT" {
					t.Fatal("Expected only a PUT disk call")
				}
				b, _ := ioutil.ReadAll(r.Body)
				expectedPutDiskBody := regexp.MustCompile(`[\s\n]`).ReplaceAllString(tt.expectedPutDiskBodies[bodyCount], "")
				bodyCount++
				if string(b) != expectedPutDiskBody {
					t.Fatalf("expected body #%d to be %q, but got %q", bodyCount, expectedPutDiskBody, string(b))
				}
				return &http.Response{
					Request:    r,
					StatusCode: 200,
				}, nil
			})

			giv := compute.NewGalleryImageVersionsClient("SubscriptionID")
			giv.Sender = autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
				if r.Method == "GET" &&
					regexp.MustCompile(`(?i)/versions/1\.2\.3$`).MatchString(r.URL.Path) {
					return &http.Response{
						Request: r,
						Body: ioutil.NopCloser(strings.NewReader(`{
							"properties": { "storageProfile": {
								"dataDiskImages":[
									{ "lun": 5 },
									{ "lun": 9 },
									{ "lun": 3 }
								]
							} }
						}`)),
						StatusCode: 200,
					}, nil
				}
				return &http.Response{
					Request:    r,
					Status:     "Unexpected request",
					StatusCode: 500,
				}, nil
			})

			state := new(multistep.BasicStateBag)
			state.Put("azureclient", &client.AzureClientSetMock{
				SubscriptionIDMock:             "SubscriptionID",
				DisksClientMock:                m,
				GalleryImageVersionsClientMock: giv,
			})
			state.Put("ui", packer.TestUi(t))

			if got := s.Run(context.TODO(), state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StepCreateNewDisk.Run() = %v, want %v", got, tt.want)
			}

			ds := state.Get(stateBagKey_Diskset)
			if tt.verifyDiskset != nil && !reflect.DeepEqual(*tt.verifyDiskset, ds) {
				t.Errorf("Error verifying diskset after Run(), got %v, want %v", ds, *tt.verifyDiskset)
			}
		})
	}
}

func resource(id string) client.Resource {
	v, err := client.ParseResourceID(id)
	if err != nil {
		panic(err)
	}
	return v
}
