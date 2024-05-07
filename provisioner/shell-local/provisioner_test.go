// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package shell

import (
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestProvisioner_impl(t *testing.T) {
	var _ packersdk.Provisioner = new(Provisioner)
}

func TestConfigPrepare(t *testing.T) {
	cases := []struct {
		Key   string
		Value interface{}
		Err   bool
	}{
		{
			"unknown_key",
			"bad",
			true,
		},

		{
			"command",
			nil,
			true,
		},
	}

	for _, tc := range cases {
		raw := testConfig(t)

		if tc.Value == nil {
			delete(raw, tc.Key)
		} else {
			raw[tc.Key] = tc.Value
		}

		var p Provisioner
		err := p.Prepare(raw)
		if tc.Err {
			testConfigErr(t, err, tc.Key)
		} else {
			testConfigOk(t, err)
		}
	}
}

func testConfig(t *testing.T) map[string]interface{} {
	return map[string]interface{}{
		"command": "echo foo",
	}
}

func testConfigErr(t *testing.T, err error, extra string) {
	if err == nil {
		t.Fatalf("should error: %s", extra)
	}
}

func testConfigOk(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
}
