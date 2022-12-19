package registry

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2021-04-30/models"
	sdkpacker "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

// HCLMetadataRegistry is a HCP handler made for handling HCL configurations
type HCLMetadataRegistry struct {
	configuration *hcl2template.PackerConfig
	bucket        *Bucket
	ui            sdkpacker.Ui
}

const (
	// Known HCP Packer Image Datasource, whose id is the SourceImageId for some build.
	hcpImageDatasourceType     string = "hcp-packer-image"
	hcpIterationDatasourceType string = "hcp-packer-iteration"
	buildLabel                 string = "build"
)

// PopulateIteration creates the metadata on HCP for a build
func (h *HCLMetadataRegistry) PopulateIteration(ctx context.Context) error {
	err := h.bucket.Initialize(ctx, models.HashicorpCloudPackerIterationTemplateTypeHCL2)
	if err != nil {
		return err
	}

	err = h.bucket.populateIteration(ctx)
	if err != nil {
		return err
	}

	iterationID := h.bucket.Iteration.ID

	h.configuration.HCPVars["iterationID"] = cty.StringVal(iterationID)

	sha, err := getGitSHA(h.configuration.Basedir)
	if err != nil {
		log.Printf("failed to get GIT SHA from environment, won't set as build labels")
	} else {
		h.bucket.Iteration.AddSHAToBuildLabels(sha)
	}

	return nil
}

// StartBuild is invoked when one build for the configuration is starting to be processed
func (h *HCLMetadataRegistry) StartBuild(ctx context.Context, build sdkpacker.Build) error {
	name := build.Name()
	cb, ok := build.(*packer.CoreBuild)
	if ok {
		name = cb.Type
	}
	return h.bucket.startBuild(ctx, name)
}

// CompleteBuild is invoked when one build for the configuration has finished
func (h *HCLMetadataRegistry) CompleteBuild(
	ctx context.Context,
	build sdkpacker.Build,
	artifacts []sdkpacker.Artifact,
	buildErr error,
) ([]sdkpacker.Artifact, error) {
	name := build.Name()
	cb, ok := build.(*packer.CoreBuild)
	if ok {
		name = cb.Type
	}
	return h.bucket.completeBuild(ctx, name, artifacts, buildErr)
}

// IterationStatusSummary prints a status report in the UI if the iteration is not yet done
func (h *HCLMetadataRegistry) IterationStatusSummary() {
	h.bucket.Iteration.iterationStatusSummary(h.ui)
}

func NewHCLMetadataRegistry(config *hcl2template.PackerConfig, ui sdkpacker.Ui) (*HCLMetadataRegistry, hcl.Diagnostics) {
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

	ui.Say(fmt.Sprintf("Tracking build on HCP Packer with fingerprint %q", bucket.Iteration.Fingerprint))

	return &HCLMetadataRegistry{
		configuration: config,
		bucket:        bucket,
		ui:            ui,
	}, nil
}

type hcpImage struct {
	ID          string
	ChannelID   string
	IterationID string
}

func imageValueToDSOutput(imageVal map[string]cty.Value) hcpImage {
	image := hcpImage{}
	for k, v := range imageVal {
		switch k {
		case "id":
			image.ID = v.AsString()
		case "channel_id":
			image.ChannelID = v.AsString()
		case "iteration_id":
			image.IterationID = v.AsString()
		}
	}

	return image
}

type hcpIteration struct {
	ID        string
	ChannelID string
}

func iterValueToDSOutput(iterVal map[string]cty.Value) hcpIteration {
	iter := hcpIteration{}
	for k, v := range iterVal {
		switch k {
		case "id":
			iter.ID = v.AsString()
		case "channel_id":
			iter.ChannelID = v.AsString()
		}
	}
	return iter
}

func withDatasourceConfiguration(vals map[string]cty.Value) bucketConfigurationOpts {
	return func(bucket *Bucket) hcl.Diagnostics {
		var diags hcl.Diagnostics

		imageDS, imageOK := vals[hcpImageDatasourceType]
		iterDS, iterOK := vals[hcpIterationDatasourceType]

		if !imageOK && !iterOK {
			return nil
		}

		iterations := map[string]hcpIteration{}

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
				iter := iterValueToDSOutput(iterVals)
				iterations[k] = iter
			}
		}

		images := map[string]hcpImage{}

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
				img := imageValueToDSOutput(imageVals)
				images[k] = img
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
