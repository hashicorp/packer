// Copyright (c) 2017 Oracle America, Inc.
// The contents of this file are subject to the Mozilla Public License Version
// 2.0 (the "License"); you may not use this file except in compliance with the
// License. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/

package bmcs

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
