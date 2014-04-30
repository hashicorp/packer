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

	Insecure     bool   `mapstructure:"insecure"`
	Cluster      string `mapstructure:"cluster"`
	Datacenter   string `mapstructure:"datacenter"`
	Datastore    string `mapstructure:"datastore"`
	DiskMode     string `mapstructure:"disk_mode"`
	Host         string `mapstructure:"host"`
	Password     string `mapstructure:"password"`
	ResourcePool string `mapstructure:"resource_pool"`
	Username     string `mapstructure:"username"`
	VMFolder     string `mapstructure:"vm_folder"`
	VMName       string `mapstructure:"vm_name"`
	VMNetwork    string `mapstructure:"vm_network"`

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

	// Defaults
	if p.config.DiskMode == "" {
		p.config.DiskMode = "thick"
	}

	// Accumulate any errors
	errs := new(packer.MultiError)

	if _, err := exec.LookPath("ovftool"); err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("ovftool not found: %s", err))
	}

	// First define all our templatable parameters that are _required_
	templates := map[string]*string{
		"cluster":       &p.config.Cluster,
		"datacenter":    &p.config.Datacenter,
		"diskmode":      &p.config.DiskMode,
		"host":          &p.config.Host,
		"password":      &p.config.Password,
		"resource_pool": &p.config.ResourcePool,
		"username":      &p.config.Username,
		"vm_name":       &p.config.VMName,
	}
	for key, ptr := range templates {
		if *ptr == "" {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set", key))
		}
	}

	// Then define the ones that are optional
	templates["datastore"] = &p.config.Datastore
	templates["vm_network"] = &p.config.VMNetwork
	templates["vm_folder"] = &p.config.VMFolder

	// Template process
	for key, ptr := range templates {
		*ptr, err = p.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", key, err))
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

	args := []string{
		fmt.Sprintf("--noSSLVerify=%t", p.config.Insecure),
		"--acceptAllEulas",
		fmt.Sprintf("--name=%s", p.config.VMName),
		fmt.Sprintf("--datastore=%s", p.config.Datastore),
		fmt.Sprintf("--diskMode=%s", p.config.DiskMode),
		fmt.Sprintf("--network=%s", p.config.VMNetwork),
		fmt.Sprintf("--vmFolder=%s", p.config.VMFolder),
		fmt.Sprintf("%s", vmx),
		fmt.Sprintf("vi://%s:%s@%s/%s/host/%s/Resources/%s",
			p.config.Username,
			p.config.Password,
			p.config.Host,
			p.config.Datacenter,
			p.config.Cluster,
			p.config.ResourcePool),
	}

	ui.Message(fmt.Sprintf("Uploading %s to vSphere", vmx))
	var out bytes.Buffer
	cmd := exec.Command("ovftool", args...)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, false, fmt.Errorf("Failed: %s\nStdout: %s", err, out.String())
	}

	ui.Message(fmt.Sprintf("%s", out.String()))

	return artifact, false, nil
}
