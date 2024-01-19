package registry

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	sdkpacker "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

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

func withDeprecatedDatasourceConfiguration(vals map[string]cty.Value, ui sdkpacker.Ui) bucketConfigurationOpts {
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
			ui.Say("[WARN] Deprecation: `hcp-packer-iteration` datasource has been deprecated. " +
				"Please use `hcp-packer-version` datasource instead.")
			hcpData := map[string]cty.Value{}
			err = gocty.FromCtyValue(iterDS, &hcpData)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid HCP datasources",
					Detail:   fmt.Sprintf("Failed to decode hcp-packer-iteration datasources: %s", err),
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
			ui.Say("[WARN] Deprecation: `hcp-packer-image` datasource has been deprecated. " +
				"Please use `hcp-packer-artifact` datasource instead.")
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
			sourceIteration := ParentVersion{}

			sourceIteration.VersionID = img.IterationID

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

			bucket.SourceExternalIdentifierToParentVersions[img.ID] = sourceIteration
		}

		return diags
	}
}
