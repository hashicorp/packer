package vsphere_tpl

import (
	"fmt"
	"net/url"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)
type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Host                string `mapstructure:"host"`
	Username            string `mapstructure:"username"`
	Password            string `mapstructure:"password"`
	Datacenter          string `mapstructure:"datacenter"`
	VMName              string `mapstructure:"vm_name"`
	Folder              string `mapstructure:"folder"`

	ctx interpolate.Context
}

type PostProcessor struct {
	config Config
	url    *url.URL
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)

	if err != nil {
		return err
	}

	errs := new(packer.MultiError)
	vc := map[string]*string{
		"host":     &p.config.Host,
		"username": &p.config.Username,
		"password": &p.config.Password,
	}

	for key, ptr := range vc {
		if *ptr == "" {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set", key))
		}
	}

	if len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {

	return artifact, true, nil
}

