// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package packer

import (
	"os/exec"
	"testing"
)

func TestProvisioner_NoExist(t *testing.T) {
	c := NewClient(&PluginClientConfig{Cmd: exec.Command("i-should-not-exist")})
	defer c.Kill()

	_, err := c.Provisioner()
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestProvisioner_Good(t *testing.T) {
	c := NewClient(&PluginClientConfig{Cmd: helperProcess("provisioner")})
	defer c.Kill()

	_, err := c.Provisioner()
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}
