// vagrant_cloud implements the packer.PostProcessor interface and adds a
// post-processor that uploads artifacts from the vagrant post-processor
// to Vagrant Cloud (vagrantcloud.com)
package vagrantcloud

import (
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
)

const VAGRANT_CLOUD_URL = "https://vagrantcloud.com"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Tag     string `mapstructure:"box_tag"`
	Version string `mapstructure:"version"`

	AccessToken     string `mapstructure:"access_token"`
	VagrantCloudUrl string `mapstructure:"vagrant_cloud_url"`

	tpl *packer.ConfigTemplate
}

type PostProcessor struct {
	config Config
	client *VagrantCloudClient
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	_, err := common.DecodeConfig(&p.config, raws...)
	if err != nil {
		return err
	}

	p.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	p.config.tpl.UserVars = p.config.PackerUserVars

	// Default configuration
	if p.config.VagrantCloudUrl == "" {
		p.config.VagrantCloudUrl = VAGRANT_CLOUD_URL
	}

	// Accumulate any errors
	errs := new(packer.MultiError)

	// required configuration
	templates := map[string]*string{
		"box_tag":      &p.config.Tag,
		"version":      &p.config.Version,
		"access_token": &p.config.AccessToken,
	}

	for key, ptr := range templates {
		if *ptr == "" {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set", key))
		}
	}

	// Template process
	for key, ptr := range templates {
		*ptr, err = p.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", key, err))
		}
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	config := p.config

	// Only accepts input from the vagrant post-processor
	if artifact.BuilderId() != "mitchellh.post-processor.vagrant" {
		return nil, false, fmt.Errorf(
			"Unknown artifact type, requires box from vagrant post-processor: %s", artifact.BuilderId())
	}

	// The name of the provider for vagrant cloud, and vagrant
	provider := providerFromBuilderName(artifact.Id())
	version := p.config.Version
	tag := p.config.Tag

	// create the HTTP client
	p.client = VagrantCloudClient{}.New(p.config.VagrantCloudUrl, p.config.AccessToken)

	ui.Say(fmt.Sprintf("Verifying box is accessible: %s", tag))

	box, err := p.client.Box(tag)

	if err != nil {
		return nil, false, err
	}

	if box.Tag != tag {
		ui.Say(fmt.Sprintf("Could not verify box is correct: %s", tag))
		return nil, false, err
	}

	ui.Say(fmt.Sprintf("Creating Version %s", version))
	ui.Say(fmt.Sprintf("Creating Provider %s", version))
	ui.Say(fmt.Sprintf("Uploading Box %s", version))
	ui.Say(fmt.Sprintf("Verifying upload %s", version))
	ui.Say(fmt.Sprintf("Releasing version %s", version))

	return NewArtifact(provider, config.Tag), true, nil
}

// Runs a cleanup if the post processor fails to upload
func (p *PostProcessor) Cleanup() {
	// Delete the version
}

// converts a packer builder name to the corresponding vagrant
// provider
func providerFromBuilderName(name string) string {
	switch name {
	case "aws":
		return "aws"
	case "digitalocean":
		return "digitalocean"
	case "virtualbox":
		return "virtualbox"
	case "vmware":
		return "vmware_desktop"
	case "parallels":
		return "parallels"
	default:
		return name
	}
}
