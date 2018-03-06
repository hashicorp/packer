package arm

import (
	"bytes"
	"fmt"
	"net/url"
	"path"
	"strings"
)

const (
	BuilderId = "Azure.ResourceManagement.VMImage"
)

type AdditionalDiskArtifact struct {
	AdditionalDiskUri            string
	AdditionalDiskUriReadOnlySas string
}

type Artifact struct {
	// VHD
	StorageAccountLocation string
	OSDiskUri              string
	TemplateUri            string
	OSDiskUriReadOnlySas   string
	TemplateUriReadOnlySas string

	// Managed Image
	ManagedImageResourceGroupName string
	ManagedImageName              string
	ManagedImageLocation          string

	// Additional Disks
	AdditionalDisks *[]AdditionalDiskArtifact
}

func NewManagedImageArtifact(resourceGroup, name, location string) (*Artifact, error) {
	return &Artifact{
		ManagedImageResourceGroupName: resourceGroup,
		ManagedImageName:              name,
		ManagedImageLocation:          location,
	}, nil
}

func NewArtifact(template *CaptureTemplate, getSasUrl func(name string) string) (*Artifact, error) {
	if template == nil {
		return nil, fmt.Errorf("nil capture template")
	}

	if len(template.Resources) != 1 {
		return nil, fmt.Errorf("malformed capture template, expected one resource")
	}

	vhdUri, err := url.Parse(template.Resources[0].Properties.StorageProfile.OSDisk.Image.Uri)
	if err != nil {
		return nil, err
	}

	templateUri, err := storageUriToTemplateUri(vhdUri)
	if err != nil {
		return nil, err
	}

	var additional_disks *[]AdditionalDiskArtifact
	if template.Resources[0].Properties.StorageProfile.DataDisks != nil {
		data_disks := make([]AdditionalDiskArtifact, len(template.Resources[0].Properties.StorageProfile.DataDisks))
		for i, additionaldisk := range template.Resources[0].Properties.StorageProfile.DataDisks {
			additionalVhdUri, err := url.Parse(additionaldisk.Image.Uri)
			if err != nil {
				return nil, err
			}
			data_disks[i].AdditionalDiskUri = additionalVhdUri.String()
			data_disks[i].AdditionalDiskUriReadOnlySas = getSasUrl(getStorageUrlPath(additionalVhdUri))
		}
		additional_disks = &data_disks
	}

	return &Artifact{
		OSDiskUri:              vhdUri.String(),
		OSDiskUriReadOnlySas:   getSasUrl(getStorageUrlPath(vhdUri)),
		TemplateUri:            templateUri.String(),
		TemplateUriReadOnlySas: getSasUrl(getStorageUrlPath(templateUri)),

		AdditionalDisks: additional_disks,

		StorageAccountLocation: template.Resources[0].Location,
	}, nil
}

func getStorageUrlPath(u *url.URL) string {
	parts := strings.Split(u.Path, "/")
	return strings.Join(parts[3:], "/")
}

func storageUriToTemplateUri(su *url.URL) (*url.URL, error) {
	// packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd -> 4085bb15-3644-4641-b9cd-f575918640b4
	filename := path.Base(su.Path)
	parts := strings.Split(filename, ".")

	if len(parts) < 3 {
		return nil, fmt.Errorf("malformed URL")
	}

	// packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd -> packer
	prefixParts := strings.Split(parts[0], "-")
	prefix := strings.Join(prefixParts[:len(prefixParts)-1], "-")

	templateFilename := fmt.Sprintf("%s-vmTemplate.%s.json", prefix, parts[1])

	// https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-osDisk.4085bb15-3644-4641-b9cd-f575918640b4.vhd"
	//   ->
	// https://storage.blob.core.windows.net/system/Microsoft.Compute/Images/images/packer-vmTemplate.4085bb15-3644-4641-b9cd-f575918640b4.json"
	return url.Parse(strings.Replace(su.String(), filename, templateFilename, 1))
}

func (a *Artifact) isManagedImage() bool {
	return a.ManagedImageResourceGroupName != ""
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (*Artifact) Files() []string {
	return []string{}
}

func (a *Artifact) Id() string {
	return a.OSDiskUri
}

func (a *Artifact) State(name string) interface{} {
	switch name {
	case "atlas.artifact.metadata":
		return a.stateAtlasMetadata()
	default:
		return nil
	}
}

func (a *Artifact) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%s:\n\n", a.BuilderId()))
	if a.isManagedImage() {
		buf.WriteString(fmt.Sprintf("ManagedImageResourceGroupName: %s\n", a.ManagedImageResourceGroupName))
		buf.WriteString(fmt.Sprintf("ManagedImageName: %s\n", a.ManagedImageName))
		buf.WriteString(fmt.Sprintf("ManagedImageLocation: %s\n", a.ManagedImageLocation))
	} else {
		buf.WriteString(fmt.Sprintf("StorageAccountLocation: %s\n", a.StorageAccountLocation))
		buf.WriteString(fmt.Sprintf("OSDiskUri: %s\n", a.OSDiskUri))
		buf.WriteString(fmt.Sprintf("OSDiskUriReadOnlySas: %s\n", a.OSDiskUriReadOnlySas))
		buf.WriteString(fmt.Sprintf("TemplateUri: %s\n", a.TemplateUri))
		buf.WriteString(fmt.Sprintf("TemplateUriReadOnlySas: %s\n", a.TemplateUriReadOnlySas))
		if a.AdditionalDisks != nil {
			for i, additionaldisk := range *a.AdditionalDisks {
				buf.WriteString(fmt.Sprintf("AdditionalDiskUri (datadisk-%d): %s\n", i+1, additionaldisk.AdditionalDiskUri))
				buf.WriteString(fmt.Sprintf("AdditionalDiskUriReadOnlySas (datadisk-%d): %s\n", i+1, additionaldisk.AdditionalDiskUriReadOnlySas))
			}
		}
	}

	return buf.String()
}

func (*Artifact) Destroy() error {
	return nil
}

func (a *Artifact) stateAtlasMetadata() interface{} {
	metadata := make(map[string]string)
	metadata["StorageAccountLocation"] = a.StorageAccountLocation
	metadata["OSDiskUri"] = a.OSDiskUri
	metadata["OSDiskUriReadOnlySas"] = a.OSDiskUriReadOnlySas
	metadata["TemplateUri"] = a.TemplateUri
	metadata["TemplateUriReadOnlySas"] = a.TemplateUriReadOnlySas

	return metadata
}
