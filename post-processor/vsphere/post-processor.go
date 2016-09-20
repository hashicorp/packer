package vsphere

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

var builtins = map[string]string{
	"mitchellh.vmware":     "vmware",
	"mitchellh.vmware-esx": "vmware",
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Cluster      string   `mapstructure:"cluster"`
	Datacenter   string   `mapstructure:"datacenter"`
	Datastore    string   `mapstructure:"datastore"`
	DiskMode     string   `mapstructure:"disk_mode"`
	Host         string   `mapstructure:"host"`
	Insecure     bool     `mapstructure:"insecure"`
	Options      []string `mapstructure:"options"`
	Overwrite    bool     `mapstructure:"overwrite"`
	Password     string   `mapstructure:"password"`
	ResourcePool string   `mapstructure:"resource_pool"`
	Username     string   `mapstructure:"username"`
	VMFolder     string   `mapstructure:"vm_folder"`
	VMName       string   `mapstructure:"vm_name"`
	VMNetwork    string   `mapstructure:"vm_network"`

	ctx interpolate.Context
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

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
		"cluster":    &p.config.Cluster,
		"datacenter": &p.config.Datacenter,
		"diskmode":   &p.config.DiskMode,
		"host":       &p.config.Host,
		"password":   &p.config.Password,
		"username":   &p.config.Username,
		"vm_name":    &p.config.VMName,
	}
	for key, ptr := range templates {
		if *ptr == "" {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set", key))
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

	source := ""
	for _, path := range artifact.Files() {
		if strings.HasSuffix(path, ".vmx") || strings.HasSuffix(path, ".ovf") || strings.HasSuffix(path, ".ova") {
			source = path
			break
		}
	}

	if source == "" {
		return nil, false, fmt.Errorf("VMX, OVF or OVA file not found")
	}

	ovftool_uri := fmt.Sprintf("vi://%s:%s@%s/%s/host/%s",
		url.QueryEscape(p.config.Username),
		url.QueryEscape(p.config.Password),
		p.config.Host,
		p.config.Datacenter,
		p.config.Cluster)

	if p.config.ResourcePool != "" {
		ovftool_uri += "/Resources/" + p.config.ResourcePool
	}

	args, err := p.BuildArgs(source, ovftool_uri)
	if err != nil {
		ui.Message(fmt.Sprintf("Failed: %s\n", err))
	}

	ui.Message(fmt.Sprintf("Uploading %s to vSphere", source))

	log.Printf("Starting ovftool with parameters: %s", strings.Join(args, " "))
	cmd := exec.Command("ovftool", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, false, fmt.Errorf("Failed: %s\n", err)
	}

	return artifact, false, nil
}

func (p *PostProcessor) BuildArgs(source, ovftool_uri string) ([]string, error) {
	args := []string{
		"--acceptAllEulas",
		fmt.Sprintf(`--name=%s`, p.config.VMName),
		fmt.Sprintf(`--datastore=%s`, p.config.Datastore),
	}

	if p.config.Insecure {
		args = append(args, fmt.Sprintf(`--noSSLVerify=%t`, p.config.Insecure))
	}

	if p.config.DiskMode != "" {
		args = append(args, fmt.Sprintf(`--diskMode=%s`, p.config.DiskMode))
	}

	if p.config.VMFolder != "" {
		args = append(args, fmt.Sprintf(`--vmFolder=%s`, p.config.VMFolder))
	}

	if p.config.VMNetwork != "" {
		args = append(args, fmt.Sprintf(`--network=%s`, p.config.VMNetwork))
	}

	if p.config.Overwrite == true {
		args = append(args, "--overwrite")
	}

	if len(p.config.Options) > 0 {
		args = append(args, p.config.Options...)
	}

	args = append(args, fmt.Sprintf(`%s`, source))
	args = append(args, fmt.Sprintf(`%s`, ovftool_uri))

	return args, nil
}
