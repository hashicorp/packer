//go:generate mapstructure-to-hcl2 -type Config

package vsphere

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

var ovftool string = "ovftool"

var (
	// Regular expression to validate RFC1035 hostnames from full fqdn or simple hostname.
	// For example "packer-esxi1". Requires proper DNS setup and/or correct DNS search domain setting.
	hostnameRegex = regexp.MustCompile(`^[[:alnum:]][[:alnum:]\-]{0,61}[[:alnum:]]|[[:alpha:]]$`)

	// Simple regular expression to validate IPv4 values.
	// For example "192.168.1.1".
	ipv4Regex = regexp.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$`)
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Cluster      string   `mapstructure:"cluster"`
	Datacenter   string   `mapstructure:"datacenter"`
	Datastore    string   `mapstructure:"datastore"`
	DiskMode     string   `mapstructure:"disk_mode"`
	Host         string   `mapstructure:"host"`
	ESXiHost     string   `mapstructure:"esxi_host"`
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

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

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

	if runtime.GOOS == "windows" {
		ovftool = "ovftool.exe"
	}

	if _, err := exec.LookPath(ovftool); err != nil {
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

func (p *PostProcessor) PostProcess(ctx context.Context, ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, bool, error) {
	source := ""
	for _, path := range artifact.Files() {
		if strings.HasSuffix(path, ".vmx") || strings.HasSuffix(path, ".ovf") || strings.HasSuffix(path, ".ova") {
			source = path
			break
		}
	}

	if source == "" {
		return nil, false, false, fmt.Errorf("VMX, OVF or OVA file not found")
	}

	password := escapeWithSpaces(p.config.Password)
	ovftool_uri := fmt.Sprintf("vi://%s:%s@%s/%s/host/%s",
		escapeWithSpaces(p.config.Username),
		password,
		p.config.Host,
		p.config.Datacenter,
		p.config.Cluster)

	if p.config.ResourcePool != "" {
		ovftool_uri += "/Resources/" + p.config.ResourcePool
	}

	if p.config.ESXiHost != "" {
		if ipv4Regex.MatchString(p.config.ESXiHost) {
			ovftool_uri += "/?ip=" + p.config.ESXiHost
		} else if hostnameRegex.MatchString(p.config.ESXiHost) {
			ovftool_uri += "/?dns=" + p.config.ESXiHost
		}
	}

	args, err := p.BuildArgs(source, ovftool_uri)
	if err != nil {
		ui.Message(fmt.Sprintf("Failed: %s\n", err))
	}

	ui.Message(fmt.Sprintf("Uploading %s to vSphere", source))

	log.Printf("Starting ovftool with parameters: %s",
		strings.Replace(
			strings.Join(args, " "),
			password,
			"<password>",
			-1))

	var errWriter io.Writer
	var errOut bytes.Buffer
	cmd := exec.Command(ovftool, args...)
	errWriter = io.MultiWriter(os.Stderr, &errOut)
	cmd.Stdout = os.Stdout
	cmd.Stderr = errWriter

	if err := cmd.Run(); err != nil {
		err := fmt.Errorf("Error uploading virtual machine: %s\n%s\n", err, p.filterLog(errOut.String()))
		return nil, false, false, err
	}

	ui.Message(p.filterLog(errOut.String()))

	artifact = NewArtifact(p.config.Datastore, p.config.VMFolder, p.config.VMName, artifact.Files())

	return artifact, false, false, nil
}

func (p *PostProcessor) filterLog(s string) string {
	password := escapeWithSpaces(p.config.Password)
	return strings.Replace(s, password, "<password>", -1)
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

// Encode everything except for + signs
func escapeWithSpaces(stringToEscape string) string {
	escapedString := url.QueryEscape(stringToEscape)
	escapedString = strings.Replace(escapedString, "+", `%20`, -1)
	return escapedString
}
