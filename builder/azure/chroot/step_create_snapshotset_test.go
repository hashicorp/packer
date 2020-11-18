package chroot

import (
	"context"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/go-autorest/autorest"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
)

func TestStepCreateSnapshot_Run(t *testing.T) {
	type fields struct {
		OSDiskSnapshotID         string
		DataDiskSnapshotIDPrefix string
		Location                 string
	}
	tests := []struct {
		name            string
		fields          fields
		diskset         Diskset
		want            multistep.StepAction
		wantSnapshotset Diskset
		expectedPutBody string
	}{
		{
			name: "happy path",
			fields: fields{
				OSDiskSnapshotID: "/subscriptions/1234/resourceGroups/rg/providers/Microsoft.Compute/snapshots/osdisk-snap",
				Location:         "region1",
			},
			diskset: diskset("/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/disks/disk1"),
			expectedPutBody: `{
				"location": "region1",
				"properties": {
					"creationData": {
						"createOption": "Copy",
						"sourceResourceId": "/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/disks/disk1"
					},
					"incremental": false
				}
			}`,
			wantSnapshotset: diskset("/subscriptions/1234/resourceGroups/rg/providers/Microsoft.Compute/snapshots/osdisk-snap"),
		},
		{
			name: "multi disk",
			fields: fields{
				OSDiskSnapshotID:         "/subscriptions/1234/resourceGroups/rg/providers/Microsoft.Compute/snapshots/osdisk-snap",
				DataDiskSnapshotIDPrefix: "/subscriptions/1234/resourceGroups/rg/providers/Microsoft.Compute/snapshots/datadisk-snap",
				Location:                 "region1",
			},
			diskset: diskset(
				"/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/disks/osdisk",
				"/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/disks/datadisk1",
				"/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/disks/datadisk2",
				"/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/disks/datadisk3"),
			wantSnapshotset: diskset(
				"/subscriptions/1234/resourceGroups/rg/providers/Microsoft.Compute/snapshots/osdisk-snap",
				"/subscriptions/1234/resourceGroups/rg/providers/Microsoft.Compute/snapshots/datadisk-snap0",
				"/subscriptions/1234/resourceGroups/rg/providers/Microsoft.Compute/snapshots/datadisk-snap1",
				"/subscriptions/1234/resourceGroups/rg/providers/Microsoft.Compute/snapshots/datadisk-snap2",
			),
		},
		{
			name: "invalid ResourceID",
			fields: fields{
				OSDiskSnapshotID: "notaresourceid",
				Location:         "region1",
			},
			diskset: diskset("/subscriptions/12345/resourceGroups/group1/providers/Microsoft.Compute/disks/disk1"),
			want:    multistep.ActionHalt,
		},
	}
	for _, tt := range tests {
		expectedPutBody := regexp.MustCompile(`[\s\n]`).ReplaceAllString(tt.expectedPutBody, "")

		m := compute.NewSnapshotsClient("subscriptionId")
		m.Sender = autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != "PUT" {
				t.Fatal("Expected only a PUT call")
			}
			if expectedPutBody != "" {
				b, _ := ioutil.ReadAll(r.Body)
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
			SnapshotsClientMock: m,
		})
		state.Put("ui", packer.TestUi(t))
		state.Put(stateBagKey_Diskset, tt.diskset)

		t.Run(tt.name, func(t *testing.T) {
			s := &StepCreateSnapshotset{
				OSDiskSnapshotID:         tt.fields.OSDiskSnapshotID,
				DataDiskSnapshotIDPrefix: tt.fields.DataDiskSnapshotIDPrefix,
				Location:                 tt.fields.Location,
			}
			if got := s.Run(context.TODO(), state); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StepCreateSnapshot.Run() = %v, want %v", got, tt.want)
			}

			if len(tt.wantSnapshotset) > 0 {
				got := state.Get(stateBagKey_Snapshotset).(Diskset)
				if !reflect.DeepEqual(got, tt.wantSnapshotset) {
					t.Errorf("Snapshotset = %v, want %v", got, tt.wantSnapshotset)
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

	s := &StepCreateSnapshotset{
		SkipCleanup: true,
	}
	s.Cleanup(state)
}

func TestStepCreateSnapshot_Cleanup(t *testing.T) {
	m := compute.NewSnapshotsClient("subscriptionId")
	{
		expectedCalls := []string{
			"POST /subscriptions/subscriptionId/resourceGroups/rg/providers/Microsoft.Compute/snapshots/ossnap/endGetAccess",
			"DELETE /subscriptions/subscriptionId/resourceGroups/rg/providers/Microsoft.Compute/snapshots/ossnap",
			"POST /subscriptions/subscriptionId/resourceGroups/rg/providers/Microsoft.Compute/snapshots/datasnap1/endGetAccess",
			"DELETE /subscriptions/subscriptionId/resourceGroups/rg/providers/Microsoft.Compute/snapshots/datasnap1",
			"POST /subscriptions/subscriptionId/resourceGroups/rg/providers/Microsoft.Compute/snapshots/datasnap2/endGetAccess",
			"DELETE /subscriptions/subscriptionId/resourceGroups/rg/providers/Microsoft.Compute/snapshots/datasnap2",
		}

		m.Sender = autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
			got := r.Method + " " + r.URL.Path
			found := false
			for i, call := range expectedCalls {
				if call == got {
					// swap i with last and drop last
					expectedCalls[i] = expectedCalls[len(expectedCalls)-1]
					expectedCalls = expectedCalls[:len(expectedCalls)-1]
					found = true
					break
				}
			}
			if !found {
				t.Errorf("unexpected HTTP call: %v, wanted one of %q", got, expectedCalls)
				return &http.Response{
					Request:    r,
					StatusCode: 599, // 500 is retried
				}, nil
			}
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

	s := &StepCreateSnapshotset{
		SkipCleanup: false,
		snapshots: diskset(
			"/subscriptions/1234/resourceGroups/rg/providers/Microsoft.Compute/snapshots/ossnap",
			"/subscriptions/1234/resourceGroups/rg/providers/Microsoft.Compute/snapshots/datasnap1",
			"/subscriptions/1234/resourceGroups/rg/providers/Microsoft.Compute/snapshots/datasnap2"),
	}
	s.Cleanup(state)
}

func TestStepCreateSnapshotset_Cleanup(t *testing.T) {
	type fields struct {
		OSDiskSnapshotID         string
		DataDiskSnapshotIDPrefix string
		Location                 string
		SkipCleanup              bool
		snapshots                Diskset
	}
	type args struct {
		state multistep.StateBag
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StepCreateSnapshotset{
				OSDiskSnapshotID:         tt.fields.OSDiskSnapshotID,
				DataDiskSnapshotIDPrefix: tt.fields.DataDiskSnapshotIDPrefix,
				Location:                 tt.fields.Location,
				SkipCleanup:              tt.fields.SkipCleanup,
				snapshots:                tt.fields.snapshots,
			}
			s.Cleanup(tt.args.state)
		})
	}
}
