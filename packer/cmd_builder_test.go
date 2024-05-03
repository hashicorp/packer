// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package packer

import (
	"os/exec"
	"testing"
)

func TestBuilder_NoExist(t *testing.T) {
	c := NewClient(&PluginClientConfig{Cmd: exec.Command("i-should-not-exist")})
	defer c.Kill()

	_, err := c.Builder()
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilder_Good(t *testing.T) {
	c := NewClient(&PluginClientConfig{Cmd: helperProcess("builder")})
	defer c.Kill()

	_, err := c.Builder()
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}
