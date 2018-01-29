package googlecomputeexport

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/hashicorp/packer/builder/googlecompute"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Paths             []string `mapstructure:"paths"`
	KeepOriginalImage bool     `mapstructure:"keep_input_artifact"`

	ctx interpolate.Context
}

type PostProcessor struct {
	config Config
	runner multistep.Runner
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
	}, raws...)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	ui.Say("Starting googlecompute-export...")
	ui.Say(fmt.Sprintf("Exporting image to destinations: %v", p.config.Paths))
	if artifact.BuilderId() != googlecompute.BuilderId {
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only export from Google Compute Engine builder artifacts.",
			artifact.BuilderId())
		return nil, p.config.KeepOriginalImage, err
	}

	result := &Artifact{paths: p.config.Paths}

	if len(p.config.Paths) > 0 {
		accountKeyFilePath := artifact.State("AccountFilePath").(string)
		imageName := artifact.State("ImageName").(string)
		imageSizeGb := artifact.State("ImageSizeGb").(int64)
		projectId := artifact.State("ProjectId").(string)
		zone := artifact.State("BuildZone").(string)

		// Set up instance configuration.
		instanceName := fmt.Sprintf("%s-exporter", artifact.Id())
		metadata := map[string]string{
			"image_name":     imageName,
			"name":           instanceName,
			"paths":          strings.Join(p.config.Paths, " "),
			"startup-script": StartupScript,
			"zone":           zone,
		}
		exporterConfig := googlecompute.Config{
			InstanceName:         instanceName,
			SourceImageProjectId: "debian-cloud",
			SourceImage:          "debian-8-jessie-v20160629",
			DiskName:             instanceName,
			DiskSizeGb:           imageSizeGb + 10,
			DiskType:             "pd-standard",
			Metadata:             metadata,
			MachineType:          "n1-standard-4",
			Zone:                 zone,
			Network:              "default",
			RawStateTimeout:      "5m",
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/compute",
				"https://www.googleapis.com/auth/devstorage.full_control",
			},
		}
		exporterConfig.CalcTimeout()

		// Set up credentials and GCE driver.
		b, err := ioutil.ReadFile(accountKeyFilePath)
		if err != nil {
			err = fmt.Errorf("Error fetching account credentials: %s", err)
			return nil, p.config.KeepOriginalImage, err
		}
		accountKeyContents := string(b)
		googlecompute.ProcessAccountFile(&exporterConfig.Account, accountKeyContents)
		driver, err := googlecompute.NewDriverGCE(ui, projectId, &exporterConfig.Account)
		if err != nil {
			return nil, p.config.KeepOriginalImage, err
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
		p.runner.Run(state)
	}

	return result, p.config.KeepOriginalImage, nil
}
