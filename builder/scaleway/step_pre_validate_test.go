package scaleway

import (
	"bytes"
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

// 1. Configure a httptest server to return the list of fakeImgNames or fakeSnapNames
//    (depending on the endpoint).
// 2. Instantiate a Scaleway API client and wire it to send requests to the httptest
//    server.
// 3. Return a state (containing the client) ready to be passed to the step.Run() method.
// 4. Return a teardown function meant to be deferred from the test.
func setup(t *testing.T, fakeImgNames []string, fakeSnapNames []string) (*multistep.BasicStateBag, func()) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		switch r.URL.Path {
		case "/instance/v1/zones/fr-par-1/images":
			var imgs instance.ListImagesResponse
			for _, name := range fakeImgNames {
				imgs.Images = append(imgs.Images, &instance.Image{
					ID:   strconv.Itoa(rand.Int()),
					Name: name,
					Zone: "fr-par-1",
				})
			}
			imgs.TotalCount = uint32(len(fakeImgNames))
			if err := enc.Encode(imgs); err != nil {
				t.Fatalf("fake server: encoding reply: %s", err)
			}
		case "/instance/v1/zones/fr-par-1/snapshots":
			var snaps instance.ListSnapshotsResponse
			for _, name := range fakeSnapNames {
				snaps.Snapshots = append(snaps.Snapshots, &instance.Snapshot{
					ID:   strconv.Itoa(rand.Int()),
					Name: name,
					Zone: "fr-par-1",
				})
			}
			snaps.TotalCount = uint32(len(fakeSnapNames))
			if err := enc.Encode(snaps); err != nil {
				t.Fatalf("fake server: encoding reply: %s", err)
			}
		default:
			t.Fatalf("fake server: unexpected path: %q", r.URL.Path)
		}
	}))

	clientOpts := []scw.ClientOption{
		scw.WithDefaultZone(scw.ZoneFrPar1),
		scw.WithAPIURL(ts.URL),
	}

	client, err := scw.NewClient(clientOpts...)
	if err != nil {
		ts.Close()
		t.Fatalf("setup: client: %s", err)
	}

	state := multistep.BasicStateBag{}
	state.Put("ui", &packersdk.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	state.Put("client", client)

	teardown := func() {
		ts.Close()
	}
	return &state, teardown
}

func TestStepPreValidate(t *testing.T) {
	testCases := []struct {
		name          string
		fakeImgNames  []string
		fakeSnapNames []string
		step          stepPreValidate
		wantAction    multistep.StepAction
	}{
		{"happy path: both image name and snapshot name are new",
			[]string{"image-old"},
			[]string{"snapshot-old"},
			stepPreValidate{
				Force:        false,
				ImageName:    "image-new",
				SnapshotName: "snapshot-new",
			},
			multistep.ActionContinue,
		},
		{"want failure: old image name",
			[]string{"image-old"},
			[]string{"snapshot-old"},
			stepPreValidate{
				Force:        false,
				ImageName:    "image-old",
				SnapshotName: "snapshot-new",
			},
			multistep.ActionHalt,
		},
		{"want failure: old snapshot name",
			[]string{"image-old"},
			[]string{"snapshot-old"},
			stepPreValidate{
				Force:        false,
				ImageName:    "image-new",
				SnapshotName: "snapshot-old",
			},
			multistep.ActionHalt,
		},
		{"old image name but force flag",
			[]string{"image-old"},
			[]string{"snapshot-old"},
			stepPreValidate{
				Force:        true,
				ImageName:    "image-old",
				SnapshotName: "snapshot-new",
			},
			multistep.ActionContinue,
		},
		{"old snapshot name but force flag",
			[]string{"image-old"},
			[]string{"snapshot-old"},
			stepPreValidate{
				Force:        true,
				ImageName:    "image-new",
				SnapshotName: "snapshot-old",
			},
			multistep.ActionContinue,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			state, teardown := setup(t, tc.fakeImgNames, tc.fakeSnapNames)
			defer teardown()

			if action := tc.step.Run(context.Background(), state); action != tc.wantAction {
				t.Fatalf("step.Run: want: %v; got: %v", tc.wantAction, action)
			}
		})
	}
}
