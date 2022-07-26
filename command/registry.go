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

const (
	// Known HCP Packer Image Datasource, whose id is the SourceImageId for some build.
	hcpImageDatasourceType     string = "hcp-packer-image"
	hcpIterationDatasourceType string = "hcp-packer-iteration"
	buildLabel                 string = "build"
)

// TrySetupHCP attempts to setup the HCP-related structures if HCP is enabled
// for the command
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
			Detail: "SetupRegistry called with an unknown Handler. " +
				"This is a Packer bug and should be brought to the attention " +
				"of the Packer team, please consider opening an issue for this.",
		},
	}
}

func setupRegistryForPackerConfig(pc *hcl2template.PackerConfig) hcl.Diagnostics {
	var diags hcl.Diagnostics

	hasHCP := false

	for _, build := range pc.Builds {
		if build.HCPPackerRegistry != nil {
			hasHCP = true
		}
	}

	if hasHCP && len(pc.Builds) > 1 {
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

	if env.IsPAREnabled() {
		hasHCP = true
	}

	if !hasHCP {
		return diags
	}

	var err error
	pc.Bucket, err = registry.NewBucketWithIteration(registry.IterationOptions{
		TemplateBaseDir: pc.Basedir,
	})

	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  "Unable to create a valid bucket object for HCP Packer Registry",
			Detail:   fmt.Sprintf("%s", err),
			Severity: hcl.DiagError,
		})

		return diags
	}

	pc.Bucket.LoadDefaultSettingsFromEnv()

	for _, build := range pc.Builds {
		build.HCPPackerRegistry.WriteToBucketConfig(pc.Bucket)

		// If at this point the bucket.Slug is still empty,
		// last try is to use the build.Name if present
		if pc.Bucket.Slug == "" && build.Name != "" {
			pc.Bucket.Slug = build.Name
		}

		// If the description is empty, use the one from the build block
		if pc.Bucket.Description == "" {
			pc.Bucket.Description = build.Description
		}

		for _, source := range build.Sources {
			pc.Bucket.RegisterBuildForComponent(source.String())
		}
	}

	vals, diags := pc.Datasources.Values()
	if diags != nil {
		return diags
	}

	imageDS, imageOK := vals[hcpImageDatasourceType]
	iterDS, iterOK := vals[hcpIterationDatasourceType]

	// If we don't have any image or iteration defined, we can return directly
	if !imageOK && !iterOK {
		return nil
	}

	iterations := map[string]iterds.DatasourceOutput{}

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

		pc.Bucket.SourceImagesToParentIterations[img.ID] = sourceIteration
	}

	return diags
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

func setupRegistryForPackerCore(cfg *CoreWrapper) hcl.Diagnostics {
	if !env.IsPAREnabled() {
		return nil
	}

	var diags hcl.Diagnostics

	var err error

	core := cfg.Core

	core.Bucket, err = registry.NewBucketWithIteration(registry.IterationOptions{
		TemplateBaseDir: filepath.Dir(core.Template.Path),
	})
	if err != nil {
		return append(diags, &hcl.Diagnostic{
			Summary:  "bucket creation failure",
			Detail:   fmt.Sprintf("failed to create Bucket: %s", err),
			Severity: hcl.DiagError,
		})
	}
	core.Bucket.LoadDefaultSettingsFromEnv()

	for _, b := range core.Template.Builders {
		// Get all builds slated within config ignoring any only or exclude flags.
		core.Bucket.RegisterBuildForComponent(b.Name)
	}

	return nil
}
