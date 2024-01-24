package registry

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type hcpVersion struct {
	VersionID string
	ChannelID string
}

func versionValueToDSOutput(iterVal map[string]cty.Value) hcpVersion {
	version := hcpVersion{}
	for k, v := range iterVal {
		switch k {
		case "id":
			version.VersionID = v.AsString()
		case "channel_id":
			version.ChannelID = v.AsString()
		}
	}
	return version
}

type hcpArtifact struct {
	ExternalIdentifier string
	ChannelID          string
	VersionID          string
}

func artifactValueToDSOutput(imageVal map[string]cty.Value) hcpArtifact {
	artifact := hcpArtifact{}
	for k, v := range imageVal {
		switch k {
		case "external_identifier":
			artifact.ExternalIdentifier = v.AsString()
		case "channel_id":
			artifact.ChannelID = v.AsString()
		case "version_id":
			artifact.VersionID = v.AsString()
		}
	}

	return artifact
}

func withDatasourceConfiguration(vals map[string]cty.Value) bucketConfigurationOpts {
	return func(bucket *Bucket) hcl.Diagnostics {
		var diags hcl.Diagnostics

		versionDS, versionOK := vals[hcpVersionDatasourceType]
		artifactDS, artifactOK := vals[hcpArtifactDatasourceType]

		if !artifactOK && !versionOK {
			return nil
		}

		versions := map[string]hcpVersion{}

		var err error
		if versionOK {
			hcpData := map[string]cty.Value{}
			err = gocty.FromCtyValue(versionDS, &hcpData)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid HCP datasources",
					Detail: fmt.Sprintf(
						"Failed to decode hcp-packer-version datasources: %s", err,
					),
				})
				return diags
			}

			for k, v := range hcpData {
				versionVals := v.AsValueMap()
				version := versionValueToDSOutput(versionVals)
				versions[k] = version
			}
		}

		artifacts := map[string]hcpArtifact{}

		if artifactOK {
			hcpData := map[string]cty.Value{}
			err = gocty.FromCtyValue(artifactDS, &hcpData)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid HCP datasources",
					Detail: fmt.Sprintf(
						"Failed to decode hcp-packer-artifact datasources: %s", err,
					),
				})
				return diags
			}

			for k, v := range hcpData {
				artifactVals := v.AsValueMap()
				artifact := artifactValueToDSOutput(artifactVals)
				artifacts[k] = artifact
			}
		}

		for _, a := range artifacts {
			parentVersion := ParentVersion{}
			parentVersion.VersionID = a.VersionID

			if a.ChannelID != "" {
				parentVersion.ChannelID = a.ChannelID
			} else {
				for _, v := range versions {
					if v.VersionID == a.VersionID {
						parentVersion.ChannelID = v.ChannelID
						break
					}
				}
			}

			bucket.SourceExternalIdentifierToParentVersions[a.ExternalIdentifier] = parentVersion
		}

		return diags
	}
}
