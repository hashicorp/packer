// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package packer

import (
	"bytes"
	"io"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestCoreConfig(t *testing.T) *CoreConfig {
	// Create some test components
	components := ComponentFinder{
		PluginConfig: &PluginConfig{
			Builders: MapOfBuilder{
				"test": func() (packersdk.Builder, error) { return &packersdk.MockBuilder{}, nil },
			},
		},
	}

	return &CoreConfig{
		Components: components,
	}
}

func TestCore(t *testing.T, c *CoreConfig) *Core {
	core := NewCore(c)
	err := core.Initialize(InitializeOptions{})
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return core
}

func TestUi(t *testing.T) packersdk.Ui {
	var buf bytes.Buffer
	return &packersdk.BasicUi{
		Reader:      &buf,
		Writer:      io.Discard,
		ErrorWriter: io.Discard,
	}
}

// TestBuilder sets the builder with the name n to the component finder
// and returns the mock.
func TestBuilder(t *testing.T, c *CoreConfig, n string) *packersdk.MockBuilder {
	var b packersdk.MockBuilder

	c.Components.PluginConfig.Builders = MapOfBuilder{
		n: func() (packersdk.Builder, error) { return &b, nil },
	}

	return &b
}

// TestProvisioner sets the prov. with the name n to the component finder
// and returns the mock.
func TestProvisioner(t *testing.T, c *CoreConfig, n string) *packersdk.MockProvisioner {
	var b packersdk.MockProvisioner

	c.Components.PluginConfig.Provisioners = MapOfProvisioner{
		n: func() (packersdk.Provisioner, error) { return &b, nil },
	}

	return &b
}

// TestPostProcessor sets the prov. with the name n to the component finder
// and returns the mock.
func TestPostProcessor(t *testing.T, c *CoreConfig, n string) *MockPostProcessor {
	var b MockPostProcessor

	c.Components.PluginConfig.PostProcessors = MapOfPostProcessor{
		n: func() (packersdk.PostProcessor, error) { return &b, nil },
	}

	return &b
}
