package proxmox

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// The unique id for the builder
const BuilderId = "proxmox.builder"

type Builder struct {
	config        Config
	runner        multistep.Runner
	proxmoxClient *proxmox.Client
}

// Builder implements packer.Builder
var _ packer.Builder = &Builder{}

var pluginVersion = "1.0.0"

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	config, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return warnings, errs
	}
	b.config = *config
	return nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	var err error
	tlsConfig := &tls.Config{
		InsecureSkipVerify: b.config.SkipCertValidation,
	}
	b.proxmoxClient, err = proxmox.NewClient(b.config.ProxmoxURL.String(), nil, tlsConfig)
	if err != nil {
		return nil, err
	}

	err = b.proxmoxClient.Login(b.config.Username, b.config.Password)
	if err != nil {
		return nil, err
	}

	// Set up the state
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("proxmoxClient", b.proxmoxClient)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		&stepStartVM{},
		&common.StepHTTPServer{
			HTTPDir:     b.config.HTTPDir,
			HTTPPortMin: b.config.HTTPPortMin,
			HTTPPortMax: b.config.HTTPPortMax,
		},
		&stepTypeBootCommand{
			BootConfig: b.config.BootConfig,
			Ctx:        b.config.ctx,
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
	// Run the steps
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)
	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	artifact := &Artifact{
		templateID:    state.Get("template_id").(int),
		proxmoxClient: b.proxmoxClient,
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
