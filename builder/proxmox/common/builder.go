package proxmox

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

func NewSharedBuilder(id string, config Config, preSteps []multistep.Step, postSteps []multistep.Step) *Builder {
	return &Builder{
		id:        id,
		config:    config,
		preSteps:  preSteps,
		postSteps: postSteps,
	}
}

type Builder struct {
	id            string
	config        Config
	preSteps      []multistep.Step
	postSteps     []multistep.Step
	runner        multistep.Runner
	proxmoxClient *proxmox.Client
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook, state multistep.StateBag) (packer.Artifact, error) {
	var err error
	tlsConfig := &tls.Config{
		InsecureSkipVerify: b.config.SkipCertValidation,
	}
	b.proxmoxClient, err = proxmox.NewClient(b.config.proxmoxURL.String(), nil, tlsConfig)
	if err != nil {
		return nil, err
	}

	err = b.proxmoxClient.Login(b.config.Username, b.config.Password, "")
	if err != nil {
		return nil, err
	}

	// Set up the state
	state.Put("config", &b.config)
	state.Put("proxmoxClient", b.proxmoxClient)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	coreSteps := []multistep.Step{
		&stepStartVM{},
		&common.StepHTTPServer{
			HTTPDir:     b.config.HTTPDir,
			HTTPPortMin: b.config.HTTPPortMin,
			HTTPPortMax: b.config.HTTPPortMax,
			HTTPAddress: b.config.HTTPAddress,
		},
		&stepTypeBootCommand{
			BootConfig: b.config.BootConfig,
			Ctx:        b.config.Ctx,
		},
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      commHost(b.config.Comm.Host()),
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
		&common.StepProvision{},
		&common.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
		&stepConvertToTemplate{},
		&stepFinalizeTemplateConfig{},
		&stepSuccess{},
	}
	steps := append(b.preSteps, coreSteps...)
	steps = append(steps, b.postSteps...)
	// Run the steps
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)
	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}
	// If we were interrupted or cancelled, then just exit.
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("build was cancelled")
	}

	// Verify that the template_id was set properly, otherwise we didn't progress through the last step
	tplID, ok := state.Get("template_id").(int)
	if !ok {
		return nil, fmt.Errorf("template ID could not be determined")
	}

	artifact := &Artifact{
		builderID:     b.id,
		templateID:    tplID,
		proxmoxClient: b.proxmoxClient,
		StateData:     map[string]interface{}{"generated_data": state.Get("generated_data")},
	}

	return artifact, nil
}

// Returns ssh_host or winrm_host (see communicator.Config.Host) config
// parameter when set, otherwise gets the host IP from running VM
func commHost(host string) func(state multistep.StateBag) (string, error) {
	if host != "" {
		return func(state multistep.StateBag) (string, error) {
			return host, nil
		}
	}
	return getVMIP
}

// Reads the first non-loopback interface's IP address from the VM.
// qemu-guest-agent package must be installed on the VM
func getVMIP(state multistep.StateBag) (string, error) {
	c := state.Get("proxmoxClient").(*proxmox.Client)
	vmRef := state.Get("vmRef").(*proxmox.VmRef)

	ifs, err := c.GetVmAgentNetworkInterfaces(vmRef)
	if err != nil {
		return "", err
	}

	// TODO: Do something smarter here? Allow specifying interface? Or address family?
	// For now, just go for first non-loopback
	for _, iface := range ifs {
		for _, addr := range iface.IPAddresses {
			if addr.IsLoopback() {
				continue
			}
			return addr.String(), nil
		}
	}

	return "", fmt.Errorf("Found no IP addresses on VM")
}
