package vagrant

import (
	"fmt"
	"strings"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type AzureProvider struct{}

func (p *AzureProvider) KeepInputArtifact() bool {
	return true
}

func (p *AzureProvider) Process(ui packersdk.Ui, artifact packersdk.Artifact, dir string) (vagrantfile string, metadata map[string]interface{}, err error) {
	// Create the metadata
	metadata = map[string]interface{}{"provider": "azure"}

	var AzureImageProps map[string]string
	AzureImageProps = make(map[string]string)

	// HACK(double16): It appears we can not access the Azure Artifact directly, so parse String()
	artifactString := artifact.String()
	ui.Message(fmt.Sprintf("artifact string: '%s'", artifactString))
	lines := strings.Split(artifactString, "\n")
	for l := 0; l < len(lines); l++ {
		split := strings.Split(lines[l], ": ")
		if len(split) > 1 {
			AzureImageProps[strings.TrimSpace(split[0])] = strings.TrimSpace(split[1])
		}
	}
	ui.Message(fmt.Sprintf("artifact string parsed: %+v", AzureImageProps))

	if AzureImageProps["ManagedImageId"] != "" {
		vagrantfile = fmt.Sprintf(managedImageVagrantfile, AzureImageProps["ManagedImageLocation"], AzureImageProps["ManagedImageId"])
	} else if AzureImageProps["OSDiskUri"] != "" {
		vagrantfile = fmt.Sprintf(vhdVagrantfile, AzureImageProps["StorageAccountLocation"], AzureImageProps["OSDiskUri"], AzureImageProps["OSType"])
	} else {
		err = fmt.Errorf("No managed image nor VHD URI found in artifact: %s", artifactString)
		return
	}
	return
}

var managedImageVagrantfile = `
Vagrant.configure("2") do |config|
	config.vm.provider :azure do |azure, override|
		azure.location = "%s"
		azure.vm_managed_image_id = "%s"
		override.winrm.transport = :ssl
		override.winrm.port = 5986
	end
end
`

var vhdVagrantfile = `
Vagrant.configure("2") do |config|
	config.vm.provider :azure do |azure, override|
		azure.location = "%s"
		azure.vm_vhd_uri = "%s"
		azure.vm_operating_system = "%s"
		override.winrm.transport = :ssl
		override.winrm.port = 5986
	end
end
`
