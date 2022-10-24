package hcl2template

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	packerregistry "github.com/hashicorp/packer/internal/registry"
)

// ConfiguredArtifactMetadataPublisher returns a configured image bucket that can be used for publishing
// build image artifacts to a configured Packer Registry destination.
func (cfg *PackerConfig) ConfiguredArtifactMetadataPublisher() (*packerregistry.Bucket, hcl.Diagnostics) {
	// If this was a PAR (HCP Packer registry) build either the env. variables are set, or if there is a hcp_packer_registry block
	// defined we would have a non-nil bucket. So if nil assume we are not in a some sort of PAR mode.
	if cfg.Bucket == nil {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Summary: "Publishing build artifacts to HCP Packer Registry not enabled",
				Detail: "No Packer Registry configuration detected; skipping all publishing steps " +
					"See publishing to a Packer registry for Packer configuration details",
				Severity: hcl.DiagWarning,
			},
		}
	}

	err := cfg.Bucket.Validate()
	if err != nil {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Summary: "Invalid HCP Packer configuration",
				Detail: fmt.Sprintf("Packer could not validate the provided "+
					"HCP Packer registry configuration. Check the error message for details "+
					"or contact HCP Packer support for further assistance.\nError: %s", err),
				Severity: hcl.DiagError,
			},
		}
	}

	return cfg.Bucket, nil
}
