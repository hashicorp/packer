// vagrant implements the packer.PostProcessor interface and adds a
// post-processor that turns artifacts of known builders into Vagrant
// boxes.
package vagrant

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/packer"
	"log"
	"text/template"
)

var builtins = map[string]string{
	"mitchellh.amazonebs":  "aws",
	"mitchellh.virtualbox": "virtualbox",
	"mitchellh.vmware":     "vmware",
}

type Config struct {
	OutputPath string `mapstructure:"output"`

	PackerBuildName string `mapstructure:"packer_build_name"`
}

type PostProcessor struct {
	config  Config
	premade map[string]packer.PostProcessor
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	for _, raw := range raws {
		err := mapstructure.Decode(raw, &p.config)
		if err != nil {
			return err
		}
	}

	if p.config.OutputPath == "" {
		p.config.OutputPath = "packer_{{ .BuildName }}_{{.Provider}}.box"
	}

	_, err := template.New("output").Parse(p.config.OutputPath)
	if err != nil {
		return fmt.Errorf("output invalid template: %s", err)
	}

	// TODO(mitchellh): Properly handle multiple raw configs
	var mapConfig map[string]interface{}
	if err := mapstructure.Decode(raws[0], &mapConfig); err != nil {
		return err
	}

	packerConfig := map[string]interface{}{
		packer.BuildNameConfigKey: p.config.PackerBuildName,
	}

	p.premade = make(map[string]packer.PostProcessor)
	errors := make([]error, 0)
	for k, raw := range mapConfig {
		pp := keyToPostProcessor(k)
		if pp == nil {
			continue
		}

		if err := pp.Configure(raw, packerConfig); err != nil {
			errors = append(errors, err)
		}

		p.premade[k] = pp
	}

	if len(errors) > 0 {
		return &packer.MultiError{errors}
	}

	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	ppName, ok := builtins[artifact.BuilderId()]
	if !ok {
		return nil, false, fmt.Errorf("Unknown artifact type, can't build box: %s", artifact.BuilderId())
	}

	// Use the premade PostProcessor if we have one. Otherwise, we
	// create it and configure it here.
	pp, ok := p.premade[ppName]
	if !ok {
		log.Printf("Premade post-processor for '%s' not found. Creating.", ppName)
		pp = keyToPostProcessor(ppName)
		if pp == nil {
			return nil, false, fmt.Errorf("Vagrant box post-processor not found: %s", ppName)
		}

		config := map[string]string{"output": p.config.OutputPath}
		if err := pp.Configure(config); err != nil {
			return nil, false, err
		}
	}

	ui.Say(fmt.Sprintf("Creating Vagrant box for '%s' provider", ppName))
	return pp.PostProcess(ui, artifact)
}

func keyToPostProcessor(key string) packer.PostProcessor {
	switch key {
	case "aws":
		return new(AWSBoxPostProcessor)
	case "virtualbox":
		return new(VBoxBoxPostProcessor)
	case "vmware":
		return new(VMwareBoxPostProcessor)
	default:
		return nil
	}
}
