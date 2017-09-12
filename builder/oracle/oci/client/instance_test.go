package oci

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestGetInstance(t *testing.T) {
	setup()
	defer teardown()

	id := "ocid1.instance.oc1.phx.a"
	path := fmt.Sprintf("/instances/%s", id)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"id":"%s"}`, id)
	})

	instance, err := client.Compute.Instances.Get(&GetInstanceParams{ID: id})
	if err != nil {
		t.Errorf("Client.Compute.Instances.Get() returned error: %v", err)
	}

	want := Instance{ID: id}

	if !reflect.DeepEqual(instance, want) {
		t.Errorf("Client.Compute.Instances.Get() returned %+v, want %+v", instance, want)
	}
}

func TestLaunchInstance(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/instances/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"displayName": "go-oci test"}`)
	})

	params := &LaunchInstanceParams{
		AvailabilityDomain: "aaaa:PHX-AD-1",
		CompartmentID:      "ocid1.compartment.oc1..a",
		DisplayName:        "go-oci test",
		ImageID:            "ocid1.image.oc1.phx.a",
		Shape:              "VM.Standard1.1",
		SubnetID:           "ocid1.subnet.oc1.phx.a",
	}

	instance, err := client.Compute.Instances.Launch(params)
	if err != nil {
		t.Errorf("Client.Compute.Instances.Launch() returned error: %v", err)
	}

	want := Instance{DisplayName: "go-oci test"}

	if !reflect.DeepEqual(instance, want) {
		t.Errorf("Client.Compute.Instances.Launch() returned %+v, want %+v", instance, want)
	}
}

func TestTerminateInstance(t *testing.T) {
	setup()
	defer teardown()

	id := "ocid1.instance.oc1.phx.a"
	path := fmt.Sprintf("/instances/%s", id)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.Compute.Instances.Terminate(&TerminateInstanceParams{ID: id})
	if err != nil {
		t.Errorf("Client.Compute.Instances.Terminate() returned error: %v", err)
	}
}

func TestInstanceGetResourceState(t *testing.T) {
	setup()
	defer teardown()

	id := "ocid1.instance.oc1.phx.a"
	path := fmt.Sprintf("/instances/%s", id)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"LifecycleState": "RUNNING"}`)
	})

	state, err := client.Compute.Instances.GetResourceState(id)
	if err != nil {
		t.Errorf("Client.Compute.Instances.GetResourceState() returned error: %v", err)
	}

	want := "RUNNING"
	if state != want {
		t.Errorf("Client.Compute.Instances.GetResourceState() returned %+v, want %+v", state, want)
	}
}
