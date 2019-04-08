package command

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/hashicorp/packer/builder/file"
	"github.com/hashicorp/packer/packer"
	shell_local "github.com/hashicorp/packer/provisioner/shell-local"
	"github.com/hashicorp/packer/provisioner/sleep"
)

// testCoreConfigBuilder creates a packer CoreConfig that has a file builder
// available. This allows us to test a builder that writes files to disk.
func testCoreConfigSleepBuilder(t *testing.T) *packer.CoreConfig {
	components := packer.ComponentFinder{
		Builder: func(n string) (packer.Builder, error) {
			switch n {
			case "file":
				return &file.Builder{}, nil
			default:
				panic(n)
			}
		},
		Provisioner: func(n string) (packer.Provisioner, error) {
			switch n {
			case "shell-local":
				return &shell_local.Provisioner{}, nil
			case "sleep":
				return &sleep.Provisioner{}, nil
			default:
				panic(n)
			}
		},
	}
	return &packer.CoreConfig{
		Components: components,
	}
}

// testMetaFile creates a Meta object that includes a file builder
func testMetaSleepFile(t *testing.T) Meta {
	var out, err bytes.Buffer
	return Meta{
		CoreConfig: testCoreConfigSleepBuilder(t),
		Ui: &packer.BasicUi{
			Writer:      &out,
			ErrorWriter: &err,
		},
	}
}

func TestBuildSleepTimeout(t *testing.T) {
	defer cleanup()

	c := &BuildCommand{
		Meta: testMetaSleepFile(t),
	}

	args := []string{
		filepath.Join(testFixture("timeout"), "template.json"),
	}

	defer cleanup()

	if code := c.Run(args); code == 0 {
		fatalCommand(t, c.Meta)
	}

	for _, f := range []string{"roses.txt", "fuchsias.txt", "lilas.txt"} {
		if !fileExists(f) {
			t.Errorf("Expected to find %s", f)
		}
	}

	for _, f := range []string{"campanules.txt"} {
		if fileExists(f) {
			t.Errorf("Expected to not find %s", f)
		}
	}
}
