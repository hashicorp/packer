package vagrant

import (
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
	"path/filepath"
)

type VMwareBoxConfig struct {
	common.PackerConfig `mapstructure:",squash"`

	OutputPath          string `mapstructure:"output"`
	VagrantfileTemplate string `mapstructure:"vagrantfile_template"`

	tpl *common.Template
}

type VMwareBoxPostProcessor struct {
	config VMwareBoxConfig
}

func (p *VMwareBoxPostProcessor) Configure(raws ...interface{}) error {
	md, err := common.DecodeConfig(&p.config, raws...)
	if err != nil {
		return err
	}

	p.config.tpl, err = common.NewTemplate()
	if err != nil {
		return err
	}
	p.config.tpl.UserVars = p.config.PackerUserVars

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)

	templates := map[string]*string{
		"output": &p.config.OutputPath,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = p.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	validates := map[string]*string{
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

func (p *VMwareBoxPostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	// Compile the output path
	outputPath, err := ProcessOutputPath(p.config.OutputPath,
		p.config.PackerBuildName, "vmware", artifact)
	if err != nil {
		return nil, false, err
	}

	// Create a temporary directory for us to build the contents of the box in
	dir, err := ioutil.TempDir("", "packer")
	if err != nil {
		return nil, false, err
	}
	defer os.RemoveAll(dir)

	// Copy all of the original contents into the temporary directory
	for _, path := range artifact.Files() {
		ui.Message(fmt.Sprintf("Copying: %s", path))

		dstPath := filepath.Join(dir, filepath.Base(path))
		if err := CopyContents(dstPath, path); err != nil {
			return nil, false, err
		}
	}

	if p.config.VagrantfileTemplate != "" {
		f, err := os.Open(p.config.VagrantfileTemplate)
		if err != nil {
			return nil, false, err
		}
		defer f.Close()

		contents, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, false, err
		}

		// Create the Vagrantfile from the template
		vf, err := os.Create(filepath.Join(dir, "Vagrantfile"))
		if err != nil {
			return nil, false, err
		}
		defer vf.Close()

		vagrantfileContents, err := p.config.tpl.Process(string(contents), nil)
		if err != nil {
			return nil, false, fmt.Errorf("Error writing Vagrantfile: %s", err)
		}
		vf.Write([]byte(vagrantfileContents))
		vf.Close()
	}

	// Create the metadata
	metadata := map[string]string{"provider": "vmware_desktop"}
	if err := WriteMetadata(dir, metadata); err != nil {
		return nil, false, err
	}

	// Compress the directory to the given output path
	ui.Message(fmt.Sprintf("Compressing box..."))
	if err := DirToBox(outputPath, dir); err != nil {
		return nil, false, err
	}

	return NewArtifact("vmware", outputPath), false, nil
}
