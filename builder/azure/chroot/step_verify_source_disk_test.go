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
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func Test_StepVerifySourceDisk_Run(t *testing.T) {
	type fields struct {
		SourceDiskResourceID string
		Location             string

		GetDiskResponseCode int
		GetDiskResponseBody string
	}
	type args struct {
		state multistep.StateBag
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		want       multistep.StepAction
		errormatch string
	}{
		{
			name: "HappyPath",
			fields: fields{
				SourceDiskResourceID: "/subscriptions/subid1/resourcegroups/rg1/providers/Microsoft.Compute/disks/disk1",
				Location:             "westus2",

				GetDiskResponseCode: 200,
				GetDiskResponseBody: `{"location":"westus2"}`,
			},
			want: multistep.ActionContinue,
		},
		{
			name: "NotAResourceID",
			fields: fields{
				SourceDiskResourceID: "/other",
				Location:             "westus2",

				GetDiskResponseCode: 200,
				GetDiskResponseBody: `{"location":"westus2"}`,
			},
			want:       multistep.ActionHalt,
			errormatch: "Could not parse resource id",
		},
		{
			name: "DiskNotFound",
			fields: fields{
				SourceDiskResourceID: "/subscriptions/subid1/resourcegroups/rg1/providers/Microsoft.Compute/disks/disk1",
				Location:             "westus2",

				GetDiskResponseCode: 404,
				GetDiskResponseBody: `{}`,
			},
			want:       multistep.ActionHalt,
			errormatch: "Unable to retrieve",
		},
		{
			name: "NotADisk",
			fields: fields{
				SourceDiskResourceID: "/subscriptions/subid1/resourcegroups/rg1/providers/Microsoft.Compute/images/image1",
				Location:             "westus2",

				GetDiskResponseCode: 404,
			},
			want:       multistep.ActionHalt,
			errormatch: "not a managed disk",
		},
		{
			name: "OtherSubscription",
			fields: fields{
				SourceDiskResourceID: "/subscriptions/subid2/resourcegroups/rg1/providers/Microsoft.Compute/disks/disk1",
				Location:             "westus2",

				GetDiskResponseCode: 200,
				GetDiskResponseBody: `{"location":"westus2"}`,
			},
			want:       multistep.ActionHalt,
			errormatch: "different subscription",
		},
		{
			name: "OtherLocation",
			fields: fields{
				SourceDiskResourceID: "/subscriptions/subid1/resourcegroups/rg1/providers/Microsoft.Compute/disks/disk1",
				Location:             "eastus",

				GetDiskResponseCode: 200,
				GetDiskResponseBody: `{"location":"westus2"}`,
			},
			want:       multistep.ActionHalt,
			errormatch: "different location",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := StepVerifySourceDisk{
				SourceDiskResourceID: tt.fields.SourceDiskResourceID,
				Location:             tt.fields.Location,
			}

			m := compute.NewDisksClient("subscriptionId")
			m.Sender = autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					Request:    r,
					Body:       ioutil.NopCloser(strings.NewReader(tt.fields.GetDiskResponseBody)),
					StatusCode: tt.fields.GetDiskResponseCode,
				}, nil
			})

			ui, getErr := testUI()

			state := new(multistep.BasicStateBag)
			state.Put("azureclient", &client.AzureClientSetMock{
				DisksClientMock:    m,
				SubscriptionIDMock: "subid1",
			})
			state.Put("ui", ui)

			got := s.Run(context.TODO(), state)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StepVerifySourceDisk.Run() = %v, want %v", got, tt.want)
			}

			if tt.errormatch != "" {
				errs := getErr()
				if !regexp.MustCompile(tt.errormatch).MatchString(errs) {
					t.Errorf("Expected the error output (%q) to match %q", errs, tt.errormatch)
				}
			}

			if got == multistep.ActionHalt {
				if _, ok := state.GetOk("error"); !ok {
					t.Fatal("Expected 'error' to be set in statebag after failure")
				}
			}
		})
	}
}

type uiThatRemebersErrors struct {
	packersdk.Ui
	LastError string
}
