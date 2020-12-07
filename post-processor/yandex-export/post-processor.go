//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config
//go:generate go run ./scripts/script-to-var.go ./scripts/export.sh CloudInitScript cloud-init-script.go

package yandexexport

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/builder/yandex"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/packerbuilderdata"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer/post-processor/artifice"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

const defaultStorageEndpoint = "storage.yandexcloud.net"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	yandex.AccessConfig `mapstructure:",squash"`
	yandex.CommonConfig `mapstructure:",squash"`
	ExchangeConfig      `mapstructure:",squash"`

	// List of paths to Yandex Object Storage where exported image will be uploaded.
	// Please be aware that use of space char inside path not supported.
	// Also this param support [build](/docs/templates/engine) template function.
	// Check available template data for [Yandex](/docs/builders/yandex#build-template-data) builder.
	// Paths to Yandex Object Storage where exported image will be uploaded.
	Paths []string `mapstructure:"paths" required:"true"`

	// Path to a PEM encoded private key file to use to authenticate with SSH.
	// The `~` can be used in path and will be expanded to the home directory
	// of current user. Login for attach: `ubuntu`
	SSHPrivateKeyFile string `mapstructure:"ssh_private_key_file" required:"false"`

	ctx interpolate.Context
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
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"paths",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	// Accumulate any errors
	var errs *packersdk.MultiError

	errs = packersdk.MultiErrorAppend(errs, p.config.AccessConfig.Prepare(&p.config.ctx)...)

	// Set defaults.
	if p.config.DiskSizeGb == 0 {
		p.config.DiskSizeGb = 100
	}

	errs = p.config.CommonConfig.Prepare(errs)
	errs = p.config.ExchangeConfig.Prepare(errs)

	if len(p.config.Paths) == 0 {
		errs = packersdk.MultiErrorAppend(
			errs, fmt.Errorf("paths must be specified"))
	}

	// Validate templates in 'paths'
	for _, path := range p.config.Paths {
		if err = interpolate.Validate(path, &p.config.ctx); err != nil {
			errs = packersdk.MultiErrorAppend(
				errs, fmt.Errorf("Error parsing one of 'paths' template: %s", err))
		}
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	// Due to the fact that now it's impossible to go to the object storage
	// through the internal network - we need access
	// to the global Internet: either through ipv4 or ipv6
	// TODO: delete this when access appears
	if p.config.UseIPv4Nat == false && p.config.UseIPv6 == false {
		p.config.UseIPv4Nat = true
	}
	p.config.Preemptible = true //? safety

	if p.config.Labels == nil {
		p.config.Labels = make(map[string]string)
	}
	if _, ok := p.config.Labels["role"]; !ok {
		p.config.Labels["role"] = "exporter"
	}
	if _, ok := p.config.Labels["target"]; !ok {
		p.config.Labels["target"] = "object-storage"
	}

	return nil
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packersdk.Ui, artifact packersdk.Artifact) (packersdk.Artifact, bool, bool, error) {
	switch artifact.BuilderId() {
	case yandex.BuilderID, artifice.BuilderId:
		break
	default:
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only export from Yandex Cloud builder artifact or Artifice post-processor artifact.",
			artifact.BuilderId())
		return nil, false, false, err
	}

	// prepare and render values
	var generatedData map[interface{}]interface{}
	stateData := artifact.State("generated_data")
	if stateData != nil {
		// Make sure it's not a nil map so we can assign to it later.
		generatedData = stateData.(map[interface{}]interface{})
	}
	// If stateData has a nil map generatedData will be nil
	// and we need to make sure it's not
	if generatedData == nil {
		generatedData = make(map[interface{}]interface{})
	}
	p.config.ctx.Data = generatedData

	var err error
	// Render this key since we didn't in the configure phase
	for i, path := range p.config.Paths {
		p.config.Paths[i], err = interpolate.Render(path, &p.config.ctx)
		if err != nil {
			return nil, false, false, fmt.Errorf("Error rendering one of 'path' template: %s", err)
		}
	}

	log.Printf("Rendered path items: %v", p.config.Paths)

	imageID := artifact.State("ImageID").(string)
	ui.Say(fmt.Sprintf("Exporting image %v to destination: %v", imageID, p.config.Paths))

	driver, err := yandex.NewDriverYC(ui, &p.config.AccessConfig)
	if err != nil {
		return nil, false, false, err
	}

	imageDesc, err := driver.SDK().Compute().Image().Get(ctx, &compute.GetImageRequest{
		ImageId: imageID,
	})
	if err != nil {
		return nil, false, false, err
	}
	secDiskSpec := []*compute.AttachedDiskSpec{
		{
			AutoDelete: true,
			DeviceName: "doexport",
			Disk: &compute.AttachedDiskSpec_DiskSpec_{
				DiskSpec: &compute.AttachedDiskSpec_DiskSpec{
					Source: &compute.AttachedDiskSpec_DiskSpec_ImageId{
						ImageId: imageID,
					},
					Size: imageDesc.MinDiskSize,
				},
			},
		},
	}
	// Set up exporter instance configuration.
	exporterName := fmt.Sprintf("%s-exporter", artifact.Id())
	yandexConfig := ycSaneDefaults(&p.config,
		map[string]string{
			"image_id":  imageID,
			"name":      exporterName,
			"paths":     strings.Join(p.config.Paths, " "),
			"user-data": CloudInitScript,
			"zone":      p.config.Zone,
		},
	)
	if yandexConfig.InstanceConfig.InstanceName == "" {
		yandexConfig.InstanceConfig.InstanceName = exporterName
	}
	if yandexConfig.DiskName == "" {
		yandexConfig.DiskName = exporterName
	}

	errs := yandexConfig.Communicator.Prepare(interpolate.NewContext())
	if len(errs) > 0 {
		err := &packersdk.MultiError{Errors: errs}
		return nil, false, false, err
	}
	ui.Say(fmt.Sprintf("Validating service_account_id: '%s'...", yandexConfig.ServiceAccountID))
	if err := validateServiceAccount(ctx, driver.SDK(), yandexConfig.ServiceAccountID); err != nil {
		return nil, false, false, err
	}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("config", &yandexConfig)
	state.Put("driver", driver)
	state.Put("sdk", driver.SDK())
	state.Put("ui", ui)
	state.Put("secondary_disk", secDiskSpec)

	// Build the steps.
	steps := []multistep.Step{
		&yandex.StepCreateSSHKey{
			Debug:        p.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("yc_export_pp_%s.pem", p.config.PackerBuildName),
		},
		&yandex.StepCreateInstance{
			Debug:         p.config.PackerDebug,
			SerialLogFile: yandexConfig.SerialLogFile,
			GeneratedData: &packerbuilderdata.GeneratedData{State: state},
		},
		new(yandex.StepWaitCloudInitScript),
		&yandex.StepTeardownInstance{
			SerialLogFile: yandexConfig.SerialLogFile,
		},
	}

	// Run the steps.
	p.runner = commonsteps.NewRunner(steps, p.config.PackerConfig, ui)
	p.runner.Run(ctx, state)
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, false, false, rawErr.(error)
	}

	result := &Artifact{
		paths: p.config.Paths,
		urls:  formUrls(p.config.Paths),
	}

	return result, false, false, nil
}

func ycSaneDefaults(c *Config, md map[string]string) yandex.Config {
	yandexConfig := yandex.Config{
		CommonConfig: c.CommonConfig,
		AccessConfig: c.AccessConfig,
		Communicator: communicator.Config{
			Type: "ssh",
			SSH: communicator.SSH{
				SSHUsername: "ubuntu",
			},
		},
	}
	if c.SSHPrivateKeyFile != "" {
		yandexConfig.Communicator.SSH.SSHPrivateKeyFile = c.SSHPrivateKeyFile
	}
	if yandexConfig.Metadata == nil {
		yandexConfig.Metadata = md
	} else {
		for k, v := range md {
			yandexConfig.Metadata[k] = v
		}
	}

	yandexConfig.SourceImageFamily = "ubuntu-1604-lts"
	yandexConfig.SourceImageFolderID = yandex.StandardImagesFolderID

	yandexConfig.ServiceAccountID = c.ServiceAccountID

	return yandexConfig
}

func formUrls(paths []string) []string {
	result := []string{}
	for _, path := range paths {
		url := fmt.Sprintf("https://%s/%s", defaultStorageEndpoint, strings.TrimPrefix(path, "s3://"))
		result = append(result, url)
	}
	return result
}

func validateServiceAccount(ctx context.Context, ycsdk *ycsdk.SDK, serviceAccountID string) error {
	_, err := ycsdk.IAM().ServiceAccount().Get(ctx, &iam.GetServiceAccountRequest{
		ServiceAccountId: serviceAccountID,
	})
	return err
}
