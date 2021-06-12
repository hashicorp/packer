package hcl2template

import (
	"github.com/hashicorp/hcl/v2"
	packerregistry "github.com/hashicorp/packer/internal/packer_registry"
)

func (cfg *PackerConfig) RegistryPublisher() (*packerregistry.Bucket, hcl.Diagnostics) {
	if cfg.Bucket == nil {
		return nil, hcl.Diagnostics{
			&hcl.Diagnostic{
				Summary: "Publishing build artifacts to Packer Artifact Registry not enabled",
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
				Summary:  "Invalid Packer Artifact Registry configuration",
				Detail:   err.Error(),
				Severity: hcl.DiagError,
			},
		}
	}

	return cfg.Bucket, nil
}
