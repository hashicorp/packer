package provisioneracc

import (
	"bytes"
	"testing"

	amazonebsbuilder "github.com/hashicorp/packer/builder/amazon/ebs"
	"github.com/hashicorp/packer/command"
	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	fileprovisioner "github.com/hashicorp/packer/provisioner/file"
	"github.com/hashicorp/packer/provisioner/shell"
)

// testCoreConfigBuilder creates a packer CoreConfig that has a file builder
// available. This allows us to test a builder that writes files to disk.
func testCoreConfigBuilder(t *testing.T) *packer.CoreConfig {
	components := packer.ComponentFinder{
		BuilderStore: packersdk.MapOfBuilder{
			"amazon-ebs": func() (packersdk.Builder, error) { return &amazonebsbuilder.Builder{}, nil },
		},
		ProvisionerStore: packersdk.MapOfProvisioner{
			"shell": func() (packersdk.Provisioner, error) { return &shell.Provisioner{}, nil },
			"file":  func() (packersdk.Provisioner, error) { return &fileprovisioner.Provisioner{}, nil },
		},
		PostProcessorStore: packersdk.MapOfPostProcessor{},
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
