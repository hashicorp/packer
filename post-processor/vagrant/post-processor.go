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

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
)

var builtins = map[string]string{
	"mitchellh.amazonebs":       "aws",
	"mitchellh.amazon.instance": "aws",
	"mitchellh.virtualbox":      "virtualbox",
	"mitchellh.vmware":          "vmware",
	"pearkes.digitalocean":      "digitalocean",
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Include             []string `mapstructure:"include"`
	OutputPath          string   `mapstructure:"output"`
	VagrantfileTemplate string   `mapstructure:"vagrantfile_template"`
	CompressionLevel    int      `mapstructure:"compression_level"`

	tpl *packer.ConfigTemplate
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	md, err := common.DecodeConfig(&p.config, raws...)
	if err != nil {
		return err
	}

	p.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	p.config.tpl.UserVars = p.config.PackerUserVars

	// Defaults
	if p.config.OutputPath == "" {
		p.config.OutputPath = "packer_{{ .BuildName }}_{{.Provider}}.box"
	}

	found := false
	for _, k := range md.Keys {
		if k == "compression_level" {
			found = true
			break
		}
	}

	if !found {
		p.config.CompressionLevel = flate.DefaultCompression
	}

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)

	validates := map[string]*string{
		"output":               &p.config.OutputPath,
		"vagrantfile_template": &p.config.VagrantfileTemplate,
	}

	for n, ptr := range validates {
		if err := p.config.tpl.Validate(*ptr); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error parsing %s: %s", n, err))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
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

	ui.Say(fmt.Sprintf("Creating Vagrant box for '%s' provider", name))

	outputPath, err := p.config.tpl.Process(p.config.OutputPath, &outputPathTemplate{
		ArtifactId: artifact.Id(),
		BuildName:  p.config.PackerBuildName,
		Provider:   name,
	})
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
	for _, src := range p.config.Include {
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
	if p.config.VagrantfileTemplate != "" {
		ui.Message(fmt.Sprintf(
			"Using custom Vagrantfile: %s", p.config.VagrantfileTemplate))
		customBytes, err := ioutil.ReadFile(p.config.VagrantfileTemplate)
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
	if err := DirToBox(outputPath, dir, ui, p.config.CompressionLevel); err != nil {
		return nil, false, err
	}

	return nil, false, nil
}

func providerForName(name string) Provider {
	switch name {
	case "virtualbox":
		return new(VBoxProvider)
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
