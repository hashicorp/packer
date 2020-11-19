//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package googlecomputeexport

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/builder/googlecompute"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer/post-processor/artifice"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	//The JSON file containing your account credentials.
	//If specified, the account file will take precedence over any `googlecompute` builder authentication method.
	AccountFile string `mapstructure:"account_file"`
	// This allows service account impersonation as per the [docs](https://cloud.google.com/iam/docs/impersonating-service-accounts).
	ImpersonateServiceAccount string `mapstructure:"impersonate_service_account" required:"false"`
	//The size of the export instances disk.
	//The disk is unused for the export but a larger size will increase `pd-ssd` read speed.
	//This defaults to `200`, which is 200GB.
	DiskSizeGb int64 `mapstructure:"disk_size"`
	//Type of disk used to back the export instance, like
	//`pd-ssd` or `pd-standard`. Defaults to `pd-ssd`.
	DiskType string `mapstructure:"disk_type"`
	//The export instance machine type. Defaults to `"n1-highcpu-4"`.
	MachineType string `mapstructure:"machine_type"`
	//The Google Compute network id or URL to use for the export instance.
	//Defaults to `"default"`. If the value is not a URL, it
	//will be interpolated to `projects/((builder_project_id))/global/networks/((network))`.
	//This value is not required if a `subnet` is specified.
	Network string `mapstructure:"network"`
	//A list of GCS paths where the image will be exported.
	//For example `'gs://mybucket/path/to/file.tar.gz'`
	Paths []string `mapstructure:"paths" required:"true"`
	//The Google Compute subnetwork id or URL to use for
	//the export instance. Only required if the `network` has been created with
	//custom subnetting. Note, the region of the subnetwork must match the
	//`zone` in which the VM is launched. If the value is not a URL,
	//it will be interpolated to
	//`projects/((builder_project_id))/regions/((region))/subnetworks/((subnetwork))`
	Subnetwork string `mapstructure:"subnetwork"`
	//The zone in which to launch the export instance. Defaults
	//to `googlecompute` builder zone. Example: `"us-central1-a"`
	Zone                string `mapstructure:"zone"`
	IAP                 bool   `mapstructure-to-hcl2:",skip"`
	VaultGCPOauthEngine string `mapstructure:"vault_gcp_oauth_engine"`
	ServiceAccountEmail string `mapstructure:"service_account_email"`

	account *googlecompute.ServiceAccount
	ctx     interpolate.Context
}

type PostProcessor struct {
	config Config
	runner multistep.Runner
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
	}, raws...)
	if err != nil {
		return err
	}

	errs := new(packersdk.MultiError)

	if len(p.config.Paths) == 0 {
		errs = packersdk.MultiErrorAppend(
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
		errs = packersdk.MultiErrorAppend(
			errs, fmt.Errorf("May set either account_file or "+
				"vault_gcp_oauth_engine, but not both."))
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packersdk.Ui, artifact packer.Artifact) (packer.Artifact, bool, bool, error) {
	switch artifact.BuilderId() {
	case googlecompute.BuilderId, artifice.BuilderId:
		break
	default:
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only export from Google Compute Engine builder and Artifice post-processor artifacts.",
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
		SourceImageProjectId: []string{"compute-image-tools"},
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
	cfg := googlecompute.GCEDriverConfig{
		Ui:                            ui,
		ProjectId:                     builderProjectId,
		Account:                       p.config.account,
		ImpersonateServiceAccountName: p.config.ImpersonateServiceAccount,
		VaultOauthEngineName:          p.config.VaultGCPOauthEngine,
	}

	driver, err := googlecompute.NewDriverGCE(cfg)
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
		&communicator.StepSSHKeyGen{
			CommConf: &exporterConfig.Comm,
		},
		multistep.If(p.config.PackerDebug,
			&communicator.StepDumpSSHKey{
				Path: fmt.Sprintf("gce_%s.pem", p.config.PackerBuildName),
			},
		),
		&googlecompute.StepCreateInstance{
			Debug: p.config.PackerDebug,
		},
		new(googlecompute.StepWaitStartupScript),
		new(googlecompute.StepTeardownInstance),
	}

	// Run the steps.
	p.runner = commonsteps.NewRunner(steps, p.config.PackerConfig, ui)
	p.runner.Run(ctx, state)

	result := &Artifact{paths: p.config.Paths}

	return result, false, false, nil
}
