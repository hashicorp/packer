package dockerpush

import (
	"fmt"
	"github.com/mitchellh/packer/builder/docker"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/post-processor/docker-import"
	"strings"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	tpl *packer.ConfigTemplate
}

type PostProcessor struct {
	Driver docker.Driver

	config Config
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	_, err := common.DecodeConfig(&p.config, raws...)
	if err != nil {
		return err
	}

	p.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	p.config.tpl.UserVars = p.config.PackerUserVars

	// Accumulate any errors
	errs := new(packer.MultiError)
	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	if artifact.BuilderId() != dockerimport.BuilderId {
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only import from docker-import artifacts.",
			artifact.BuilderId())
		return nil, false, err
	}

	driver := p.Driver
	if driver == nil {
		// If no driver is set, then we use the real driver
		driver = &docker.DockerDriver{Tpl: p.config.tpl, Ui: ui}
	}

	// Get the name. We strip off any tags from the name because the
	// push doesn't use those.
	name := artifact.Id()

	if i := strings.Index(name, "/"); i >= 0 {
		// This should always be true because the / is required. But we have
		// to get the index to this so we don't accidentally strip off the port
		if j := strings.Index(name[i:], ":"); j >= 0 {
			name = name[:i+j]
		}
	}

	ui.Message("Pushing: " + name)
	if err := driver.Push(name); err != nil {
		return nil, false, err
	}

	return nil, false, nil
}
