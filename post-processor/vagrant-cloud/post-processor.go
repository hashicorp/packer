// vagrant_cloud implements the packer.PostProcessor interface and adds a
// post-processor that uploads artifacts from the vagrant post-processor
// and vagrant builder to Vagrant Cloud (vagrantcloud.com) or manages
// self hosted boxes on the Vagrant Cloud
package vagrantcloud

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

var builtins = map[string]string{
	"mitchellh.post-processor.vagrant": "vagrant",
	"vagrant":                          "vagrant",
}

const VAGRANT_CLOUD_URL = "https://vagrantcloud.com/api/v1"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Tag                string `mapstructure:"box_tag"`
	Version            string `mapstructure:"version"`
	VersionDescription string `mapstructure:"version_description"`
	NoRelease          bool   `mapstructure:"no_release"`

	AccessToken           string `mapstructure:"access_token"`
	VagrantCloudUrl       string `mapstructure:"vagrant_cloud_url"`
	InsecureSkipTLSVerify bool   `mapstructure:"insecure_skip_tls_verify"`

	BoxDownloadUrl string `mapstructure:"box_download_url"`

	ctx interpolate.Context
}

type boxDownloadUrlTemplate struct {
	ArtifactId string
	Provider   string
}

type PostProcessor struct {
	config                Config
	client                *VagrantCloudClient
	runner                multistep.Runner
	warnAtlasToken        bool
	insecureSkipTLSVerify bool
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"box_download_url",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	// Default configuration
	if p.config.VagrantCloudUrl == "" {
		p.config.VagrantCloudUrl = VAGRANT_CLOUD_URL
	}

	p.insecureSkipTLSVerify = p.config.InsecureSkipTLSVerify == true && p.config.VagrantCloudUrl != VAGRANT_CLOUD_URL

	if p.config.AccessToken == "" {
		envToken := os.Getenv("VAGRANT_CLOUD_TOKEN")
		if envToken == "" {
			envToken = os.Getenv("ATLAS_TOKEN")
			if envToken != "" {
				p.warnAtlasToken = true
			}
		}
		p.config.AccessToken = envToken
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

	// create the HTTP client
	p.client, err = VagrantCloudClient{}.New(p.config.VagrantCloudUrl, p.config.AccessToken, p.insecureSkipTLSVerify)
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Failed to verify authentication token: %v", err))
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, bool, error) {
	if _, ok := builtins[artifact.BuilderId()]; !ok {
		return nil, false, false, fmt.Errorf(
			"Unknown artifact type, requires box from vagrant post-processor or vagrant builder: %s", artifact.BuilderId())
	}

	// We assume that there is only one .box file to upload
	if !strings.HasSuffix(artifact.Files()[0], ".box") {
		return nil, false, false, fmt.Errorf(
			"Unknown files in artifact, vagrant box is required: %s", artifact.Files())
	}

	if p.warnAtlasToken {
		ui.Message("Warning: Using Vagrant Cloud token found in ATLAS_TOKEN. Please make sure it is correct, or set VAGRANT_CLOUD_TOKEN")
	}

	// The name of the provider for vagrant cloud, and vagrant
	providerName := providerFromBuilderName(artifact.Id())

	p.config.ctx.Data = &boxDownloadUrlTemplate{
		ArtifactId: artifact.Id(),
		Provider:   providerName,
	}
	boxDownloadUrl, err := interpolate.Render(p.config.BoxDownloadUrl, &p.config.ctx)
	if err != nil {
		return nil, false, false, fmt.Errorf("Error processing box_download_url: %s", err)
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
	p.runner = common.NewRunner(steps, p.config.PackerConfig, ui)
	p.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, false, false, rawErr.(error)
	}

	return NewArtifact(providerName, p.config.Tag), true, false, nil
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
	case "scaleway":
		return "scaleway"
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
