package vsphere_tpl

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/mitchellh/multistep"
	"github.com/vmware/govmomi"
)

var builtins = map[string]string{
	"mitchellh.vmware-esx": "vmware",
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Host                string `mapstructure:"host"`
	Insecure            bool   `mapstructure:"insecure"`
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

	if err := p.configureURL(); err != nil {
		errs = packer.MultiErrorAppend(
			errs, err)
	}

	if len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	if _, ok := builtins[artifact.BuilderId()]; !ok {
		return nil, false, fmt.Errorf("Unknown artifact type, can't build box: %s", artifact.BuilderId())
	}

	source := ""
	for _, path := range artifact.Files() {
		if strings.HasSuffix(path, ".vmx") {
			source = path
			break
		}
	}
	//We give a vSphere-ESXI 10s to sync
	ui.Message("Waiting 10s for VMWare vSphere to start")
	time.Sleep(10 * time.Second)

	ctx := context.Background()
	c, err := govmomi.NewClient(ctx, p.url, p.config.Insecure)
	if err != nil {
		return artifact, true, fmt.Errorf("Error trying to connect: %s", err)
	}

	state := new(multistep.BasicStateBag)
	state.Put("ui", ui)
	state.Put("context", ctx)
	state.Put("client", c)

	steps := []multistep.Step{
		&StepChooseDatacenter{
			Datacenter: p.config.Datacenter,
		},
		&StepFetchVm{
			VMName: p.config.VMName,
			Source: source,
		},
		&StepCreateFolder{
			Folder: p.config.Folder,
		},
		&StepMarkAsTemplate{},
		&StepMoveTemplate{
			Folder: p.config.Folder,
		},
	}

	runner := &multistep.BasicRunner{Steps: steps}
	runner.Run(state)

	if rawErr, ok := state.GetOk("error"); ok {
		return artifact, true, rawErr.(error)
	}
	return artifact, true, nil
}

func (p *PostProcessor) configureURL() error {
	sdk, err := url.Parse(fmt.Sprintf("https://%v/sdk", p.config.Host))

	if err != nil {
		return nil
	}

	sdk.User = url.UserPassword(p.config.Username, p.config.Password)
	p.url = sdk
	return nil
}
