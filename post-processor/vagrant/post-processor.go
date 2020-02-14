//go:generate mapstructure-to-hcl2 -type Config

// vagrant implements the packer.PostProcessor interface and adds a
// post-processor that turns artifacts of known builders into Vagrant
// boxes.
package vagrant

import (
	"compress/flate"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer/tmp"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/mitchellh/mapstructure"
)

var builtins = map[string]string{
	"mitchellh.amazonebs":                 "aws",
	"mitchellh.amazon.instance":           "aws",
	"mitchellh.virtualbox":                "virtualbox",
	"mitchellh.vmware":                    "vmware",
	"mitchellh.vmware-esx":                "vmware",
	"pearkes.digitalocean":                "digitalocean",
	"packer.googlecompute":                "google",
	"hashicorp.scaleway":                  "scaleway",
	"packer.parallels":                    "parallels",
	"MSOpenTech.hyperv":                   "hyperv",
	"transcend.qemu":                      "libvirt",
	"ustream.lxc":                         "lxc",
	"Azure.ResourceManagement.VMImage":    "azure",
	"packer.post-processor.docker-import": "docker",
	"packer.post-processor.docker-tag":    "docker",
	"packer.post-processor.docker-push":   "docker",
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	CompressionLevel             int      `mapstructure:"compression_level"`
	Include                      []string `mapstructure:"include"`
	OutputPath                   string   `mapstructure:"output"`
	Override                     map[string]interface{}
	VagrantfileTemplate          string `mapstructure:"vagrantfile_template"`
	VagrantfileTemplateGenerated bool   `mapstructure:"vagrantfile_template_generated"`

	ctx interpolate.Context
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec {
	return p.config.FlatMapstructure().HCL2Spec()
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	if err := p.configureSingle(&p.config, raws...); err != nil {
		return err
	}
	return nil
}

func (p *PostProcessor) PostProcessProvider(name string, provider Provider, ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	config, err := p.specificConfig(name)
	if err != nil {
		return nil, false, err
	}

	err = CreateDummyBox(ui, config.CompressionLevel)
	if err != nil {
		return nil, false, err
	}

	ui.Say(fmt.Sprintf("Creating Vagrant box for '%s' provider", name))

	var generatedData map[interface{}]interface{}
	stateData := artifact.State("generated_data")
	if stateData != nil {
		// Make sure it's not a nil map so we can assign to it later.
		generatedData = stateData.(map[interface{}]interface{})
	}
	// If stateData has a nil map generatedData will be nil
	// and we need to make sure it's not
	if generatedData == nil {
		generatedData = make(map[interface{}]interface{})
	}
	generatedData["ArtifactId"] = artifact.Id()
	generatedData["BuildName"] = config.PackerBuildName
	generatedData["Provider"] = name
	config.ctx.Data = generatedData

	outputPath, err := interpolate.Render(config.OutputPath, &config.ctx)
	if err != nil {
		return nil, false, err
	}

	// Create a temporary directory for us to build the contents of the box in
	dir, err := tmp.Dir("packer")
	if err != nil {
		return nil, false, err
	}
	defer os.RemoveAll(dir)

	// Copy all of the includes files into the temporary directory
	for _, src := range config.Include {
		ui.Message(fmt.Sprintf("Copying from include: %s", src))
		dst := filepath.Join(dir, filepath.Base(src))
		if err := CopyContents(dst, src); err != nil {
			err = fmt.Errorf("Error copying include file: %s\n\n%s", src, err)
			return nil, false, err
		}
	}

	// Run the provider processing step
	vagrantfile, metadata, err := provider.Process(ui, artifact, dir)
	if err != nil {
		return nil, false, err
	}

	// Write the metadata we got
	if err := WriteMetadata(dir, metadata); err != nil {
		return nil, false, err
	}

	// Write our Vagrantfile
	var customVagrantfile string
	if config.VagrantfileTemplate != "" {
		ui.Message(fmt.Sprintf("Using custom Vagrantfile: %s", config.VagrantfileTemplate))
		customBytes, err := ioutil.ReadFile(config.VagrantfileTemplate)
		if err != nil {
			return nil, false, err
		}

		customVagrantfile = string(customBytes)
	}

	f, err := os.Create(filepath.Join(dir, "Vagrantfile"))
	if err != nil {
		return nil, false, err
	}

	t := template.Must(template.New("root").Parse(boxVagrantfileContents))
	err = t.Execute(f, &vagrantfileTemplate{
		ProviderVagrantfile: vagrantfile,
		CustomVagrantfile:   customVagrantfile,
	})
	f.Close()
	if err != nil {
		return nil, false, err
	}

	// Create the box
	if err := DirToBox(outputPath, dir, ui, config.CompressionLevel); err != nil {
		return nil, false, err
	}

	return NewArtifact(name, outputPath), provider.KeepInputArtifact(), nil
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, bool, error) {

	name, ok := builtins[artifact.BuilderId()]
	if !ok {
		return nil, false, false, fmt.Errorf(
			"Unknown artifact type, can't build box: %s", artifact.BuilderId())
	}

	provider := providerForName(name)
	if provider == nil {
		// This shouldn't happen since we hard code all of these ourselves
		panic(fmt.Sprintf("bad provider name: %s", name))
	}

	artifact, keep, err := p.PostProcessProvider(name, provider, ui, artifact)

	// In some cases, (e.g. AMI), deleting the input artifact would render the
	// resulting vagrant box useless. Because of these cases, we want to
	// forcibly set keep_input_artifact.

	// TODO: rework all provisioners to only forcibly keep those where it matters
	return artifact, keep, true, err
}

func (p *PostProcessor) configureSingle(c *Config, raws ...interface{}) error {
	var md mapstructure.Metadata
	err := config.Decode(c, &config.DecodeOpts{
		Metadata:           &md,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"output",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	// Defaults
	if c.OutputPath == "" {
		c.OutputPath = "packer_{{ .BuildName }}_{{.Provider}}.box"
	}

	found := false
	for _, k := range md.Keys {
		if k == "compression_level" {
			found = true
			break
		}
	}

	if !found {
		c.CompressionLevel = flate.DefaultCompression
	}

	var errs *packer.MultiError
	if c.VagrantfileTemplate != "" && c.VagrantfileTemplateGenerated == false {
		_, err := os.Stat(c.VagrantfileTemplate)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf(
				"vagrantfile_template '%s' does not exist", c.VagrantfileTemplate))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *PostProcessor) specificConfig(name string) (Config, error) {
	config := p.config
	if _, ok := config.Override[name]; ok {
		if err := mapstructure.Decode(config.Override[name], &config); err != nil {
			err = fmt.Errorf("Error overriding config for %s: %s", name, err)
			return config, err
		}
	}
	return config, nil
}

func providerForName(name string) Provider {
	switch name {
	case "aws":
		return new(AWSProvider)
	case "scaleway":
		return new(ScalewayProvider)
	case "digitalocean":
		return new(DigitalOceanProvider)
	case "virtualbox":
		return new(VBoxProvider)
	case "vmware":
		return new(VMwareProvider)
	case "parallels":
		return new(ParallelsProvider)
	case "hyperv":
		return new(HypervProvider)
	case "libvirt":
		return new(LibVirtProvider)
	case "google":
		return new(GoogleProvider)
	case "lxc":
		return new(LXCProvider)
	case "azure":
		return new(AzureProvider)
	case "docker":
		return new(DockerProvider)
	default:
		return nil
	}
}

type vagrantfileTemplate struct {
	ProviderVagrantfile string
	CustomVagrantfile   string
}

const boxVagrantfileContents string = `
# The contents below were provided by the Packer Vagrant post-processor
{{ .ProviderVagrantfile }}

# The contents below (if any) are custom contents provided by the
# Packer template during image build.
{{ .CustomVagrantfile }}
`
