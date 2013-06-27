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
}

type Config struct {
	OutputPath string `mapstructure:"output"`
}

type PostProcessor struct {
	config  Config
	premade map[string]packer.PostProcessor
}

func (p *PostProcessor) Configure(raw interface{}) error {
	err := mapstructure.Decode(raw, &p.config)
	if err != nil {
		return err
	}

	if p.config.OutputPath == "" {
		p.config.OutputPath = "packer_{{.Provider}}.box"
	}

	_, err = template.New("output").Parse(p.config.OutputPath)
	if err != nil {
		return fmt.Errorf("output invalid template: %s", err)
	}

	mapConfig, ok := raw.(map[string]interface{})
	if !ok {
		panic("Raw configuration not a map")
	}

	errors := make([]error, 0)
	for k, raw := range mapConfig {
		pp := keyToPostProcessor(k)
		if pp == nil {
			continue
		}

		if err := pp.Configure(raw); err != nil {
			errors = append(errors, err)
		}

		p.premade[k] = pp
	}

	if len(errors) > 0 {
		return &packer.MultiError{errors}
	}

	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, error) {
	ppName, ok := builtins[artifact.BuilderId()]
	if !ok {
		return nil, fmt.Errorf("Unknown artifact type, can't build box: %s", artifact.BuilderId())
	}

	// Use the premade PostProcessor if we have one. Otherwise, we
	// create it and configure it here.
	pp, ok := p.premade[ppName]
	if !ok {
		log.Printf("Premade post-processor for '%s' not found. Creating.", ppName)
		pp = keyToPostProcessor(ppName)
		if pp == nil {
			return nil, fmt.Errorf("Vagrant box post-processor not found: %s", ppName)
		}

		config := map[string]string{"output": p.config.OutputPath}
		if err := pp.Configure(config); err != nil {
			return nil, err
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
	default:
		return nil
	}
}
