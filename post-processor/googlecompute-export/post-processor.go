//go:generate mapstructure-to-hcl2 -type Config

package googlecomputeexport

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/builder/googlecompute"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"golang.org/x/oauth2/jwt"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	AccountFile string `mapstructure:"account_file"`

	DiskSizeGb          int64    `mapstructure:"disk_size"`
	DiskType            string   `mapstructure:"disk_type"`
	MachineType         string   `mapstructure:"machine_type"`
	Network             string   `mapstructure:"network"`
	Paths               []string `mapstructure:"paths"`
	Subnetwork          string   `mapstructure:"subnetwork"`
	VaultGCPOauthEngine string   `mapstructure:"vault_gcp_oauth_engine"`
	Zone                string   `mapstructure:"zone"`
	ServiceAccountEmail string   `mapstructure:"service_account_email"`

	account *jwt.Config
	ctx     interpolate.Context
}

type PostProcessor struct {
	config Config
	runner multistep.Runner
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
	}, raws...)
	if err != nil {
		return err
	}

	errs := new(packer.MultiError)

	if len(p.config.Paths) == 0 {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("paths must be specified"))
	}

	// Set defaults.
	if p.config.DiskSizeGb == 0 {
		p.config.DiskSizeGb = 200
	}

	if p.config.DiskType == "" {
		p.config.DiskType = "pd-ssd"
	}

	if p.config.MachineType == "" {
		p.config.MachineType = "n1-highcpu-4"
	}

	if p.config.Network == "" && p.config.Subnetwork == "" {
		p.config.Network = "default"
	}

	if p.config.AccountFile != "" && p.config.VaultGCPOauthEngine != "" {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("May set either account_file or "+
				"vault_gcp_oauth_engine, but not both."))
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, bool, error) {
	if artifact.BuilderId() != googlecompute.BuilderId {
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only export from Google Compute Engine builder artifacts.",
			artifact.BuilderId())
		return nil, false, false, err
	}

	builderAccountFile := artifact.State("AccountFilePath").(string)
	builderImageName := artifact.State("ImageName").(string)
	builderProjectId := artifact.State("ProjectId").(string)
	builderZone := artifact.State("BuildZone").(string)

	ui.Say(fmt.Sprintf("Exporting image %v to destination: %v", builderImageName, p.config.Paths))

	if p.config.Zone == "" {
		p.config.Zone = builderZone
	}

	// Set up credentials for GCE driver.
	if builderAccountFile != "" {
		cfg, err := googlecompute.ProcessAccountFile(builderAccountFile)
		if err != nil {
			return nil, false, false, err
		}
		p.config.account = cfg
	}
	if p.config.AccountFile != "" {
		cfg, err := googlecompute.ProcessAccountFile(p.config.AccountFile)
		if err != nil {
			return nil, false, false, err
		}
		p.config.account = cfg
	}

	// Set up exporter instance configuration.
	exporterName := fmt.Sprintf("%s-exporter", artifact.Id())
	exporterMetadata := map[string]string{
		"image_name":     builderImageName,
		"name":           exporterName,
		"paths":          strings.Join(p.config.Paths, " "),
		"startup-script": StartupScript,
		"zone":           p.config.Zone,
	}
	exporterConfig := googlecompute.Config{
		DiskName:             exporterName,
		DiskSizeGb:           p.config.DiskSizeGb,
		DiskType:             p.config.DiskType,
		InstanceName:         exporterName,
		MachineType:          p.config.MachineType,
		Metadata:             exporterMetadata,
		Network:              p.config.Network,
		NetworkProjectId:     builderProjectId,
		StateTimeout:         5 * time.Minute,
		SourceImageFamily:    "debian-9-worker",
		SourceImageProjectId: "compute-image-tools",
		Subnetwork:           p.config.Subnetwork,
		Zone:                 p.config.Zone,
		Scopes: []string{
			"https://www.googleapis.com/auth/compute",
			"https://www.googleapis.com/auth/devstorage.full_control",
			"https://www.googleapis.com/auth/userinfo.email",
		},
	}
	if p.config.ServiceAccountEmail != "" {
		exporterConfig.ServiceAccountEmail = p.config.ServiceAccountEmail
	}

	driver, err := googlecompute.NewDriverGCE(ui, builderProjectId,
		p.config.account, p.config.VaultGCPOauthEngine)
	if err != nil {
		return nil, false, false, err
	}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("config", &exporterConfig)
	state.Put("driver", driver)
	state.Put("ui", ui)

	// Build the steps.
	steps := []multistep.Step{
		&googlecompute.StepCreateSSHKey{
			Debug:        p.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("gce_%s.pem", p.config.PackerBuildName),
		},
		&googlecompute.StepCreateInstance{
			Debug: p.config.PackerDebug,
		},
		new(googlecompute.StepWaitStartupScript),
		new(googlecompute.StepTeardownInstance),
	}

	// Run the steps.
	p.runner = common.NewRunner(steps, p.config.PackerConfig, ui)
	p.runner.Run(ctx, state)

	result := &Artifact{paths: p.config.Paths}

	return result, false, false, nil
}
