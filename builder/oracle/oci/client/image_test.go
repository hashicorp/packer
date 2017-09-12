package oci

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestGetImage(t *testing.T) {
	setup()
	defer teardown()

	id := "ocid1.image.oc1.phx.a"
	path := fmt.Sprintf("/images/%s", id)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"id":"%s"}`, id)
	})

	image, err := client.Compute.Images.Get(&GetImageParams{ID: id})
	if err != nil {
		t.Errorf("Client.Compute.Images.Get() returned error: %v", err)
	}

	want := Image{ID: id}

	if !reflect.DeepEqual(image, want) {
		t.Errorf("Client.Compute.Images.Get() returned %+v, want %+v", image, want)
	}
}

func TestCreateImage(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/images/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"displayName": "go-oci test"}`)
	})

	params := &CreateImageParams{
		CompartmentID: "ocid1.compartment.oc1..a",
		DisplayName:   "go-oci test image",
		InstanceID:    "ocid1.image.oc1.phx.a",
	}

	image, err := client.Compute.Images.Create(params)
	if err != nil {
		t.Errorf("Client.Compute.Images.Create() returned error: %v", err)
	}

	want := Image{DisplayName: "go-oci test"}

	if !reflect.DeepEqual(image, want) {
		t.Errorf("Client.Compute.Images.Create() returned %+v, want %+v", image, want)
	}
}

func TestImageGetResourceState(t *testing.T) {
	setup()
	defer teardown()

	id := "ocid1.image.oc1.phx.a"
	path := fmt.Sprintf("/images/%s", id)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"LifecycleState": "AVAILABLE"}`)
	})

	state, err := client.Compute.Images.GetResourceState(id)
	if err != nil {
		t.Errorf("Client.Compute.Images.GetResourceState() returned error: %v", err)
	}

	want := "AVAILABLE"
	if state != want {
		t.Errorf("Client.Compute.Images.GetResourceState() returned %+v, want %+v", state, want)
	}
}

func TestImageGetResourceStateInvalidID(t *testing.T) {
	setup()
	defer teardown()

	id := "ocid1.image.oc1.phx.a"
	path := fmt.Sprintf("/images/%s", id)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"code": "NotAuthorizedOrNotFound"}`)
	})

	state, err := client.Compute.Images.GetResourceState(id)
	if err == nil {
		t.Errorf("Client.Compute.Images.GetResourceState() expected error, got %v", state)
	}

	want := &APIError{Code: "NotAuthorizedOrNotFound"}
	if !reflect.DeepEqual(err, want) {
		t.Errorf("Client.Compute.Images.GetResourceState() errored with %+v, want %+v", err, want)
	}
}

func TestDeleteInstance(t *testing.T) {
	setup()
	defer teardown()

	id := "ocid1.image.oc1.phx.a"
	path := fmt.Sprintf("/images/%s", id)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	err := client.Compute.Images.Delete(&DeleteImageParams{ID: id})
	if err != nil {
		t.Errorf("Client.Compute.Images.Delete() returned error: %v", err)
	}
}
