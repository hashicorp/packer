package oci

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestGetVNIC(t *testing.T) {
	setup()
	defer teardown()

	id := "ocid1.vnic.oc1.phx.a"
	path := fmt.Sprintf("/vnics/%s", id)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"id": "%s"}`, id)
	})

	vnic, err := client.Compute.VNICs.Get(&GetVNICParams{ID: id})
	if err != nil {
		t.Errorf("Client.Compute.VNICs.Get() returned error: %v", err)
	}

	want := &VNIC{ID: id}
	if reflect.DeepEqual(vnic, want) {
		t.Errorf("Client.Compute.VNICs.Get() returned %+v, want %+v", vnic, want)
	}
}
