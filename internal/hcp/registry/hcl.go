package registry

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	sdkpacker "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/hcl2template"
	hcppackerimagedatasource "github.com/hashicorp/packer/internal/hcp/datasource/hcp-packer-image"
	hcppackeriterationdatasource "github.com/hashicorp/packer/internal/hcp/datasource/hcp-packer-iteration"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// HCLMetadataRegistry is a HCP handler made for handling HCL configurations
type HCLMetadataRegistry struct {
	configuration *hcl2template.PackerConfig
	bucket        *Bucket
}

const (
	// Known HCP Packer Image Datasource, whose id is the SourceImageId for some build.
	hcpImageDatasourceType     string = "hcp-packer-image"
	hcpIterationDatasourceType string = "hcp-packer-iteration"
	buildLabel                 string = "build"
)

// PopulateIteration creates the metadata on HCP for a build
func (h *HCLMetadataRegistry) PopulateIteration(ctx context.Context) error {
	err := h.bucket.Initialize(ctx)
	if err != nil {
		return err
	}

	err = h.bucket.populateIteration(ctx)
	if err != nil {
		return err
	}

	iterationID := h.bucket.Iteration.ID

	h.configuration.HCPVars["iterationID"] = cty.StringVal(iterationID)

	return nil
}

// StartBuild is invoked when one build for the configuration is starting to be processed
func (h *HCLMetadataRegistry) StartBuild(ctx context.Context, buildName string) error {
	return h.bucket.startBuild(ctx, buildName)
}

// CompleteBuild is invoked when one build for the configuration has finished
func (h *HCLMetadataRegistry) CompleteBuild(
	ctx context.Context,
	buildName string,
	artifacts []sdkpacker.Artifact,
	buildErr error,
) ([]sdkpacker.Artifact, error) {
	return h.bucket.completeBuild(ctx, buildName, artifacts, buildErr)
}

func NewHCLMetadataRegistry(config *hcl2template.PackerConfig) (*HCLMetadataRegistry, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	if len(config.Builds) > 1 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Multiple " + buildLabel + " blocks",
			Detail: fmt.Sprintf("For Packer Registry enabled builds, only one " + buildLabel +
				" block can be defined. Please remove any additional " + buildLabel +
				" block(s). If this " + buildLabel + " is not meant for the Packer registry please " +
				"clear any HCP_PACKER_* environment variables."),
		})

		return nil, diags
	}

	withHCLBucketConfiguration := func(bb *hcl2template.BuildBlock) bucketConfigurationOpts {
		return func(bucket *Bucket) hcl.Diagnostics {
			bucket.ReadFromHCLBuildBlock(bb)
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
	vals, dsDiags := config.Datasources.Values()
	if dsDiags != nil {
		diags = append(diags, dsDiags...)
	}

	build := config.Builds[0]
	bucket, bucketDiags := createConfiguredBucket(
		config.Basedir,
		withPackerEnvConfiguration,
		withHCLBucketConfiguration(build),
		withDatasourceConfiguration(vals),
	)
	if bucketDiags != nil {
		diags = append(diags, bucketDiags...)
	}

	if diags.HasErrors() {
		return nil, diags
	}

	for _, source := range build.Sources {
		bucket.RegisterBuildForComponent(source.String())
	}

	return &HCLMetadataRegistry{
		configuration: config,
		bucket:        bucket,
	}, nil
}

func imageValueToDSOutput(imageVal map[string]cty.Value) hcppackerimagedatasource.DatasourceOutput {
	dso := hcppackerimagedatasource.DatasourceOutput{}
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

func iterValueToDSOutput(iterVal map[string]cty.Value) hcppackeriterationdatasource.DatasourceOutput {
	dso := hcppackeriterationdatasource.DatasourceOutput{}
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
	return func(bucket *Bucket) hcl.Diagnostics {
		var diags hcl.Diagnostics

		imageDS, imageOK := vals[hcpImageDatasourceType]
		iterDS, iterOK := vals[hcpIterationDatasourceType]

		if !imageOK && !iterOK {
			return nil
		}

		iterations := map[string]hcppackeriterationdatasource.DatasourceOutput{}

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

		images := map[string]hcppackerimagedatasource.DatasourceOutput{}

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
			sourceIteration := ParentIteration{}

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
