// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package packer

import (
	"context"
	"os/exec"
	"testing"

	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type helperPostProcessor byte

func (helperPostProcessor) ConfigSpec() hcldec.ObjectSpec { return nil }

func (helperPostProcessor) Configure(...interface{}) error {
	return nil
}

func (helperPostProcessor) PostProcess(context.Context, packersdk.Ui, packersdk.Artifact) (packersdk.Artifact, bool, bool, error) {
	return nil, false, false, nil
}

func TestPostProcessor_NoExist(t *testing.T) {
	c := NewClient(&PluginClientConfig{Cmd: exec.Command("i-should-not-exist")})
	defer c.Kill()

	_, err := c.PostProcessor()
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestPostProcessor_Good(t *testing.T) {
	c := NewClient(&PluginClientConfig{Cmd: helperProcess("post-processor")})
	defer c.Kill()

	_, err := c.PostProcessor()
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}
