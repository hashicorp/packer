package registry

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/hashicorp/packer/internal/hcp/env"
	"github.com/hashicorp/packer/packer"
)

// HCPConfigMode types specify the mode in which HCP configuration
// is defined for a given Packer build execution.
type HCPConfigMode int

const (
	// HCPConfigUnset mode is set when no HCP configuration has been found for the Packer execution.
	HCPConfigUnset HCPConfigMode = iota
	// HCPConfigEnabled mode is set when the HCP configuration is codified in the template.
	HCPConfigEnabled
	// HCPEnvEnabled mode is set when the HCP configuration is read from environment variables.
	HCPEnvEnabled
)

type bucketConfigurationOpts func(*Bucket) hcl.Diagnostics

// IsHCPEnabled returns true if HCP integration is enabled for a build
func IsHCPEnabled(cfg packer.Handler) bool {
	// HCP_PACKER_REGISTRY is explicitly turned off
	if env.IsHCPDisabled() {
		return false
	}

	mode := HCPConfigUnset

	switch config := cfg.(type) {
	case *hcl2template.PackerConfig:
		for _, build := range config.Builds {
			if build.HCPPackerRegistry != nil {
				mode = HCPConfigEnabled
			}
		}
	}

	// HCP_PACKER_BUCKET_NAME is set or HCP_PACKER_REGISTRY not toggled off
	if mode == HCPConfigUnset && (env.HasPackerRegistryBucket() || env.IsHCPExplicitelyEnabled()) {
		mode = HCPEnvEnabled
	}

	return mode != HCPConfigUnset
}

// createConfiguredBucket returns a bucket that can be used for connecting to the HCP Packer registry.
// Configuration for the bucket is obtained from the base iteration setting and any addition configuration
// options passed in as opts. All errors during configuration are collected and returned as Diagnostics.
func createConfiguredBucket(templateDir string, opts ...bucketConfigurationOpts) (*Bucket, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	if !env.HasHCPCredentials() {
		diags = append(diags, &hcl.Diagnostic{
			Summary: "HCP authentication information required",
			Detail: fmt.Sprintf("The client authentication requires both %s and %s environment "+
				"variables to be set for authenticating with HCP.",
				env.HCPClientID,
				env.HCPClientSecret),
			Severity: hcl.DiagError,
		})
	}

	bucket := NewBucketWithIteration()

	for _, opt := range opts {
		if optDiags := opt(bucket); optDiags.HasErrors() {
			diags = append(diags, optDiags...)
		}
	}

	if bucket.Slug == "" {
		diags = append(diags, &hcl.Diagnostic{
			Summary: "Image bucket name required",
			Detail: "You must provide an image bucket name for HCP Packer builds. " +
				"You can set the HCP_PACKER_BUCKET_NAME environment variable. " +
				"For HCL2 templates, the registry either uses the name of your " +
				"template's build block, or you can set the bucket_name argument " +
				"in an hcp_packer_registry block.",
			Severity: hcl.DiagError,
		})
	}

	err := bucket.Iteration.Initialize(IterationOptions{
		TemplateBaseDir: templateDir,
	})

	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary: "Iteration initialization failed",
			Detail: fmt.Sprintf("Initialization of the iteration failed with "+
				"the following error message: %s", err),
			Severity: hcl.DiagError,
		})
	}
	return bucket, diags
}

func withPackerEnvConfiguration(bucket *Bucket) hcl.Diagnostics {
	// Add default values for Packer settings configured via EnvVars.
	// TODO look to break this up to be more explicit on what is loaded here.
	bucket.LoadDefaultSettingsFromEnv()

	return nil
}
