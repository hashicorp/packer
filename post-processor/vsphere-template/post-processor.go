//go:generate mapstructure-to-hcl2 -type Config

package vsphere_template

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	vmwcommon "github.com/hashicorp/packer/builder/vmware/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/post-processor/vsphere"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/vmware/govmomi"
)

var builtins = map[string]string{
	vsphere.BuilderId:      "vmware",
	vmwcommon.BuilderIdESX: "vmware",
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Host                string `mapstructure:"host"`
	Insecure            bool   `mapstructure:"insecure"`
	Username            string `mapstructure:"username"`
	Password            string `mapstructure:"password"`
	Datacenter          string `mapstructure:"datacenter"`
	Folder              string `mapstructure:"folder"`
	SnapshotEnable      bool   `mapstructure:"snapshot_enable"`
	SnapshotName        string `mapstructure:"snapshot_name"`
	SnapshotDescription string `mapstructure:"snapshot_description"`

	ctx interpolate.Context
}

type PostProcessor struct {
	config Config
	url    *url.URL
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

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

	if p.config.Folder != "" && !strings.HasPrefix(p.config.Folder, "/") {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Folder must be bound to the root"))
	}

	sdk, err := url.Parse(fmt.Sprintf("https://%v/sdk", p.config.Host))
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Error invalid vSphere sdk endpoint: %s", err))
		return errs
	}

	sdk.User = url.UserPassword(p.config.Username, p.config.Password)
	p.url = sdk

	if len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, bool, error) {
	if _, ok := builtins[artifact.BuilderId()]; !ok {
		return nil, false, false, fmt.Errorf("The Packer vSphere Template post-processor "+
			"can only take an artifact from the VMware-iso builder, built on "+
			"ESXi (i.e. remote) or an artifact from the vSphere post-processor. "+
			"Artifact type %s does not fit this requirement", artifact.BuilderId())
	}

	f := artifact.State(vmwcommon.ArtifactConfFormat)
	k := artifact.State(vmwcommon.ArtifactConfKeepRegistered)
	s := artifact.State(vmwcommon.ArtifactConfSkipExport)

	if f != "" && k != "true" && s == "false" {
		return nil, false, false, errors.New("To use this post-processor with exporting behavior you need set keep_registered as true")
	}

	// In some occasions the VM state is powered on and if we immediately try to mark as template
	// (after the ESXi creates it) it will fail. If vSphere is given a few seconds this behavior doesn't reappear.
	ui.Message("Waiting 10s for VMware vSphere to start")
	time.Sleep(10 * time.Second)
	c, err := govmomi.NewClient(context.Background(), p.url, p.config.Insecure)
	if err != nil {
		return nil, false, false, fmt.Errorf("Error connecting to vSphere: %s", err)
	}

	defer c.Logout(context.Background())

	state := new(multistep.BasicStateBag)
	state.Put("ui", ui)
	state.Put("client", c)

	steps := []multistep.Step{
		&stepChooseDatacenter{
			Datacenter: p.config.Datacenter,
		},
		&stepCreateFolder{
			Folder: p.config.Folder,
		},
		NewStepCreateSnapshot(artifact, p),
		NewStepMarkAsTemplate(artifact),
	}
	runner := common.NewRunnerWithPauseFn(steps, p.config.PackerConfig, ui, state)
	runner.Run(ctx, state)
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, false, false, rawErr.(error)
	}
	return artifact, true, true, nil
}
