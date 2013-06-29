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
}

type VMwareBoxPostProcessor struct {
	config VMwareBoxConfig
}

func (p *VMwareBoxPostProcessor) Configure(raw interface{}) error {
	err := mapstructure.Decode(raw, &p.config)
	if err != nil {
		return err
	}

	return nil
}

func (p *VMwareBoxPostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, error) {
	// Compile the output path
	outputPath, err := ProcessOutputPath(p.config.OutputPath, "vmware", artifact)
	if err != nil {
		return nil, err
	}

	// Create a temporary directory for us to build the contents of the box in
	dir, err := ioutil.TempDir("", "packer")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	// Copy all of the original contents into the temporary directory
	for _, path := range artifact.Files() {
		ui.Message(fmt.Sprintf("Copying: %s", path))
		src, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer src.Close()

		dst, err := os.Create(filepath.Join(dir, filepath.Base(path)))
		if err != nil {
			return nil, err
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			return nil, err
		}
	}

	if p.config.VagrantfileTemplate != "" {
		f, err := os.Open(p.config.VagrantfileTemplate)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		contents, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}

		// Create the Vagrantfile from the template
		vf, err := os.Create(filepath.Join(dir, "Vagrantfile"))
		if err != nil {
			return nil, err
		}
		defer vf.Close()

		t := template.Must(template.New("vagrantfile").Parse(string(contents)))
		t.Execute(vf, new(struct{}))
		vf.Close()
	}

	// Create the metadata
	metadata := map[string]string{"provider": "vmware_desktop"}
	if err := WriteMetadata(dir, metadata); err != nil {
		return nil, err
	}

	// Compress the directory to the given output path
	ui.Message(fmt.Sprintf("Compressing box..."))
	if err := DirToBox(outputPath, dir); err != nil {
		return nil, err
	}

	return NewArtifact("vmware", outputPath), nil
}
