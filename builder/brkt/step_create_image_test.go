package brkt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brkt/brkt-sdk-go/brkt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

const MOCK_ACCESS_TOKEN = "ERGH"
const MOCK_MAC_KEY = "BLARGH"

const INSTANCE_ID = "instance_uuid"
const PROVISIONED_IMAGE_DEFINITION_ID = "provisioned_image_uuid"
const IMAGE_NAME = "Magic Instance Image 9000"

// getTestState is a utility function that sets up a BasicStateBag with a
// BasicUi
func getTestState() *multistep.BasicStateBag {
	state := new(multistep.BasicStateBag)
	state.Put("ui", &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})

	return state
}

// unmarshalPayload is a utility function to unmarshal the payloads that get
// sent to the mock server
func unmarshalPayload(r *http.Request, raw interface{}) {
	defer r.Body.Close()

	data, _ := ioutil.ReadAll(r.Body)

	json.Unmarshal(data, raw)
}

const INSTANCE_JSON = `
{
	"id": "%s",
	"provider_instance": {
		"state": "%s"
	},
	"internet_ip_address": "255.255.255.255",
	"ip_address": "10.0.0.1"
}
`

const CREATE_IMAGE_JSON = `
{
	"request_id": "%s"
}
`

const IMAGE_DEFINITION_JSON = `
{
	"id": "%s",
	"name": "%s",
	"state": "%s"
}
`

func TestStepCreateImage_Implements(t *testing.T) {
	var raw interface{}
	raw = new(stepCreateImage)
	if _, ok := raw.(multistep.Step); !ok {
		t.Fatalf("stepCreateImage should be a step")
	}
}

type Handler func(w http.ResponseWriter, r *http.Request)

func prepareState(handler Handler) (*httptest.Server, *brkt.API, *multistep.BasicStateBag) {
	ts := httptest.NewServer(http.HandlerFunc(handler))

	api := brkt.NewAPIForCustomPortal(MOCK_ACCESS_TOKEN, MOCK_MAC_KEY, ts.URL+"/")

	// remove CA verification
	brkt.SetClientTransport(api.ApiClient.Session.Client, true)

	state := getTestState()
	state.Put("api", api)

	return ts, api, state
}

func TestStepCreateImage_Run(t *testing.T) {
	// This handler in addition to returning mock data, runs assertions to
	// verify that the requests sent from the client are valid.
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method + " " + r.URL.Path {
		case "GET /v1/api/config/instance/" + INSTANCE_ID:
			fmt.Fprintf(w, INSTANCE_JSON, INSTANCE_ID, "READY")

		case "POST /v1/api/config/instance/" + INSTANCE_ID + "/createimage":
			payload := &brkt.InstanceCreateImagePayload{}
			unmarshalPayload(r, payload)

			if payload.ImageName != IMAGE_NAME {
				t.Fatalf("ImageName not passed correctly")
			}

			fmt.Fprintf(w, CREATE_IMAGE_JSON, PROVISIONED_IMAGE_DEFINITION_ID)

		case "GET /v1/api/config/imagedefinition/" + PROVISIONED_IMAGE_DEFINITION_ID:
			fmt.Fprintf(w, IMAGE_DEFINITION_JSON, PROVISIONED_IMAGE_DEFINITION_ID, IMAGE_NAME, "READY")

		default:
			t.Fatalf("Invalid request executed: %+v", r)
			w.WriteHeader(http.StatusBadRequest)
		}
	}

	ts, api, state := prepareState(handler)
	defer ts.Close()

	state.Put("instance", &brkt.Instance{
		Data: &brkt.InstanceData{
			Id: INSTANCE_ID,
			ProviderInstance: brkt.InstanceProviderInstanceData{
				State: "INITIALIZING",
			},
		},
		ApiClient: api.ApiClient,
	})

	step := &stepCreateImage{
		ImageName: IMAGE_NAME,
	}

	// Run
	action := step.Run(state)

	// Assert everything executed correctly
	if action != multistep.ActionContinue {
		t.Fatalf("did not get ActionContinue back from step")
	}

	_, ok := state.GetOk("error")
	if ok {
		t.Fatalf("error should not be set on successful execution")
	}

	imageId, ok := state.GetOk("imageId")
	if !ok {
		t.Fatalf("imageId was not set")
	}
	if imageId != PROVISIONED_IMAGE_DEFINITION_ID {
		t.Fatalf("imageId set incorrectly")
	}

	imageName, ok := state.GetOk("imageName")
	if !ok {
		t.Fatalf("imageName not set")
	}
	if imageName != IMAGE_NAME {
		t.Fatalf("imageName set incorrectly")
	}
}

func TestStepCreateImage_Run_FailedImage(t *testing.T) {
	// This handler in addition to returning mock data, runs assertions to
	// verify that the requests sent from the client are valid.
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method + " " + r.URL.Path {
		case "GET /v1/api/config/instance/" + INSTANCE_ID:
			fmt.Fprintf(w, INSTANCE_JSON, INSTANCE_ID, "READY")

		case "POST /v1/api/config/instance/" + INSTANCE_ID + "/createimage":
			payload := &brkt.InstanceCreateImagePayload{}
			unmarshalPayload(r, payload)

			if payload.ImageName != IMAGE_NAME {
				t.Fatalf("ImageName not passed correctly")
			}

			fmt.Fprintf(w, CREATE_IMAGE_JSON, PROVISIONED_IMAGE_DEFINITION_ID)

		case "GET /v1/api/config/imagedefinition/" + PROVISIONED_IMAGE_DEFINITION_ID:
			fmt.Fprintf(w, IMAGE_DEFINITION_JSON, PROVISIONED_IMAGE_DEFINITION_ID, IMAGE_NAME, "FAILED")

		default:
			t.Fatalf("Invalid request executed: %+v", r)
			w.WriteHeader(http.StatusBadRequest)
		}
	}

	ts, api, state := prepareState(handler)
	defer ts.Close()

	state.Put("instance", &brkt.Instance{
		Data: &brkt.InstanceData{
			Id: INSTANCE_ID,
			ProviderInstance: brkt.InstanceProviderInstanceData{
				State: "INITIALIZING",
			},
		},
		ApiClient: api.ApiClient,
	})

	step := &stepCreateImage{
		ImageName: IMAGE_NAME,
	}

	// Run
	action := step.Run(state)

	// Assert everything executed correctly
	if multistep.ActionHalt != action {
		t.Fatalf("did not get ActionHalt back from step")
	}

	if _, ok := state.GetOk("error"); !ok {
		t.Fatalf("error should be set when image definition creation fails")
	}

	if _, ok := state.GetOk("imageId"); ok {
		t.Fatalf("imageId should not be set")
	}

	if _, ok := state.GetOk("imageName"); ok {
		t.Fatalf("imageName should not be set")
	}
}
