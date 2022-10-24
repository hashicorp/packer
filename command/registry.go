package command

import (
	"fmt"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	imgds "github.com/hashicorp/packer/datasource/hcp-packer-image"
	iterds "github.com/hashicorp/packer/datasource/hcp-packer-iteration"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/hashicorp/packer/internal/registry"
	"github.com/hashicorp/packer/internal/registry/env"
	"github.com/hashicorp/packer/packer"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
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

const (
	// Known HCP Packer Image Datasource, whose id is the SourceImageId for some build.
	hcpImageDatasourceType     string = "hcp-packer-image"
	hcpIterationDatasourceType string = "hcp-packer-iteration"
	buildLabel                 string = "build"
)

// TrySetupHCP attempts to configure the HCP-related structures if
// Packer has been configured to publish to a HCP Packer registry.
func TrySetupHCP(cfg packer.Handler) hcl.Diagnostics {
	switch cfg := cfg.(type) {
	case *hcl2template.PackerConfig:
		return setupRegistryForPackerConfig(cfg)
	case *CoreWrapper:
		return setupRegistryForPackerCore(cfg)
	}

	return hcl.Diagnostics{
		&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "unknown Handler type",
			Detail: "TrySetupHCP called with an unknown Handler. " +
				"This is a Packer bug and should be brought to the attention " +
				"of the Packer team, please consider opening an issue for this.",
		},
	}
}

func setupRegistryForPackerConfig(pc *hcl2template.PackerConfig) hcl.Diagnostics {

	// HCP_PACKER_REGISTRY is explicitly turned off
	if env.IsHCPDisabled() {
		return nil
	}

	mode := HCPConfigUnset

	for _, build := range pc.Builds {
		if build.HCPPackerRegistry != nil {
			mode = HCPConfigEnabled
		}
	}

	// HCP_PACKER_BUCKET_NAME is set or  HCP_PACKER_REGISTRY not toggled off
	if mode == HCPConfigUnset && (env.HasPackerRegistryBucket() || env.IsHCPExplicitelyEnabled()) {
		mode = HCPEnvEnabled
	}

	if mode == HCPConfigUnset {
		return nil
	}

	var diags hcl.Diagnostics
	if len(pc.Builds) > 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Multiple " + buildLabel + " blocks",
			Detail: fmt.Sprintf("For Packer Registry enabled builds, only one " + buildLabel +
				" block can be defined. Please remove any additional " + buildLabel +
				" block(s). If this " + buildLabel + " is not meant for the Packer registry please " +
				"clear any HCP_PACKER_* environment variables."),
		})

		return diags
	}

	withHCLBucketConfiguration := func(bb *hcl2template.BuildBlock) bucketConfigurationOpts {
		return func(bucket *registry.Bucket) hcl.Diagnostics {
			bb.HCPPackerRegistry.WriteToBucketConfig(bucket)
			// If at this point the bucket.Slug is still empty,
			// last try is to use the build.Name if present
			if bucket.Slug == "" && bb.Name != "" {
				bucket.Slug = bb.Name
			}

			// If the description is empty, use the one from the build block
			if bucket.Description == "" && bb.Description != "" {
				bucket.Description = bb.Description
			}
			return nil
		}
	}

	// Capture Datasource configuration data
	vals, dsDiags := pc.Datasources.Values()
	if dsDiags != nil {
		diags = append(diags, dsDiags...)
	}

	build := pc.Builds[0]
	pc.Bucket, diags = createConfiguredBucket(
		pc.Basedir,
		withPackerEnvConfiguration,
		withHCLBucketConfiguration(build),
		withDatasourceConfiguration(vals),
	)

	if diags.HasErrors() {
		return diags
	}

	for _, source := range build.Sources {
		pc.Bucket.RegisterBuildForComponent(source.String())
	}

	return diags
}

func setupRegistryForPackerCore(cfg *CoreWrapper) hcl.Diagnostics {
	if env.IsHCPDisabled() {
		return nil
	}

	if !env.HasPackerRegistryBucket() && !env.IsHCPExplicitelyEnabled() {
		return nil
	}

	bucket, diags := createConfiguredBucket(
		filepath.Dir(cfg.Core.Template.Path),
		withPackerEnvConfiguration,
	)

	if diags.HasErrors() {
		return diags
	}

	cfg.Core.Bucket = bucket
	for _, b := range cfg.Core.Template.Builders {
		// Get all builds slated within config ignoring any only or exclude flags.
		cfg.Core.Bucket.RegisterBuildForComponent(packer.HCPName(b))
	}

	return diags
}

type bucketConfigurationOpts func(*registry.Bucket) hcl.Diagnostics

// createConfiguredBucket returns a bucket that can be used for connecting to the HCP Packer registry.
// Configuration for the bucket is obtained from the base iteration setting and any addition configuration
// options passed in as opts. All errors during configuration are collected and returned as Diagnostics.
func createConfiguredBucket(templateDir string, opts ...bucketConfigurationOpts) (*registry.Bucket, hcl.Diagnostics) {
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

	bucket := registry.NewBucketWithIteration()

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

	err := bucket.Iteration.Initialize(registry.IterationOptions{
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

func withPackerEnvConfiguration(bucket *registry.Bucket) hcl.Diagnostics {
	// Add default values for Packer settings configured via EnvVars.
	// TODO look to break this up to be more explicit on what is loaded here.
	bucket.LoadDefaultSettingsFromEnv()

	return nil
}

func imageValueToDSOutput(imageVal map[string]cty.Value) imgds.DatasourceOutput {
	dso := imgds.DatasourceOutput{}
	for k, v := range imageVal {
		switch k {
		case "id":
			dso.ID = v.AsString()
		case "region":
			dso.Region = v.AsString()
		case "labels":
			labels := map[string]string{}
			lbls := v.AsValueMap()
			for k, v := range lbls {
				labels[k] = v.AsString()
			}
			dso.Labels = labels
		case "packer_run_uuid":
			dso.PackerRunUUID = v.AsString()
		case "channel_id":
			dso.ChannelID = v.AsString()
		case "iteration_id":
			dso.IterationID = v.AsString()
		case "build_id":
			dso.BuildID = v.AsString()
		case "created_at":
			dso.CreatedAt = v.AsString()
		case "component_type":
			dso.ComponentType = v.AsString()
		case "cloud_provider":
			dso.CloudProvider = v.AsString()
		}
	}

	return dso
}

func iterValueToDSOutput(iterVal map[string]cty.Value) iterds.DatasourceOutput {
	dso := iterds.DatasourceOutput{}
	for k, v := range iterVal {
		switch k {
		case "author_id":
			dso.AuthorID = v.AsString()
		case "bucket_name":
			dso.BucketName = v.AsString()
		case "complete":
			// For all intents and purposes, cty.Value.True() acts
			// like a AsBool() would.
			dso.Complete = v.True()
		case "created_at":
			dso.CreatedAt = v.AsString()
		case "fingerprint":
			dso.Fingerprint = v.AsString()
		case "id":
			dso.ID = v.AsString()
		case "incremental_version":
			// Maybe when cty provides a good way to AsInt() a cty.Value
			// we can consider implementing this.
		case "updated_at":
			dso.UpdatedAt = v.AsString()
		case "channel_id":
			dso.ChannelID = v.AsString()
		}
	}
	return dso
}

func withDatasourceConfiguration(vals map[string]cty.Value) bucketConfigurationOpts {
	return func(bucket *registry.Bucket) hcl.Diagnostics {
		var diags hcl.Diagnostics

		imageDS, imageOK := vals[hcpImageDatasourceType]
		iterDS, iterOK := vals[hcpIterationDatasourceType]

		if !imageOK && !iterOK {
			return nil
		}

		iterations := map[string]iterds.DatasourceOutput{}

		var err error
		if iterOK {
			hcpData := map[string]cty.Value{}
			err = gocty.FromCtyValue(iterDS, &hcpData)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid HCP datasources",
					Detail:   fmt.Sprintf("Failed to decode hcp_packer_iteration datasources: %s", err),
				})
				return diags
			}

			for k, v := range hcpData {
				iterVals := v.AsValueMap()
				dso := iterValueToDSOutput(iterVals)
				iterations[k] = dso
			}
		}

		images := map[string]imgds.DatasourceOutput{}

		if imageOK {
			hcpData := map[string]cty.Value{}
			err = gocty.FromCtyValue(imageDS, &hcpData)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid HCP datasources",
					Detail:   fmt.Sprintf("Failed to decode hcp_packer_image datasources: %s", err),
				})
				return diags
			}

			for k, v := range hcpData {
				imageVals := v.AsValueMap()
				dso := imageValueToDSOutput(imageVals)
				images[k] = dso
			}
		}

		for _, img := range images {
			sourceIteration := registry.ParentIteration{}

			sourceIteration.IterationID = img.IterationID

			if img.ChannelID != "" {
				sourceIteration.ChannelID = img.ChannelID
			} else {
				for _, it := range iterations {
					if it.ID == img.IterationID {
						sourceIteration.ChannelID = it.ChannelID
						break
					}
				}
			}

			bucket.SourceImagesToParentIterations[img.ID] = sourceIteration
		}

		return diags
	}
}
