package chroot

import (
	"context"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/azure/common/client"
)

func TestStepVerifySharedImageSource_Run(t *testing.T) {
	type fields struct {
		SharedImageID  string
		SubscriptionID string
		Location       string
	}
	tests := []struct {
		name    string
		fields  fields
		want    multistep.StepAction
		wantErr string
	}{
		{
			name: "happy path",
			fields: fields{
				SharedImageID: "/subscriptions/subscriptionID/resourceGroups/rg/providers/Microsoft.Compute/galleries/myGallery/images/myImage/versions/1.2.3",
				Location:      "VM location",
			},
		},
		{
			name: "resource is not a shared image",
			fields: fields{
				SharedImageID: "/subscriptions/subscriptionID/resourceGroups/rg/providers/Microsoft.Compute/disks/myDisk",
				Location:      "VM location",
			},
			want:    multistep.ActionHalt,
			wantErr: "does not identify a shared image version",
		},
		{
			name: "error in resource id",
			fields: fields{
				SharedImageID: "not-a-resource-id",
			},
			want:    multistep.ActionHalt,
			wantErr: "Could not parse resource id",
		},
		{
			name: "wrong location",
			fields: fields{
				SharedImageID: "/subscriptions/subscriptionID/resourceGroups/rg/providers/Microsoft.Compute/galleries/myGallery/images/myImage/versions/1.2.3",
				Location:      "other location",
			},
			want:    multistep.ActionHalt,
			wantErr: "does not include VM location",
		},
		{
			name: "image not found",
			fields: fields{
				SharedImageID: "/subscriptions/subscriptionID/resourceGroups/rg/providers/Microsoft.Compute/galleries/myGallery/images/myImage/versions/2.3.4",
				Location:      "vm location",
			},
			want:    multistep.ActionHalt,
			wantErr: "Error retrieving shared image version",
		},
		{
			name: "windows image",
			fields: fields{
				SharedImageID: "/subscriptions/subscriptionID/resourceGroups/rg/providers/Microsoft.Compute/galleries/myGallery/images/windowsImage/versions/1.2.3",
				Location:      "VM location",
			},
			want:    multistep.ActionHalt,
			wantErr: "not a Linux image",
		},
	}
	for _, tt := range tests {
		giv := compute.NewGalleryImageVersionsClient("subscriptionID")
		giv.Sender = autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method == "GET" {
				switch {
				case strings.HasSuffix(r.URL.Path, "/versions/1.2.3"):
					return &http.Response{
						Request: r,
						Body: ioutil.NopCloser(strings.NewReader(`{
								"id": "image-version-id",
								"properties": {
									"publishingProfile": {
										"targetRegions": [
											{ "name": "vm Location" }
										]
									}
								}
							}`)),
						StatusCode: 200,
					}, nil
				case regexp.MustCompile(`(?i)^/subscriptions/subscriptionID/resourceGroups/rg/providers/Microsoft.Compute/galleries/myGallery/images/myImage/versions/\d+\.\d+\.\d+$`).
					MatchString(r.URL.Path):
					return &http.Response{
						Request:    r,
						Body:       ioutil.NopCloser(strings.NewReader(`{"error":{"code":"NotFound"}}`)),
						StatusCode: 404,
					}, nil
				}
			}

			t.Errorf("Unexpected HTTP call: %s %s", r.Method, r.URL.RequestURI())
			return &http.Response{
				Request:    r,
				Status:     "Unexpected HTTP call",
				Body:       ioutil.NopCloser(strings.NewReader(`{"code":"TestError"}`)),
				StatusCode: 599,
			}, nil
		})

		gi := compute.NewGalleryImagesClient("subscriptionID")
		gi.Sender = autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method == "GET" {
				switch {
				case strings.HasSuffix(r.URL.Path, "/images/myImage"):
					return &http.Response{
						Request: r,
						Body: ioutil.NopCloser(strings.NewReader(`{
						"id": "image-id",
						"properties": {
							"osType": "Linux"
						}
					}`)),
						StatusCode: 200,
					}, nil
				case strings.HasSuffix(r.URL.Path, "/images/windowsImage"):
					return &http.Response{
						Request: r,
						Body: ioutil.NopCloser(strings.NewReader(`{
							"id": "image-id",
							"properties": {
								"osType": "Windows"
							}
						}`)),
						StatusCode: 200,
					}, nil
				}
			}

			t.Errorf("Unexpected HTTP call: %s %s", r.Method, r.URL.RequestURI())
			return &http.Response{
				Request:    r,
				Status:     "Unexpected HTTP call",
				Body:       ioutil.NopCloser(strings.NewReader(`{"error":{"code":"TestError"}}`)),
				StatusCode: 599,
			}, nil
		})

		state := new(multistep.BasicStateBag)
		state.Put("azureclient", &client.AzureClientSetMock{
			SubscriptionIDMock:             "subscriptionID",
			GalleryImageVersionsClientMock: giv,
			GalleryImagesClientMock:        gi,
		})
		state.Put("ui", packersdk.TestUi(t))

		t.Run(tt.name, func(t *testing.T) {
			s := &StepVerifySharedImageSource{
				SharedImageID:  tt.fields.SharedImageID,
				SubscriptionID: tt.fields.SubscriptionID,
				Location:       tt.fields.Location,
			}
			if got := s.Run(context.TODO(), state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StepVerifySharedImageSource.Run() = %v, want %v", got, tt.want)
			}
			d, _ := state.GetOk("error")
			err, _ := d.(error)
			if tt.wantErr != "" {
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Wanted error %q, got %q", tt.wantErr, err)
				}
			} else if err != nil && err.Error() != "" {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
