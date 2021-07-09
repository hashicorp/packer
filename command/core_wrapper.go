package command

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	packerregistry "github.com/hashicorp/packer/internal/packer_registry"
	"github.com/hashicorp/packer/internal/packer_registry/env"
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

func (c *CoreWrapper) RegistryPublisher() (*packerregistry.Bucket, hcl.Diagnostics) {
	if !env.InPARMode() && (env.HasClientID() && env.HasClientSecret()) {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Summary: "Publishing build artifacts to Packer Artifact Registry not enabled",
				Detail: fmt.Sprintf("Packer has detected HCP client environment variables but one or more of the "+
					"required registry variables are missing. Please check that for the following environment variables "+
					"%q %q", env.HCPPackerRegistry, env.HCPPackerBucket),
				Severity: hcl.DiagWarning,
			},
		}
	}

	if !env.InPARMode() {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Summary: "Publishing build artifacts to Packer Artifact Registry not enabled",
				Detail: "No Packer Registry configuration detected; skipping all publishing steps " +
					"See publishing to a Packer registry for Packer configuration details",
				Severity: hcl.DiagWarning,
			},
		}
	}

	bucket := packerregistry.NewBucketWithIteration(packerregistry.IterationOptions{})
	// JSON templates don't support reading Packer registry data from a config template so we load all config settings from environment variables.
	bucket.Canonicalize()

	err := bucket.Validate()
	if err != nil {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Summary:  "Invalid Packer Artifact Registry configuration",
				Detail:   err.Error(),
				Severity: hcl.DiagError,
			},
		}
	}

	return bucket, nil
}
