package vagrant

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/packer"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

type VMwareBoxConfig struct {
	OutputPath          string `mapstructure:"output"`
	VagrantfileTemplate string `mapstructure:"vagrantfile_template"`

	PackerBuildName string `mapstructure:"packer_build_name"`
}

type VMwareBoxPostProcessor struct {
	config VMwareBoxConfig
}

func (p *VMwareBoxPostProcessor) Configure(raws ...interface{}) error {
	for _, raw := range raws {
		err := mapstructure.Decode(raw, &p.config)
		if err != nil {
			return err
		}
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
		src, err := os.Open(path)
		if err != nil {
			return nil, false, err
		}
		defer src.Close()

		dst, err := os.Create(filepath.Join(dir, filepath.Base(path)))
		if err != nil {
			return nil, false, err
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
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

		t := template.Must(template.New("vagrantfile").Parse(string(contents)))
		t.Execute(vf, new(struct{}))
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
