package vsphere

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"os/exec"
	"strings"
)

var builtins = map[string]string{
	"mitchellh.vmware": "vmware",
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Insecure           bool   `mapstructure:"insecure"`
	Cluster            string `mapstructure:"cluster"`
	Datacenter         string `mapstructure:"datacenter"`
	Datastore          string `mapstructure:"datastore"`
	Debug              bool   `mapstructure:"debug"`
	Host               string `mapstructure:"host"`
	Password           string `mapstructure:"password"`
	ResourcePool       string `mapstructure:"resource_pool"`
	Username           string `mapstructure:"username"`
	VMFolder           string `mapstructure:"vm_folder"`
	VMName             string `mapstructure:"vm_name"`
	VMNetwork          string `mapstructure:"vm_network"`

	tpl *packer.ConfigTemplate
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	_, err := common.DecodeConfig(&p.config, raws...)
	if err != nil {
		return err
	}

	p.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	p.config.tpl.UserVars = p.config.PackerUserVars

	// Accumulate any errors
	errs := new(packer.MultiError)

	if _, err := exec.LookPath("ovftool"); err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("ovftool not found: %s", err))
	}

	validates := map[string]*string{
		"cluster":      &p.config.Cluster,
		"datacenter":   &p.config.Datacenter,
		"datastore":    &p.config.Datastore,
		"host":         &p.config.Host,
		"vm_network":   &p.config.VMNetwork,
		"password":     &p.config.Password,
		"resource_pool": &p.config.ResourcePool,
		"username":     &p.config.Username,
		"vm_folder":    &p.config.VMFolder,
		"vm_name":      &p.config.VMName,
	}

	for n := range validates {
		if *validates[n] == "" {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set", n))
		}
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

	vmx := ""
	for _, path := range artifact.Files() {
		if strings.HasSuffix(path, ".vmx") {
			vmx = path
			break
		}
	}

	if vmx == "" {
		return nil, false, fmt.Errorf("VMX file not found")
	}

	vm_name, err := p.config.tpl.Process(p.config.VMName, p.config.PackerUserVars)
	if err != nil {
		return nil, false, fmt.Errorf("Failed: %s", err)
	}

	username, err := p.config.tpl.Process(p.config.Username, p.config.PackerUserVars)
	if err != nil {
		return nil, false, fmt.Errorf("Failed: %s", err)
	}

	password, err := p.config.tpl.Process(p.config.Password, p.config.PackerUserVars)
	if err != nil {
		return nil, false, fmt.Errorf("Failed: %s", err)
	}

	ui.Message(fmt.Sprintf("Uploading %s to vSphere", vmx))

	args := []string{
		fmt.Sprintf("--noSSLVerify=%t", p.config.Insecure),
		"--acceptAllEulas",
		fmt.Sprintf("--name=%s", vm_name),
		fmt.Sprintf("--datastore=%s", p.config.Datastore),
		fmt.Sprintf("--network=%s", p.config.VMNetwork),
		fmt.Sprintf("--vmFolder=%s", p.config.VMFolder),
		fmt.Sprintf("%s", vmx),
		fmt.Sprintf("vi://%s:%s@%s/%s/host/%s/Resources/%s",
			username,
			password,
			p.config.Host,
			p.config.Datacenter,
			p.config.Cluster,
			p.config.ResourcePool),
	}

	if p.config.Debug {
		ui.Message(fmt.Sprintf("DEBUG: %s", args))
	}

	var out bytes.Buffer
	cmd := exec.Command("ovftool", args...)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, false, fmt.Errorf("Failed: %s\nStdout: %s", err, out.String())
	}

	ui.Message(fmt.Sprintf("%s", out.String()))

	return artifact, false, nil
}
