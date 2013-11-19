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
	Datacenter         string `mapstructure:"datacenter"`
	Datastore          string `mapstructure:"datastore"`
	Host               string `mapstructure:"host"`
	Password           string `mapstructure:"password"`
	PathToResourcePool string `mapstructure:"path_to_resource_pool"`
	Username           string `mapstructure:"username"`
	VMFolder           string `mapstructure:"vm_folder"`
	VMName             string `mapstructure:"vm_name"`
	VMNetwork          string `mapstructure:"vm_network"`
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	_, err := common.DecodeConfig(&p.config, raws...)
	if err != nil {
		return err
	}

	tpl, err := packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	tpl.UserVars = p.config.PackerUserVars

	// Accumulate any errors
	errs := new(packer.MultiError)

	if _, err := exec.LookPath("ovftool"); err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("ovftool not found: %s", err))
	}

	validates := map[string]*string{
		"datacenter":            &p.config.Datacenter,
		"datastore":             &p.config.Datastore,
		"host":                  &p.config.Host,
		"vm_network":            &p.config.VMNetwork,
		"password":              &p.config.Password,
		"path_to_resource_pool": &p.config.PathToResourcePool,
		"username":              &p.config.Username,
		"vm_folder":             &p.config.VMFolder,
		"vm_name":               &p.config.VMName,
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

	ui.Message(fmt.Sprintf("Uploading %s to vSphere", vmx))

	args := []string{
		fmt.Sprintf("--noSSLVerify=%t", p.config.Insecure),
		"--acceptAllEulas",
		fmt.Sprintf("--name=%s", p.config.VMName),
		fmt.Sprintf("--datastore=%s", p.config.Datastore),
		fmt.Sprintf("--network=%s", p.config.VMNetwork),
		fmt.Sprintf("--vmFolder=%s", p.config.VMFolder),
		fmt.Sprintf("vi://%s:%s@%s/%s/%s",
			p.config.Username,
			p.config.Password,
			p.config.Host,
			p.config.Datacenter,
			p.config.PathToResourcePool),
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
