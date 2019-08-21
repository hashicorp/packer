// vagrant_cloud implements the packer.PostProcessor interface and adds a
// post-processor that uploads artifacts from the vagrant post-processor
// and vagrant builder to Vagrant Cloud (vagrantcloud.com) or manages
// self hosted boxes on the Vagrant Cloud
package vagrantcloud

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	"packer.post-processor.artifice":   "artifice",
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

func (p *PostProcessor) PostProcess(ctx context.Context, ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, bool, error) {
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

	// Determine the name of the provider for Vagrant Cloud, and Vagrant
	providerName, err := getProvider(artifact.Id(), artifact.Files()[0], builtins[artifact.BuilderId()])

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
	p.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, false, false, rawErr.(error)
	}

	return NewArtifact(providerName, p.config.Tag), true, false, nil
}

func getProvider(builderName, boxfile, builderId string) (providerName string, err error) {
	if builderId == "artifice" {
		// The artifice post processor cannot embed any data in the
		// supplied artifact so the provider information must be extracted
		// from the box file directly
		providerName, err = providerFromVagrantBox(boxfile)
	} else {
		// For the Vagrant builder and Vagrant post processor the provider can
		// be determined from information embedded in the artifact
		providerName = providerFromBuilderName(builderName)
	}
	return providerName, err
}

// Converts a packer builder name to the corresponding vagrant provider
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

// Returns the Vagrant provider the box is intended for use with by
// reading the metadata file packaged inside the box
func providerFromVagrantBox(boxfile string) (providerName string, err error) {
	f, err := os.Open(boxfile)
	if err != nil {
		return "", fmt.Errorf("Error attempting to open box file: %s", err)
	}
	defer f.Close()

	// Vagrant boxes are gzipped tar archives
	ar, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("Error unzipping box archive: %s", err)
	}
	tr := tar.NewReader(ar)

	// The metadata.json file in the tar archive contains a 'provider' key
	type metadata struct {
		ProviderName string `json:"provider"`
	}
	md := metadata{}

	// Loop through the files in the archive and read the provider
	// information from the boxes metadata.json file
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			if md.ProviderName == "" {
				return "", fmt.Errorf("Error: Provider info was not found in box: %s", boxfile)
			}
			break
		}
		if err != nil {
			return "", fmt.Errorf("Error reading header info from box tar archive: %s", err)
		}

		if hdr.Name == "metadata.json" {
			contents, err := ioutil.ReadAll(tr)
			if err != nil {
				return "", fmt.Errorf("Error reading contents of metadata.json file from box file: %s", err)
			}
			err = json.Unmarshal(contents, &md)
			if err != nil {
				return "", fmt.Errorf("Error parsing metadata.json file: %s", err)
			}
			if md.ProviderName == "" {
				return "", fmt.Errorf("Error: Could not determine Vagrant provider from box metadata.json file")
			}
			break
		}
	}
	return md.ProviderName, nil
}
