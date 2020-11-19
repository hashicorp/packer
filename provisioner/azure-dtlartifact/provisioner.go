//go:generate mapstructure-to-hcl2 -type Config,DtlArtifact,ArtifactParameter

package devtestlabsartifacts

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/devtestlabs/mgmt/2018-09-15/dtl"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/builder/azure/common/client"
	dtlBuilder "github.com/hashicorp/packer/builder/azure/dtl"

	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"

	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
)

type DtlArtifact struct {
	ArtifactName string              `mapstructure:"artifact_name"`
	ArtifactId   string              `mapstructure:"artifact_id"`
	Parameters   []ArtifactParameter `mapstructure:"parameters"`
}

type ArtifactParameter struct {
	Name  string `mapstructure:"name"`
	Value string `mapstructure:"value"`
	Type  string `mapstructure:"type"`
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Authentication via OAUTH
	ClientConfig client.Config `mapstructure:",squash"`

	DtlArtifacts []DtlArtifact `mapstructure:"dtl_artifacts"`
	LabName      string        `mapstructure:"lab_name"`

	ResourceGroupName string `mapstructure:"lab_resource_group_name"`

	VMName string `mapstructure:"vm_name"`

	// The default PollingDuration for azure is 15mins, this property will override
	// that value. See [Azure DefaultPollingDuration](https://godoc.org/github.com/Azure/go-autorest/autorest#pkg-constants)
	// If your Packer build is failing on the
	// ARM deployment step with the error `Original Error:
	// context deadline exceeded`, then you probably need to increase this timeout from
	// its default of "15m" (valid time units include `s` for seconds, `m` for
	// minutes, and `h` for hours.)
	PollingDurationTimeout time.Duration `mapstructure:"polling_duration_timeout" required:"false"`

	AzureTags map[string]*string `mapstructure:"azure_tags"`

	Json map[string]interface{}

	ctx interpolate.Context
}

type Provisioner struct {
	config       Config
	communicator packer.Communicator
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *Provisioner) Prepare(raws ...interface{}) error {
	// // Create passthrough for winrm password so we can fill it in once we know
	// // it
	// p.config.ctx.Data = &EnvVarsTemplate{
	// 	WinRMPassword: `{{.WinRMPassword}}`,
	// }
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         "azure-dtlartifact",
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"execute_command",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	p.config.ClientConfig.CloudEnvironmentName = "Public"

	return nil
}

func (p *Provisioner) Communicator() packer.Communicator {
	return p.communicator
}

func (p *Provisioner) Provision(ctx context.Context, ui packersdk.Ui, comm packer.Communicator, _ map[string]interface{}) error {

	p.communicator = comm

	err := p.config.ClientConfig.SetDefaultValues()
	if err != nil {
		ui.Say(fmt.Sprintf("Error saving debug key: %s", err))
		return nil
	}

	/////////////////////////////////////////////
	// Polling Duration Timeout
	if p.config.PollingDurationTimeout == 0 {
		// In the sdk, the default is 15 m.
		p.config.PollingDurationTimeout = 15 * time.Minute
	}
	// FillParameters function captures authType and sets defaults.
	err = p.config.ClientConfig.FillParameters()
	if err != nil {
		return err
	}

	spnCloud, err := p.config.ClientConfig.GetServicePrincipalToken(ui.Say, p.config.ClientConfig.CloudEnvironment().ResourceManagerEndpoint)
	if err != nil {
		return err
	}

	ui.Message("Creating Azure Resource Manager (ARM) client ...")
	azureClient, err := dtlBuilder.NewAzureClient(
		p.config.ClientConfig.SubscriptionID,
		"",
		p.config.ClientConfig.CloudEnvironment(),
		p.config.PollingDurationTimeout,
		p.config.PollingDurationTimeout,
		spnCloud)

	if err != nil {
		ui.Say(fmt.Sprintf("Error saving debug key: %s", err))
		return err
	}

	ui.Say("Installing Artifact DTL")
	dtlArtifacts := []dtl.ArtifactInstallProperties{}

	if p.config.DtlArtifacts != nil {
		for i := range p.config.DtlArtifacts {
			p.config.DtlArtifacts[i].ArtifactId = fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.DevTestLab/labs/%s/artifactSources/public repo/artifacts/%s",
				p.config.ClientConfig.SubscriptionID,
				p.config.ResourceGroupName,
				p.config.LabName,
				p.config.DtlArtifacts[i].ArtifactName)

			dparams := []dtl.ArtifactParameterProperties{}
			for j := range p.config.DtlArtifacts[i].Parameters {
				dp := &dtl.ArtifactParameterProperties{}
				dp.Name = &p.config.DtlArtifacts[i].Parameters[j].Name
				dp.Value = &p.config.DtlArtifacts[i].Parameters[j].Value

				dparams = append(dparams, *dp)
			}
			Aip := dtl.ArtifactInstallProperties{
				ArtifactID:    &p.config.DtlArtifacts[i].ArtifactId,
				Parameters:    &dparams,
				ArtifactTitle: &p.config.DtlArtifacts[i].ArtifactName,
			}
			dtlArtifacts = append(dtlArtifacts, Aip)
		}
	}

	dtlApplyArifactRequest := dtl.ApplyArtifactsRequest{
		Artifacts: &dtlArtifacts,
	}

	ui.Say("Applying artifact ")
	f, err := azureClient.DtlVirtualMachineClient.ApplyArtifacts(ctx, p.config.ResourceGroupName, p.config.LabName, p.config.VMName, dtlApplyArifactRequest)

	if err == nil {
		err = f.WaitForCompletionRef(ctx, azureClient.DtlVirtualMachineClient.Client)
	}
	if err != nil {
		ui.Say(fmt.Sprintf("Error Applying artifact: %s", err))
	}
	ui.Say("Aftifact installed")
	return err
}
