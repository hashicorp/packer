//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package yandexexport

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/builder/yandex"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/packerbuilderdata"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/post-processor/artifice"
	"github.com/hashicorp/packer/template/interpolate"
)

const defaultStorageEndpoint = "storage.yandexcloud.net"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	yandex.AccessConfig `mapstructure:",squash"`

	// List of paths to Yandex Object Storage where exported image will be uploaded.
	// Please be aware that use of space char inside path not supported.
	// Also this param support [build](/docs/templates/engine) template function.
	// Check available template data for [Yandex](/docs/builders/yandex#build-template-data) builder.
	// Paths to Yandex Object Storage where exported image will be uploaded.
	Paths []string `mapstructure:"paths" required:"true"`
	// The folder ID that will be used to launch a temporary instance.
	// Alternatively you may set value by environment variable `YC_FOLDER_ID`.
	FolderID string `mapstructure:"folder_id" required:"true"`
	// Service Account ID with proper permission to modify an instance, create and attach disk and
	// make upload to specific Yandex Object Storage paths.
	ServiceAccountID string `mapstructure:"service_account_id" required:"true"`
	// The size of the disk in GB. This defaults to `100`, which is 100GB.
	DiskSizeGb int `mapstructure:"disk_size" required:"false"`
	// Specify disk type for the launched instance. Defaults to `network-ssd`.
	DiskType string `mapstructure:"disk_type" required:"false"`
	// Identifier of the hardware platform configuration for the instance. This defaults to `standard-v2`.
	PlatformID string `mapstructure:"platform_id" required:"false"`
	// The Yandex VPC subnet id to use for
	// the launched instance. Note, the zone of the subnet must match the
	// zone in which the VM is launched.
	SubnetID string `mapstructure:"subnet_id" required:"false"`
	// The name of the zone to launch the instance.  This defaults to `ru-central1-a`.
	Zone string `mapstructure:"zone" required:"false"`

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
	var errs *packer.MultiError

	errs = packer.MultiErrorAppend(errs, p.config.AccessConfig.Prepare(&p.config.ctx)...)

	if len(p.config.Paths) == 0 {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("paths must be specified"))
	}

	// Validate templates in 'paths'
	for _, path := range p.config.Paths {
		if err = interpolate.Validate(path, &p.config.ctx); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error parsing one of 'paths' template: %s", err))
		}
	}

	if p.config.FolderID == "" {
		p.config.FolderID = os.Getenv("YC_FOLDER_ID")
	}

	// Set defaults.
	if p.config.DiskSizeGb == 0 {
		p.config.DiskSizeGb = 100
	}

	if p.config.DiskType == "" {
		p.config.DiskType = "network-ssd"
	}

	if p.config.PlatformID == "" {
		p.config.PlatformID = "standard-v2"
	}

	if p.config.Zone == "" {
		p.config.Zone = "ru-central1-a"
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, bool, error) {
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

	// Set up exporter instance configuration.
	exporterName := fmt.Sprintf("%s-exporter", artifact.Id())
	exporterMetadata := map[string]string{
		"image_id":  imageID,
		"name":      exporterName,
		"paths":     strings.Join(p.config.Paths, " "),
		"user-data": CloudInitScript,
		"zone":      p.config.Zone,
	}

	yandexConfig := ycSaneDefaults()
	yandexConfig.DiskName = exporterName
	yandexConfig.InstanceName = exporterName
	yandexConfig.DiskSizeGb = p.config.DiskSizeGb
	yandexConfig.Metadata = exporterMetadata
	yandexConfig.SubnetID = p.config.SubnetID
	yandexConfig.FolderID = p.config.FolderID
	yandexConfig.Zone = p.config.Zone

	if p.config.ServiceAccountID != "" {
		yandexConfig.ServiceAccountID = p.config.ServiceAccountID
	}

	if p.config.PlatformID != "" {
		yandexConfig.PlatformID = p.config.PlatformID
	}

	driver, err := yandex.NewDriverYC(ui, &p.config.AccessConfig)
	if err != nil {
		return nil, false, false, err
	}

	// Set up the state.
	state := new(multistep.BasicStateBag)
	state.Put("config", &yandexConfig)
	state.Put("driver", driver)
	state.Put("sdk", driver.SDK())
	state.Put("ui", ui)

	// Build the steps.
	steps := []multistep.Step{
		&yandex.StepCreateSSHKey{
			Debug:        p.config.PackerDebug,
			DebugKeyPath: fmt.Sprintf("yc_export_pp_%s.pem", p.config.PackerBuildName),
		},
		&yandex.StepCreateInstance{
			Debug:         p.config.PackerDebug,
			GeneratedData: &packerbuilderdata.GeneratedData{State: state},
		},
		new(yandex.StepWaitCloudInitScript),
		new(yandex.StepTeardownInstance),
	}

	// Run the steps.
	p.runner = common.NewRunner(steps, p.config.PackerConfig, ui)
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

func ycSaneDefaults() yandex.Config {
	return yandex.Config{
		DiskType:       "network-ssd",
		InstanceCores:  2,
		InstanceMemory: 2,
		Labels: map[string]string{
			"role":   "exporter",
			"target": "object-storage",
		},
		PlatformID:          "standard-v2",
		Preemptible:         true,
		SourceImageFamily:   "ubuntu-1604-lts",
		SourceImageFolderID: yandex.StandardImagesFolderID,
		UseIPv4Nat:          true,
		Zone:                "ru-central1-a",
		StateTimeout:        3 * time.Minute,
	}
}

func formUrls(paths []string) []string {
	result := []string{}
	for _, path := range paths {
		url := fmt.Sprintf("https://%s/%s", defaultStorageEndpoint, strings.TrimPrefix(path, "s3://"))
		result = append(result, url)
	}
	return result
}
