package packer

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"
)

func TestCoreConfig(t *testing.T) *CoreConfig {
	// Create a UI that is effectively /dev/null everywhere
	var buf bytes.Buffer
	ui := &BasicUi{
		Reader:      &buf,
		Writer:      ioutil.Discard,
		ErrorWriter: ioutil.Discard,
	}

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
		Cache:      &FileCache{CacheDir: os.TempDir()},
		Components: components,
		Ui:         ui,
	}
}

func TestCore(t *testing.T, c *CoreConfig) *Core {
	core, err := NewCore(c)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return core
}
