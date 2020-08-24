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

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/scaleway/scaleway-cli/pkg/api"
)

const (
	organization = "Beagle Boys"
	token        = "NumberOne"
	userAgent    = "UA"
	region       = "par1"
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
		case "/images":
			var imgs api.ScalewayImages
			for _, name := range fakeImgNames {
				imgs.Images = append(imgs.Images, api.ScalewayImage{
					Identifier: strconv.Itoa(rand.Int()),
					Name:       name,
				})
			}
			enc.Encode(imgs)
		case "/snapshots":
			var snaps api.ScalewaySnapshots
			for _, name := range fakeSnapNames {
				snaps.Snapshots = append(snaps.Snapshots, api.ScalewaySnapshot{
					Identifier: strconv.Itoa(rand.Int()),
					Name:       name,
				})
			}
			enc.Encode(snaps)
		default:
			t.Fatalf("fake server: unexpected path: %q", r.URL.Path)
		}
	}))

	// Ugly but the only way to wire the httptest server to the client...
	api.ComputeAPIPar1 = ts.URL
	api.ComputeAPIAms1 = ts.URL

	client, err := api.NewScalewayAPI(organization, token, userAgent, region)
	if err != nil {
		ts.Close()
		t.Fatalf("setup: client: %s", err)
	}
	client.Logger = api.NewDisableLogger()

	state := multistep.BasicStateBag{}
	state.Put("ui", &packer.BasicUi{
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

func TestGetImages(t *testing.T) {
	state, teardown := setup(t, []string{"image-old-1", "image-old-2"}, []string{})
	defer teardown()
	client := state.Get("client").(*api.ScalewayAPI)

	images, err := getImages(client)
	if err != nil {
		t.Fatalf("getImages: %v", err)
	}
	if len(images.Images) != 2 {
		t.Fatalf("getImages: len(images): want: 2; got: %d", len(images.Images))
	}
}
