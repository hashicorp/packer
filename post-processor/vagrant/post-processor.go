// vagrant implements the packer.PostProcessor interface and adds a
// post-processor that turns artifacts of known builders into Vagrant
// boxes.
package vagrant

import (
	"compress/flate"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

var builtins = map[string]string{
	"mitchellh.amazonebs":       "aws",
	"mitchellh.amazon.instance": "aws",
	"mitchellh.virtualbox":      "virtualbox",
	"mitchellh.vmware":          "vmware",
	"pearkes.digitalocean":      "digitalocean",
	"packer.parallels":          "parallels",
	"MSOpenTech.hyperv":         "hyperv",
	"transcend.qemu":            "libvirt",
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	CompressionLevel    int      `mapstructure:"compression_level"`
	Include             []string `mapstructure:"include"`
	OutputPath          string   `mapstructure:"output"`
	Override            map[string]interface{}
	VagrantfileTemplate string `mapstructure:"vagrantfile_template"`

	ctx interpolate.Context
}

type PostProcessor struct {
	configs map[string]*Config
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	p.configs = make(map[string]*Config)
	p.configs[""] = new(Config)
	if err := p.configureSingle(p.configs[""], raws...); err != nil {
		return err
	}

	// Go over any of the provider-specific overrides and load those up.
	for name, override := range p.configs[""].Override {
		subRaws := make([]interface{}, len(raws)+1)
		copy(subRaws, raws)
		subRaws[len(raws)] = override

		config := new(Config)
		p.configs[name] = config
		if err := p.configureSingle(config, subRaws...); err != nil {
			return fmt.Errorf("Error configuring %s: %s", name, err)
		}
	}

	return nil
}

func (p *PostProcessor) PostProcessProvider(name string, provider Provider, ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	config := p.configs[""]
	if specificConfig, ok := p.configs[name]; ok {
		config = specificConfig
	}

	ui.Say(fmt.Sprintf("Creating Vagrant box for '%s' provider", name))

	config.ctx.Data = &outputPathTemplate{
		ArtifactId: artifact.Id(),
		BuildName:  config.PackerBuildName,
		Provider:   name,
	}
	outputPath, err := interpolate.Render(config.OutputPath, &config.ctx)
	if err != nil {
		return nil, false, err
	}

	// Create a temporary directory for us to build the contents of the box in
	dir, err := ioutil.TempDir("", "packer")
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

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {

	name, ok := builtins[artifact.BuilderId()]
	if !ok {
		return nil, false, fmt.Errorf(
			"Unknown artifact type, can't build box: %s", artifact.BuilderId())
	}

	provider := providerForName(name)
	if provider == nil {
		// This shouldn't happen since we hard code all of these ourselves
		panic(fmt.Sprintf("bad provider name: %s", name))
	}

	return p.PostProcessProvider(name, provider, ui, artifact)
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
	if c.VagrantfileTemplate != "" {
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

func providerForName(name string) Provider {
	switch name {
	case "aws":
		return new(AWSProvider)
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
	default:
		return nil
	}
}

// OutputPathTemplate is the structure that is availalable within the
// OutputPath variables.
type outputPathTemplate struct {
	ArtifactId string
	BuildName  string
	Provider   string
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
