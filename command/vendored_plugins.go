package command

import (
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"

	// Previously core-bundled components, split into their own plugins but
	// still vendored with Packer for now. Importing as library instead of
	// forcing use of packer init, until packer v1.8.0
	dockerbuilder "github.com/hashicorp/packer-plugin-docker/builder/docker"
	dockerimportpostprocessor "github.com/hashicorp/packer-plugin-docker/post-processor/docker-import"
	dockerpushpostprocessor "github.com/hashicorp/packer-plugin-docker/post-processor/docker-push"
	dockersavepostprocessor "github.com/hashicorp/packer-plugin-docker/post-processor/docker-save"
	dockertagpostprocessor "github.com/hashicorp/packer-plugin-docker/post-processor/docker-tag"
)

// VendoredBuilders are builder components that were once bundle with Packer core, but are now being shim with there multi-component counterparts.
var VendoredBuilders = map[string]packersdk.Builder{
	"docker": new(dockerbuilder.Builder),
}

// VendoredProvisioners are components that were once bundle with Packer core, but are now being shim with there multi-component counterparts.
var VendoredProvisioners = map[string]packersdk.Provisioner{}

// VendoredPostProcessors are components that were once bundle with Packer core, but are now being shim with there multi-component counterparts.
var VendoredPostProcessors = map[string]packersdk.PostProcessor{
	"docker-import": new(dockerimportpostprocessor.PostProcessor),
	"docker-push":   new(dockerpushpostprocessor.PostProcessor),
	"docker-save":   new(dockersavepostprocessor.PostProcessor),
	"docker-tag":    new(dockertagpostprocessor.PostProcessor),
}

// Upon init lets load up any plugins that were vendored manually into the default
// set of plugins.
func init() {
	for k, v := range VendoredBuilders {
		if _, ok := Builders[k]; ok {
			continue
		}
		Builders[k] = v
	}

	for k, v := range VendoredProvisioners {
		if _, ok := Provisioners[k]; ok {
			continue
		}
		Provisioners[k] = v
	}

	for k, v := range VendoredPostProcessors {
		if _, ok := PostProcessors[k]; ok {
			continue
		}
		PostProcessors[k] = v
	}
}
