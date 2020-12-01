package testshelper

import (
	"bytes"
	"os"
	"testing"

	amazonebsbuilder "github.com/hashicorp/packer/builder/amazon/ebs"
	"github.com/hashicorp/packer/command"
	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	fileprovisioner "github.com/hashicorp/packer/provisioner/file"
	"github.com/hashicorp/packer/provisioner/shell"
)

// fileExists returns true if the filename is found
func FileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}

// testCoreConfigBuilder creates a packer CoreConfig that has a file builder
// available. This allows us to test a builder that writes files to disk.
func testCoreConfigBuilder(t *testing.T) *packer.CoreConfig {
	components := packer.ComponentFinder{
		BuilderStore: packer.MapOfBuilder{
			"amazon-ebs": func() (packersdk.Builder, error) { return &amazonebsbuilder.Builder{}, nil },
		},
		ProvisionerStore: packer.MapOfProvisioner{
			"shell": func() (packersdk.Provisioner, error) { return &shell.Provisioner{}, nil },
			"file":  func() (packersdk.Provisioner, error) { return &fileprovisioner.Provisioner{}, nil },
		},
		PostProcessorStore: packer.MapOfPostProcessor{},
	}
	return &packer.CoreConfig{
		Components: components,
	}
}

// TestMetaFile creates a Meta object that includes a file builder
func TestMetaFile(t *testing.T) command.Meta {
	var out, err bytes.Buffer
	return command.Meta{
		CoreConfig: testCoreConfigBuilder(t),
		Ui: &packersdk.BasicUi{
			Writer:      &out,
			ErrorWriter: &err,
		},
	}
}

func CleanupFiles(moreFiles ...string) {
	for _, file := range moreFiles {
		os.RemoveAll(file)
	}
}
