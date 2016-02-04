package brkt

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/brkt/brkt-sdk-go/brkt"
	"github.com/mitchellh/multistep"
)

func TestStepGetMachineType_Implements(t *testing.T) {
	var raw interface{}
	raw = new(stepGetMachineType)
	if _, ok := raw.(multistep.Step); !ok {
		t.Fatalf("stepGetMachineType should implement the multistep.Step interface")
	}
}

func TestStepDeployInstance_RunWithId(t *testing.T) {
	// Setup
	const MACHINE_TYPE_ID = "machine_type_uuid"

	// This handler in addition to returning mock data, runs assertions to
	// verify that the requests sent from the client are valid.
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method + " " + r.URL.Path {
		case "GET /v1/api/config/machinetype/" + MACHINE_TYPE_ID:
			machineType := &brkt.MachineTypeData{Id: MACHINE_TYPE_ID}
			bytes, err := json.Marshal(machineType)
			if err != nil {
				t.Fatalf("should be possible to marshal brkt.MachineTypeData")
			}

			fmt.Fprintf(w, string(bytes))

		default:
			t.Fatalf("Invalid request executed: %+v", r)
			w.WriteHeader(http.StatusBadRequest)
		}
	}

	ts, _, state := prepareState(handler)
	defer ts.Close()

	step := &stepGetMachineType{
		MachineType: MACHINE_TYPE_ID,
	}

	// Run
	action := step.Run(state)

	// Assert everything executed correctly
	if action != multistep.ActionContinue {
		t.Fatalf("did not get ActionContinue back from step")
	}

	if _, ok := state.GetOk("error"); ok {
		t.Fatalf("error should not be set on successful execution")
	}

	machineTypeInterface, ok := state.GetOk("machineType")
	if !ok {
		t.Fatalf("machineType should be set on successful execution")
	}

	machineType, ok := machineTypeInterface.(string)
	if !ok {
		t.Fatalf("machineType should be a string")
	}
	if machineType != MACHINE_TYPE_ID {
		t.Fatalf("machineType should be set to MACHINE_TYPE_ID")
	}
}

func TestStepDeployInstance_RunWithIdFail(t *testing.T) {
	// Setup
	const MACHINE_TYPE_ID = "machine_type_uuid"

	// This handler in addition to returning mock data, runs assertions to
	// verify that the requests sent from the client are valid.
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method + " " + r.URL.Path {
		case "GET /v1/api/config/machinetype/" + MACHINE_TYPE_ID:
			w.WriteHeader(http.StatusNotFound)

		default:
			t.Fatalf("Invalid request executed: %+v", r)
			w.WriteHeader(http.StatusBadRequest)
		}
	}

	ts, _, state := prepareState(handler)
	defer ts.Close()

	step := &stepGetMachineType{
		MachineType: MACHINE_TYPE_ID,
	}

	// Run
	action := step.Run(state)

	// Assert everything executed correctly
	if multistep.ActionHalt != action {
		t.Fatalf("did not get ActionHalt back from step")
	}

	if _, ok := state.GetOk("error"); !ok {
		t.Fatalf("error should be set on failed execution")
	}

	if _, ok := state.GetOk("machineType"); ok {
		t.Fatalf("machineType should not be set on failed execution")
	}
}

var MACHINE_TYPES []*brkt.MachineTypeData = []*brkt.MachineTypeData{
	&brkt.MachineTypeData{
		Id: "cheap_but_not_matching_machine_type",

		CpuCores:   1,
		Ram:        1.5,
		HourlyCost: "0.1",
	},
	&brkt.MachineTypeData{
		Id: "matching_machine_type",

		CpuCores:   4,
		Ram:        4,
		HourlyCost: "0.2",
	},
	&brkt.MachineTypeData{
		Id: "expensive_matching_machine_type",

		CpuCores:   3,
		Ram:        3,
		HourlyCost: "0.25",
	},
}

func TestStepDeployInstance_RunWithMinimums(t *testing.T) {
	// This handler in addition to returning mock data, runs assertions to
	// verify that the requests sent from the client are valid.
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method + " " + r.URL.Path {
		case "GET /v1/api/config/machinetype":
			bytes, err := json.Marshal(MACHINE_TYPES)
			if err != nil {
				t.Fatalf("should be possible to marshal []*brkt.MachineTypeData")
			}

			fmt.Fprintf(w, string(bytes))

		default:
			t.Fatalf("Invalid request executed: %+v", r)
			w.WriteHeader(http.StatusBadRequest)
		}
	}

	ts, _, state := prepareState(handler)
	defer ts.Close()

	step := &stepGetMachineType{
		MinCpuCores: 2,
		MinRam:      2,
	}

	// Run
	action := step.Run(state)

	// Assert everything executed correctly
	if multistep.ActionContinue != action {
		t.Fatalf("did not get ActionContinue back from step")
	}

	if _, ok := state.GetOk("error"); ok {
		t.Fatalf("error should not be set on successful execution")
	}

	machineTypeInterface, ok := state.GetOk("machineType")
	if !ok {
		t.Fatalf("machineType should be set on successful execution")
	}

	machineType, ok := machineTypeInterface.(string)
	if !ok {
		t.Fatalf("machineType should be a string")
	}
	if machineType != "matching_machine_type" {
		t.Fatalf("machineType should be set to `matching_machine_type`")
	}
}

func TestStepDeployInstance_RunWithMinimumsFail(t *testing.T) {
	// This handler in addition to returning mock data, runs assertions to
	// verify that the requests sent from the client are valid.
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method + " " + r.URL.Path {
		case "GET /v1/api/config/machinetype":
			bytes, err := json.Marshal(MACHINE_TYPES)
			if err != nil {
				t.Fatalf("should be possible to marshal []*brkt.MachineTypeData")
			}

			fmt.Fprintf(w, string(bytes))

		default:
			t.Fatalf("Invalid request executed: %+v", r)
			w.WriteHeader(http.StatusBadRequest)
		}
	}

	ts, _, state := prepareState(handler)
	defer ts.Close()

	step := &stepGetMachineType{
		MinCpuCores: 8,
		MinRam:      8,
	}

	// Run
	action := step.Run(state)

	// Assert everything executed correctly
	if action != multistep.ActionHalt {
		t.Fatalf("did not get ActionHalt back from step")
	}

	if _, ok := state.GetOk("error"); !ok {
		t.Fatalf("error should be set on failed execution")
	}
}
