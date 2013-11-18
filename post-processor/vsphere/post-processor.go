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

	Insecure bool `mapstructure:"insecure"`

	Datacenter        string `mapstructure:"datacenter"`
	Datastore         string `mapstructure:"datastore"`
	Host              string `mapstructure:"host"`
	VMNetwork         string `mapstructure:"vm_network"`
	Password          string `mapstructure:"password"`
	PathToResoucePool string `mapstructure:"path_to_resouce_pool"`
	Username          string `mapstructure:"username"`
	VMFolder          string `mapstructure:"vm_folder"`
	VMName            string `mapstructure:"vm_name"`
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

	program := "ovftool"
	_, erro := exec.LookPath(program)
	if erro != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Error : %s not set", erro))
	}

	validates := map[string]*string{
		"datacenter":           &p.config.Datacenter,
		"datastore":            &p.config.Datastore,
		"host":                 &p.config.Host,
		"vm_network":           &p.config.VMNetwork,
		"password":             &p.config.Password,
		"path_to_resouce_pool": &p.config.PathToResoucePool,
		"username":             &p.config.Username,
		"vm_folder":            &p.config.VMFolder,
		"vm_name":              &p.config.VMName,
	}

	for n := range validates {
		if *validates[n] == "" {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Argument %s: not set", n))
		}
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	_, ok := builtins[artifact.BuilderId()]
	if !ok {
		return nil, false, fmt.Errorf("Unknown artifact type, can't build box: %s", artifact.BuilderId())
	}

	vmx := ""
	for _, path := range artifact.Files() {
		if strings.HasSuffix(path, ".vmx") {
			vmx = path
		}
	}

	if vmx == "" {
		return nil, false, fmt.Errorf("VMX file not found")
	}

	ui.Message(fmt.Sprintf("uploading %s to vSphere", vmx))

	program := "ovftool"
	nossl := fmt.Sprintf("--noSSLVerify=%t", p.config.Insecure)
	accepteulas := "--acceptAllEulas"
	name := "--name=" + p.config.VMName
	datastore := "--datastore=" + p.config.Datastore
	network := "--network=" + p.config.VMNetwork
	vm_folder := "--vmFolder=" + p.config.VMFolder
	url := "vi://" + p.config.Username + ":" + p.config.Password + "@" + p.config.Host + "/" + p.config.Datacenter + "/" + p.config.PathToResoucePool

	cmd := exec.Command(program, nossl, accepteulas, name, datastore, network, vm_folder, vmx, url)

	var out bytes.Buffer
	cmd.Stdout = &out
	err_run := cmd.Run()
	if err_run != nil {
		return nil, false, fmt.Errorf("%s", out.String())
	}

	ui.Message(fmt.Sprintf("%s", out.String()))

	return artifact, false, nil
}
