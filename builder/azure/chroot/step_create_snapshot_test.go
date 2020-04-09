package chroot

import (
	"context"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

func Test_parseSnapshotResourceID(t *testing.T) {

	tests := []struct {
		name               string
		resourceID         string
		wantSubscriptionID string
		wantResourceGroup  string
		wantSnapshotName   string
		wantErr            bool
	}{
		{
			name:               "happy path",
			resourceID:         "/subscriptions/1234/resourceGroups/rg/providers/microsoft.compute/snapshots/disksnapshot1",
			wantErr:            false,
			wantSubscriptionID: "1234",
			wantResourceGroup:  "rg",
			wantSnapshotName:   "disksnapshot1",
		},
		{
			name:       "error: nonsense",
			resourceID: "nonsense",
			wantErr:    true,
		},
		{
			name:       "error: other resource type",
			resourceID: "/subscriptions/1234/resourceGroups/rg/providers/microsoft.compute/disks/disksnapshot1",
			wantErr:    true,
		},
		{
			name:       "error: no name",
			resourceID: "/subscriptions/1234/resourceGroups/rg/providers/microsoft.compute/snapshots",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSubscriptionID, gotResourceGroup, gotSnapshotName, err := parseSnapshotResourceID(tt.resourceID)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSnapshotResourceID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotSubscriptionID != tt.wantSubscriptionID {
				t.Errorf("parseSnapshotResourceID() gotSubscriptionID = %v, want %v", gotSubscriptionID, tt.wantSubscriptionID)
			}
			if gotResourceGroup != tt.wantResourceGroup {
				t.Errorf("parseSnapshotResourceID() gotResourceGroup = %v, want %v", gotResourceGroup, tt.wantResourceGroup)
			}
			if gotSnapshotName != tt.wantSnapshotName {
				t.Errorf("parseSnapshotResourceID() gotSnapshotName = %v, want %v", gotSnapshotName, tt.wantSnapshotName)
			}
		})
	}
}

func TestStepCreateSnapshot_Run(t *testing.T) {
	type fields struct {
		ResourceID string
		Location   string
	}
	tests := []struct {
		name            string
		fields          fields
		want            multistep.StepAction
		wantSnapshotID  string
		expectedPutBody string
	}{
		{
			name: "happy path",
			fields: fields{
				ResourceID: "/subscriptions/1234/resourceGroups/rg/providers/Microsoft.Compute/snapshots/snap1",
				Location:   "region1",
			},
			expectedPutBody: `{
				"location": "region1",
				"properties": {
					"creationData": {
						"createOption": "Copy",
						"sourceResourceId": "osdiskresourceid"
					},
					"incremental": false
				}
			}`,
		},
		{
			name: "invalid ResourceID",
			fields: fields{
				ResourceID: "notaresourceid",
				Location:   "region1",
			},
			want: multistep.ActionHalt,
		},
	}
	for _, tt := range tests {
		expectedPutBody := regexp.MustCompile(`[\s\n]`).ReplaceAllString(tt.expectedPutBody, "")

		m := compute.NewSnapshotsClient("subscriptionId")
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
			SnapshotsClientMock: m,
		})
		state.Put("ui", packer.TestUi(t))
		state.Put(stateBagKey_OSDiskResourceID, "osdiskresourceid")

		t.Run(tt.name, func(t *testing.T) {
			s := &StepCreateSnapshot{
				ResourceID: tt.fields.ResourceID,
				Location:   tt.fields.Location,
			}
			if got := s.Run(context.TODO(), state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StepCreateSnapshot.Run() = %v, want %v", got, tt.want)
			}

			if tt.wantSnapshotID != "" {
				got := state.Get(stateBagKey_OSDiskSnapshotResourceID).(string)
				if !strings.EqualFold(got, tt.wantSnapshotID) {
					t.Errorf("OSDiskSnapshotResourceID = %v, want %v", got, tt.wantSnapshotID)
				}
			}

		})
	}
}

func TestStepCreateSnapshot_Cleanup_skipped(t *testing.T) {
	m := compute.NewSnapshotsClient("subscriptionId")
	m.Sender = autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatalf("clean up should be skipped, did not expect HTTP calls")
		return nil, nil
	})

	state := new(multistep.BasicStateBag)
	state.Put("azureclient", &client.AzureClientSetMock{
		SnapshotsClientMock: m,
	})
	state.Put("ui", packer.TestUi(t))

	s := &StepCreateSnapshot{
		SkipCleanup: true,
	}
	s.Cleanup(state)
}

func TestStepCreateSnapshot_Cleanup(t *testing.T) {
	m := compute.NewSnapshotsClient("subscriptionId")
	{
		expectedCalls := []string{
			"POST /subscriptions/subscriptionId/resourceGroups/rg/providers/Microsoft.Compute/snapshots/snap1/endGetAccess?api-version=2019-07-01",
			"DELETE /subscriptions/subscriptionId/resourceGroups/rg/providers/Microsoft.Compute/snapshots/snap1?api-version=2019-07-01",
		}
		i := 0
		m.Sender = autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
			want := expectedCalls[i]
			got := r.Method + " " + r.URL.RequestURI()
			if want != got {
				t.Errorf("unexpected HTTP call: %v, wanted %v", got, want)
				return &http.Response{
					Request:    r,
					StatusCode: 599, // 500 is retried
				}, nil
			}
			i++
			return &http.Response{
				Request:    r,
				StatusCode: 200,
			}, nil
		})
	}
	state := new(multistep.BasicStateBag)
	state.Put("azureclient", &client.AzureClientSetMock{
		SnapshotsClientMock: m,
	})
	state.Put("ui", packer.TestUi(t))

	s := &StepCreateSnapshot{
		ResourceID:     "/subscriptions/1234/resourceGroups/rg/providers/Microsoft.Compute/snapshots/snap1",
		Location:       "region1",
		SkipCleanup:    false,
		resourceGroup:  "rg",
		snapshotName:   "snap1",
		subscriptionID: "1234",
	}
	s.Cleanup(state)
}
