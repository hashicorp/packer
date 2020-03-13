// +build integration

package shell_integration

import (
	"bytes"
	"github.com/hashicorp/go-uuid"
	amazonebsbuilder "github.com/hashicorp/packer/builder/amazon/ebs"
	"github.com/hashicorp/packer/command"
	"github.com/hashicorp/packer/packer"
	fileprovisioner "github.com/hashicorp/packer/provisioner/file"
	"github.com/hashicorp/packer/provisioner/shell"
	testshelper "github.com/hashicorp/packer/test/helper"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildShellProvisionerWithBuildVariablesSharing(t *testing.T) {
	UUID, _ := uuid.GenerateUUID()
	os.Setenv("PACKER_RUN_UUID", UUID)
	c := &command.BuildCommand{
		Meta: testMetaFile(t),
	}

	file := "provisioner.shell." + UUID + ".txt"
	defer os.RemoveAll(file)

	args := []string{
		filepath.Join("./test-fixtures", "shell-provisioner.json"),
	}
	if code := c.Run(args); code != 0 {
		ui := c.Meta.Ui.(*packer.BasicUi)
		out := ui.Writer.(*bytes.Buffer)
		err := ui.ErrorWriter.(*bytes.Buffer)
		t.Fatalf(
			"Bad exit code.\n\nStdout:\n\n%s\n\nStderr:\n\n%s",
			out.String(),
			err.String())
	}

	if _, err := os.Stat(file); err != nil {
		t.Errorf("Expected to find %s", file)
	} else {
		helper := testshelper.AWSHelper{
			Region:  "us-east-1",
			AMIName: "packer-test-shell-interpolate",
		}
		helper.CleanUpAmi(t)
	}
}

func testMetaFile(t *testing.T) command.Meta {
	var out, err bytes.Buffer
	return command.Meta{
		CoreConfig: testCoreConfigBuilder(t),
		Ui: &packer.BasicUi{
			Writer:      &out,
			ErrorWriter: &err,
		},
	}
}

func testCoreConfigBuilder(t *testing.T) *packer.CoreConfig {
	components := packer.ComponentFinder{
		BuilderStore: packer.MapOfBuilder{
			"amazon-ebs": func() (packer.Builder, error) { return &amazonebsbuilder.Builder{}, nil },
		},
		ProvisionerStore: packer.MapOfProvisioner{
			"shell":       func() (packer.Provisioner, error) { return &shell.Provisioner{}, nil },
			"file":       func() (packer.Provisioner, error) { return &fileprovisioner.Provisioner{}, nil },
		},
		PostProcessorStore: packer.MapOfPostProcessor{},
	}
	return &packer.CoreConfig{
		Components: components,
	}
}
