package command

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/packer/builder/docker"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/provisioner/file"
	"github.com/hashicorp/packer/provisioner/shell"
	shell_local "github.com/hashicorp/packer/provisioner/shell-local"
)

func dockerPackerMeta(*testing.T) Meta {
	packerCfg := packer.CoreConfig{}

	packerCfg.Components = packer.ComponentFinder{
		Builder: func(n string) (packer.Builder, error) {
			switch n {
			case "docker":
				return &docker.Builder{}, nil
			default:
				panic(n)
			}
		},
		Provisioner: func(n string) (packer.Provisioner, error) {
			switch n {
			case "shell-local":
				return &shell_local.Provisioner{}, nil
			case "shell":
				return &shell.Provisioner{}, nil
			case "file":
				return &file.Provisioner{}, nil
			default:
				panic(n)
			}
		},
	}
	var out, err bytes.Buffer
	return Meta{
		CoreConfig: &packerCfg,
		Ui: &packer.BasicUi{
			Writer:      &out,
			ErrorWriter: &err,
		},
	}
}

func TestBuildDocker(t *testing.T) {
	if v := os.Getenv("CIRCLECI"); v == "" {
		t.Skipf("CIRCLECI '%s'. skipping", v)
	}

	defer cleanup()

	c := &BuildCommand{
		Meta: dockerPackerMeta(t),
	}

	args := []string{
		filepath.Join(testFixture("docker"), "basic.json"),
	}

	if code := c.Run(args); code != 0 {
		fatalCommand(t, c.Meta)
	}

	for _, f := range []string{"whale.txt"} {
		if !fileExists(f) {
			t.Errorf("Expected to find %s", f)
		}
	}
}
