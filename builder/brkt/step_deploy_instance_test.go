package brkt

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/brkt/brkt-sdk-go/brkt"
	"github.com/mitchellh/multistep"
)

const WORKLOAD_JSON = `
{
    "id": "%s",
    "state": "%s",
    "name": "image_provisioning_workload",
    "billing_group": "%s",
    "zone": "%s"
}
`

const CLOUDINIT_JSON = `
{
    "id": "%s"
}
`

func TestStepDeployInstance_Implements(t *testing.T) {
	var raw interface{}
	raw = new(stepDeployInstance)
	if _, ok := raw.(multistep.Step); !ok {
		t.Fatalf("stepCreateImage should be a step")
	}
}

func TestStepDeployInstance_Run(t *testing.T) {
	// Setup
	const INSTANCE_ID = "instance_uuid"
	const MACHINE_TYPE_ID = "machine_type_uuid"
	const BILLING_GROUP_ID = "billing_group_uuid"
	const ZONE_ID = "zone_uuid"
	const IMAGE_DEFINITION_ID = "image_definition_uuid"
	const WORKLOAD_ID = "workload_id"
	const CLOUDINIT_ID = "cloud_config_id"

	// This handler in addition to returning mock data, runs assertions to
	// verify that the requests sent from the client are valid.
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method + " " + r.URL.Path {
		case "POST /v1/api/config/cloudinit":
			fmt.Fprintf(w, CLOUDINIT_JSON, CLOUDINIT_ID)

		case "POST /v2/api/config/workload":
			fmt.Fprintf(w, WORKLOAD_JSON, WORKLOAD_ID, "INITIALIZING", BILLING_GROUP_ID, ZONE_ID)

		case "POST /v2/api/config/instance":
			fmt.Fprintf(w, INSTANCE_JSON, INSTANCE_ID, "INITIALIZING")

		case "POST /v1/api/config/instance/" + INSTANCE_ID:
			fmt.Fprintf(w, INSTANCE_JSON, INSTANCE_ID, "INITIALIZING")

		case "GET /v2/api/config/instance/" + INSTANCE_ID:
			fmt.Fprintf(w, INSTANCE_JSON, INSTANCE_ID, "INITIALIZING")

		default:
			t.Fatalf("Invalid request executed: %+v", r)
			w.WriteHeader(http.StatusBadRequest)
		}
	}

	ts, _, state := prepareState(handler)
	defer ts.Close()

	state.Put("machineType", MACHINE_TYPE_ID)

	step := &stepDeployInstance{
		ImageDefinition: IMAGE_DEFINITION_ID,
		BillingGroup:    BILLING_GROUP_ID,
		Zone:            ZONE_ID,
		CloudConfig:     make(map[string]interface{}),
	}

	// Run
	action := step.Run(state)

	// Assert everything executed correctly
	if action != multistep.ActionContinue {
		t.Fatalf("did not get ActionContinue back from step")
	}

	if err, ok := state.GetOk("error"); ok {
		fmt.Sprintf("error should not be set on successful execution, got: %s", err)
	}

	// Assert it stores internal state for later cleanup
	if step.workload == nil {
		t.Fatalf("step.workload must be set")
	}
	if step.cloudInit == nil {
		t.Fatalf("step.cloudInit must be set")
	}
	if step.instance == nil {
		t.Fatalf("step.instance must be set")
	}

	// Assert that it creates an instance that can be used to create an image
	instanceInterface, ok := state.GetOk("instance")
	if !ok {
		t.Fatalf("instance not set")
	}

	instance, ok := instanceInterface.(*brkt.Instance)
	if !ok {
		t.Fatalf("instance should be of type *brkt.Instance")
	}
	if instance.Data.Id != INSTANCE_ID {
		t.Fatalf("instanceId not set correctly")
	}
}

func TestStepDeployInstance_Cleanup(t *testing.T) {
	// Setup
	const INSTANCE_ID = "instance_uuid"
	const MACHINE_TYPE_ID = "machine_type_uuid"
	const BILLING_GROUP_ID = "billing_group_uuid"
	const ZONE_ID = "zone_uuid"
	const IMAGE_DEFINITION_ID = "image_definition_uuid"
	const WORKLOAD_ID = "workload_id"
	const CLOUDINIT_ID = "cloud_config_id"

	deletedWorkload := false
	deletedCloudInit := false

	// This handler in addition to returning mock data, runs assertions to
	// verify that the requests sent from the client are valid.
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method + " " + r.URL.Path {
		case "DELETE /v1/api/config/workload/" + WORKLOAD_ID:
			if deletedWorkload {
				t.Fatalf("should only try to delete workload once")
			}
			deletedWorkload = true
			fmt.Fprintf(w, WORKLOAD_JSON, WORKLOAD_ID, "TERMINATED", BILLING_GROUP_ID, ZONE_ID) // success!

		case "DELETE /v1/api/config/cloudinit/" + CLOUDINIT_ID:
			if deletedCloudInit {
				t.Fatalf("should only try to delete CloudInit once")
			}
			deletedCloudInit = true
			fmt.Fprintf(w, CLOUDINIT_JSON, CLOUDINIT_ID)

		default:
			t.Fatalf("Invalid request executed: %+v", r)
			w.WriteHeader(http.StatusBadRequest)
		}
	}

	ts, api, state := prepareState(handler)
	defer ts.Close()

	step := &stepDeployInstance{
		workload: &brkt.Workload{
			Data: &brkt.WorkloadData{
				Id: WORKLOAD_ID,
			},
			ApiClient: api.ApiClient,
		},
		cloudInit: &brkt.CloudInit{
			Data: &brkt.CloudInitData{
				Id: CLOUDINIT_ID,
			},
			ApiClient: api.ApiClient,
		},
	}

	// Run
	step.Cleanup(state)

	// Assert everything executed correctly
	if err, ok := state.GetOk("error"); ok {
		fmt.Sprintf("error should not be set on successful execution, got: %s", err)
	}

	if !deletedCloudInit {
		t.Fatalf("did not delete cloudInit")
	}
	if !deletedWorkload {
		t.Fatalf("did not delete workload")
	}
}
