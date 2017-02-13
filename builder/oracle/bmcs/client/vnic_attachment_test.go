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

func TestListVNICAttachments(t *testing.T) {
	setup()
	defer teardown()

	id := "ocid1.image.oc1.phx.a"
	mux.HandleFunc("/vnicAttachments/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `[{"id":"%s"}]`, id)
	})

	params := &ListVnicAttachmentsParams{InstanceID: id}

	vnicAttachment, err := client.Compute.VNICAttachments.List(params)
	if err != nil {
		t.Errorf("Client.Compute.VNICAttachments.List() returned error: %v", err)
	}

	want := []VNICAttachment{{ID: id}}

	if !reflect.DeepEqual(vnicAttachment, want) {
		t.Errorf("Client.Compute.VNICAttachments.List() returned %+v, want %+v", vnicAttachment, want)
	}
}
