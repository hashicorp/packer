package packer

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestCoreConfig(t *testing.T) *CoreConfig {
	// Create some test components
	components := ComponentFinder{
		Builder: func(n string) (Builder, error) {
			if n != "test" {
				return nil, nil
			}

			return &MockBuilder{}, nil
		},
	}

	return &CoreConfig{
		Components: components,
	}
}

func TestCore(t *testing.T, c *CoreConfig) *Core {
	core, err := NewCore(c)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return core
}

func TestUi(t *testing.T) Ui {
	var buf bytes.Buffer
	return &BasicUi{
		Reader:      &buf,
		Writer:      ioutil.Discard,
		ErrorWriter: ioutil.Discard,
	}
}

// TestBuilder sets the builder with the name n to the component finder
// and returns the mock.
func TestBuilder(t *testing.T, c *CoreConfig, n string) *MockBuilder {
	var b MockBuilder

	c.Components.Builder = func(actual string) (Builder, error) {
		if actual != n {
			return nil, nil
		}

		return &b, nil
	}

	return &b
}

// TestProvisioner sets the prov. with the name n to the component finder
// and returns the mock.
func TestProvisioner(t *testing.T, c *CoreConfig, n string) *MockProvisioner {
	var b MockProvisioner

	c.Components.Provisioner = func(actual string) (Provisioner, error) {
		if actual != n {
			return nil, nil
		}

		return &b, nil
	}

	return &b
}

// TestPostProcessor sets the prov. with the name n to the component finder
// and returns the mock.
func TestPostProcessor(t *testing.T, c *CoreConfig, n string) *MockPostProcessor {
	var b MockPostProcessor

	c.Components.PostProcessor = func(actual string) (PostProcessor, error) {
		if actual != n {
			return nil, nil
		}

		return &b, nil
	}

	return &b
}
