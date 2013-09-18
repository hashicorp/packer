// vagrant implements the packer.PostProcessor interface and adds a
// post-processor that turns artifacts of known builders into Vagrant
// boxes.
package vagrant

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"log"
)

var builtins = map[string]string{
	"mitchellh.amazonebs":  "aws",
	"mitchellh.virtualbox": "virtualbox",
	"mitchellh.vmware":     "vmware",
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	OutputPath string `mapstructure:"output"`
}

type PostProcessor struct {
	config      Config
	premade     map[string]packer.PostProcessor
	extraConfig map[string]interface{}
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	_, err := common.DecodeConfig(&p.config, raws...)
	if err != nil {
		return err
	}

	tpl, err := packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	tpl.UserVars = p.config.PackerUserVars

	// Defaults
	if p.config.OutputPath == "" {
		p.config.OutputPath = "packer_{{ .BuildName }}_{{.Provider}}.box"
	}

	// Accumulate any errors
	errs := new(packer.MultiError)
	if err := tpl.Validate(p.config.OutputPath); err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Error parsing output template: %s", err))
	}

	// Store extra configuration we'll send to each post-processor type
	p.extraConfig = make(map[string]interface{})
	p.extraConfig["output"] = p.config.OutputPath

	// TODO(mitchellh): Properly handle multiple raw configs. This isn't
	// very pressing at the moment because at the time of this comment
	// only the first member of raws can contain the actual type-overrides.
	var mapConfig map[string]interface{}
	if err := mapstructure.Decode(raws[0], &mapConfig); err != nil {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("Failed to decode config: %s", err))
		return errs
	}

	p.premade = make(map[string]packer.PostProcessor)
	for k, raw := range mapConfig {
		pp, err := p.subPostProcessor(k, raw, p.extraConfig)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, err)
			continue
		}

		if pp == nil {
			continue
		}

		p.premade[k] = pp
	}

	if len(errs.Errors) > 0 {
		return errs
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

		var err error
		pp, err = p.subPostProcessor(ppName, nil, p.extraConfig)
		if err != nil {
			return nil, false, err
		}

		if pp == nil {
			return nil, false, fmt.Errorf("Vagrant box post-processor not found: %s", ppName)
		}
	}

	ui.Say(fmt.Sprintf("Creating Vagrant box for '%s' provider", ppName))
	return pp.PostProcess(ui, artifact)
}

func (p *PostProcessor) subPostProcessor(key string, specific interface{}, extra map[string]interface{}) (packer.PostProcessor, error) {
	pp := keyToPostProcessor(key)
	if pp == nil {
		return nil, nil
	}

	if err := pp.Configure(extra, specific); err != nil {
		return nil, err
	}

	return pp, nil
}

// keyToPostProcessor maps a configuration key to the actual post-processor
// it will be configuring. This returns a new instance of that post-processor.
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
