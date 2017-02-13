// Copyright (c) 2017 Oracle America, Inc.
// The contents of this file are subject to the Mozilla Public License Version
// 2.0 (the "License"); you may not use this file except in compliance with the
// License. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/

package bmcs

import (
	"errors"
	"testing"

	"github.com/mitchellh/multistep"
)

func TestStepImage(t *testing.T) {
	state := testState()
	state.Put("instance_id", "ocid1...")

	step := new(stepImage)
	defer step.Cleanup(state)

	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	if _, ok := state.GetOk("image"); !ok {
		t.Fatalf("should have image")
	}
}

func TestStepImage_CreateImageErr(t *testing.T) {
	state := testState()
	state.Put("instance_id", "ocid1...")

	step := new(stepImage)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*driverMock)
	driver.CreateImageErr = errors.New("error")

	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	if _, ok := state.GetOk("error"); !ok {
		t.Fatalf("should have error")
	}

	if _, ok := state.GetOk("image"); ok {
		t.Fatalf("should NOT have image")
	}
}

func TestStepImage_WaitForImageCreationErr(t *testing.T) {
	state := testState()
	state.Put("instance_id", "ocid1...")

	step := new(stepImage)
	defer step.Cleanup(state)

	driver := state.Get("driver").(*driverMock)
	driver.WaitForImageCreationErr = errors.New("error")

	if action := step.Run(state); action != multistep.ActionHalt {
		t.Fatalf("bad action: %#v", action)
	}

	if _, ok := state.GetOk("error"); !ok {
		t.Fatalf("should have error")
	}

	if _, ok := state.GetOk("image"); ok {
		t.Fatalf("should not have image")
	}
}
