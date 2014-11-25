// vagrant_cloud implements the packer.PostProcessor interface and adds a
// post-processor that uploads artifacts from the vagrant post-processor
// to Vagrant Cloud (vagrantcloud.com) or manages self hosted boxes on the
// Vagrant Cloud
package vagrantcloud

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"log"
	"strings"
)

const VAGRANT_CLOUD_URL = "https://vagrantcloud.com/api/v1"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Tag                string `mapstructure:"box_tag"`
	Version            string `mapstructure:"version"`
	VersionDescription string `mapstructure:"version_description"`
	NoRelease          bool   `mapstructure:"no_release"`

	AccessToken     string `mapstructure:"access_token"`
	VagrantCloudUrl string `mapstructure:"vagrant_cloud_url"`

	BoxDownloadUrl string `mapstructure:"box_download_url"`

	tpl *packer.ConfigTemplate
}

type boxDownloadUrlTemplate struct {
	ArtifactId string
	Provider   string
}

type PostProcessor struct {
	config Config
	client *VagrantCloudClient
	runner multistep.Runner
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
	// Only accepts input from the vagrant post-processor
	if artifact.BuilderId() != "mitchellh.post-processor.vagrant" {
		return nil, false, fmt.Errorf(
			"Unknown artifact type, requires box from vagrant post-processor: %s", artifact.BuilderId())
	}

	// We assume that there is only one .box file to upload
	if !strings.HasSuffix(artifact.Files()[0], ".box") {
		return nil, false, fmt.Errorf(
			"Unknown files in artifact from vagrant post-processor: %s", artifact.Files())
	}

	// create the HTTP client
	p.client = VagrantCloudClient{}.New(p.config.VagrantCloudUrl, p.config.AccessToken)

	// The name of the provider for vagrant cloud, and vagrant
	providerName := providerFromBuilderName(artifact.Id())

	boxDownloadUrl, err := p.config.tpl.Process(p.config.BoxDownloadUrl, &boxDownloadUrlTemplate{
		ArtifactId: artifact.Id(),
		Provider:   providerName,
	})
	if err != nil {
		return nil, false, fmt.Errorf("Error processing box_download_url: %s", err)
	}

	// Set up the state
	state := new(multistep.BasicStateBag)
	state.Put("config", p.config)
	state.Put("client", p.client)
	state.Put("artifact", artifact)
	state.Put("artifactFilePath", artifact.Files()[0])
	state.Put("ui", ui)
	state.Put("providerName", providerName)
	state.Put("boxDownloadUrl", boxDownloadUrl)

	// Build the steps
	steps := []multistep.Step{}
	if p.config.BoxDownloadUrl == "" {
		steps = []multistep.Step{
			new(stepVerifyBox),
			new(stepCreateVersion),
			new(stepCreateProvider),
			new(stepPrepareUpload),
			new(stepUpload),
			new(stepVerifyUpload),
			new(stepReleaseVersion),
		}
	} else {
		steps = []multistep.Step{
			new(stepVerifyBox),
			new(stepCreateVersion),
			new(stepCreateProvider),
			new(stepReleaseVersion),
		}
	}

	// Run the steps
	if p.config.PackerDebug {
		p.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		p.runner = &multistep.BasicRunner{Steps: steps}
	}

	p.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, false, rawErr.(error)
	}

	return NewArtifact(providerName, p.config.Tag), true, nil
}

// Runs a cleanup if the post processor fails to upload
func (p *PostProcessor) Cancel() {
	if p.runner != nil {
		log.Println("Cancelling the step runner...")
		p.runner.Cancel()
	}
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
