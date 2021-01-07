package packer

import (
	"bytes"
	"io/ioutil"
	"testing"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer/plugin"
)

func TestCoreConfig(t *testing.T) *CoreConfig {
	// Create some test components
	components := ComponentFinder{
		PluginConfig: &plugin.Config{
			Builders: packersdk.MapOfBuilder{
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
	err := core.Initialize()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return core
}

func TestUi(t *testing.T) packersdk.Ui {
	var buf bytes.Buffer
	return &packersdk.BasicUi{
		Reader:      &buf,
		Writer:      ioutil.Discard,
		ErrorWriter: ioutil.Discard,
	}
}

// TestBuilder sets the builder with the name n to the component finder
// and returns the mock.
func TestBuilder(t *testing.T, c *CoreConfig, n string) *packersdk.MockBuilder {
	var b packersdk.MockBuilder

	c.Components.PluginConfig.Builders = packersdk.MapOfBuilder{
		n: func() (packersdk.Builder, error) { return &b, nil },
	}

	return &b
}

// TestProvisioner sets the prov. with the name n to the component finder
// and returns the mock.
func TestProvisioner(t *testing.T, c *CoreConfig, n string) *packersdk.MockProvisioner {
	var b packersdk.MockProvisioner

	c.Components.PluginConfig.Provisioners = packersdk.MapOfProvisioner{
		n: func() (packersdk.Provisioner, error) { return &b, nil },
	}

	return &b
}

// TestPostProcessor sets the prov. with the name n to the component finder
// and returns the mock.
func TestPostProcessor(t *testing.T, c *CoreConfig, n string) *MockPostProcessor {
	var b MockPostProcessor

	c.Components.PluginConfig.PostProcessors = packersdk.MapOfPostProcessor{
		n: func() (packersdk.PostProcessor, error) { return &b, nil },
	}

	return &b
}
