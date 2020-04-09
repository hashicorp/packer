package chroot

import (
	"context"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

func TestStepVerifySharedImageDestination_Run(t *testing.T) {

	type fields struct {
		Image    SharedImageGalleryDestination
		Location string
	}
	tests := []struct {
		name    string
		fields  fields
		want    multistep.StepAction
		wantErr string
	}{
		{
			name: "happy path",
			want: multistep.ActionContinue,
			fields: fields{
				Image: SharedImageGalleryDestination{
					ResourceGroup: "rg",
					GalleryName:   "gallery",
					ImageName:     "image",
					ImageVersion:  "1.2.3",
				},
				Location: "region1",
			},
		},
		{
			name:    "not found",
			want:    multistep.ActionHalt,
			wantErr: `Error retrieving shared image "/subscriptions/subscriptionID/resourcegroup/other-rg/providers/Microsoft.Compute/galleries/gallery/images/image": compute.GalleryImagesClient#Get: Failure responding to request: StatusCode=404 -- Original Error: autorest/azure: Service returned an error. Status=404 Code="NotFound" Message="Not found" `,
			fields: fields{
				Image: SharedImageGalleryDestination{
					ResourceGroup: "other-rg",
					GalleryName:   "gallery",
					ImageName:     "image",
					ImageVersion:  "1.2.3",
				},
				Location: "region1",
			},
		},
		{
			name:    "wrong region",
			want:    multistep.ActionHalt,
			wantErr: "Destination shared image resource \"image-resourceid-goes-here\" is in a different location (\"region1\") than this VM (\"other-region\"). Packer does not know how to handle that.",
			fields: fields{
				Image: SharedImageGalleryDestination{
					ResourceGroup: "rg",
					GalleryName:   "gallery",
					ImageName:     "image",
					ImageVersion:  "1.2.3",
				},
				Location: "other-region",
			},
		},
		{
			name:    "version exists",
			want:    multistep.ActionHalt,
			wantErr: "Shared image version \"2.3.4\" already exists for image \"image-resourceid-goes-here\".",
			fields: fields{
				Image: SharedImageGalleryDestination{
					ResourceGroup: "rg",
					GalleryName:   "gallery",
					ImageName:     "image",
					ImageVersion:  "2.3.4",
				},
				Location: "region1",
			},
		},
		{
			name:    "not Linux",
			want:    multistep.ActionHalt,
			wantErr: "The shared image (\"windows-image-resourceid-goes-here\") is not a Linux image (found \"Windows\"). Currently only Linux images are supported.",
			fields: fields{
				Image: SharedImageGalleryDestination{
					ResourceGroup: "rg",
					GalleryName:   "gallery",
					ImageName:     "windowsimage",
					ImageVersion:  "1.2.3",
				},
				Location: "region1",
			},
		},
	}
	for _, tt := range tests {
		gi := compute.NewGalleryImagesClient("subscriptionID")
		gi.Sender = autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
			switch {
			case r.Method == "GET" && strings.HasPrefix(r.URL.RequestURI(),
				"/subscriptions/subscriptionID/resourceGroups/rg/providers/Microsoft.Compute/galleries/gallery/images/image"):
				return &http.Response{
					Request: r,
					Body: ioutil.NopCloser(strings.NewReader(`{
						"id": "image-resourceid-goes-here",
						"location": "region1",
						"properties": {
							"osType": "Linux"
						}
					}`)),
					StatusCode: 200,
				}, nil
			case r.Method == "GET" && strings.HasPrefix(r.URL.RequestURI(),
				"/subscriptions/subscriptionID/resourceGroups/rg/providers/Microsoft.Compute/galleries/gallery/images/windowsimage"):
				return &http.Response{
					Request: r,
					Body: ioutil.NopCloser(strings.NewReader(`{
						"id": "windows-image-resourceid-goes-here",
						"location": "region1",
						"properties": {
							"osType": "Windows"
						}
					}`)),
					StatusCode: 200,
				}, nil
			}
			return &http.Response{
				Request: r,
				Body: ioutil.NopCloser(strings.NewReader(`{
					"Code": "NotFound",
					"Message": "Not found"
				}`)),
				StatusCode: 404,
			}, nil
		})

		giv := compute.NewGalleryImageVersionsClient("subscriptionID")
		giv.Sender = autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
			if !(r.Method == "GET" && strings.HasPrefix(r.URL.RequestURI(),
				"/subscriptions/subscriptionID/resourceGroups/rg/providers/Microsoft.Compute/galleries/gallery/images/image/versions")) {
				t.Errorf("Unexpected HTTP call: %s %s", r.Method, r.URL.RequestURI())
			}
			return &http.Response{
				Request: r,
				Body: ioutil.NopCloser(strings.NewReader(`{
						"value": [
							{
								"name": "2.3.4"
							}
						]
					}`)),
				StatusCode: 200,
			}, nil
		})

		state := new(multistep.BasicStateBag)
		state.Put("azureclient", &client.AzureClientSetMock{
			SubscriptionIDMock:             "subscriptionID",
			GalleryImagesClientMock:        gi,
			GalleryImageVersionsClientMock: giv,
		})
		state.Put("ui", packer.TestUi(t))

		t.Run(tt.name, func(t *testing.T) {
			s := &StepVerifySharedImageDestination{
				Image:    tt.fields.Image,
				Location: tt.fields.Location,
			}
			if got := s.Run(context.TODO(), state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StepVerifySharedImageDestination.Run() = %v, want %v", got, tt.want)
			}
		})
		if err, ok := state.GetOk("error"); ok {
			if err.(error).Error() != tt.wantErr {
				t.Errorf("Unexpected error, got: %q, want: %q", err, tt.wantErr)
			}
		} else if tt.wantErr != "" {
			t.Errorf("Expected error, but didn't get any")
		}
	}
}
