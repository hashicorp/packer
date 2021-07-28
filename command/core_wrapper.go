package command

import (
	"github.com/hashicorp/hcl/v2"
	packerregistry "github.com/hashicorp/packer/internal/packer_registry"
	"github.com/hashicorp/packer/packer"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
)

// CoreWrapper wraps a packer.Core in order to have it's Initialize func return
// a diagnostic.
type CoreWrapper struct {
	*packer.Core
}

func (c *CoreWrapper) Initialize(_ packer.InitializeOptions) hcl.Diagnostics {
	err := c.Core.Initialize()
	if err != nil {
		return hcl.Diagnostics{
			&hcl.Diagnostic{
				Detail:   err.Error(),
				Severity: hcl.DiagError,
			},
		}
	}
	return nil
}

func (c *CoreWrapper) PluginRequirements() (plugingetter.Requirements, hcl.Diagnostics) {
	return nil, hcl.Diagnostics{
		&hcl.Diagnostic{
			Summary:  "Packer init is supported for HCL2 configuration templates only",
			Detail:   "Please manually install plugins or use a HCL2 configuration that will do that for you.",
			Severity: hcl.DiagError,
		},
	}
}

// ConfiguredArtifactMetadataPublisher returns a configured image bucket that can be used for publishing
// build image artifacts to a configured Packer Registry destination.
func (c *CoreWrapper) ConfiguredArtifactMetadataPublisher() (*packerregistry.Bucket, hcl.Diagnostics) {
	bucket := c.Core.GetRegistryBucket()

	// If at this point the bucket is nil, it means the HCP Packer registry is not enabled
	if bucket == nil {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Summary: "Publishing build artifacts to HCP Packer Registry not enabled",
				Detail: "No Packer Registry configuration detected; skipping all publishing steps " +
					"See publishing to a Packer registry for Packer configuration details",
				Severity: hcl.DiagWarning,
			},
		}
	}

	err := bucket.Validate()
	if err != nil {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Summary:  "Invalid HCP Packer Registry configuration",
				Detail:   err.Error(),
				Severity: hcl.DiagError,
			},
		}
	}

	return bucket, nil
}
