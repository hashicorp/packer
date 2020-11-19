//go:generate mapstructure-to-hcl2 -type Config

package vsphere

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	shelllocal "github.com/hashicorp/packer/packer-plugin-sdk/shell-local"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
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
		PluginType:         BuilderId,
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

func (p *PostProcessor) generateURI() (*url.URL, error) {
	// use net/url lib to encode and escape url elements
	ovftool_uri := fmt.Sprintf("vi://%s/%s/host/%s",
		p.config.Host,
		p.config.Datacenter,
		p.config.Cluster)

	if p.config.ResourcePool != "" {
		ovftool_uri += "/Resources/" + p.config.ResourcePool
	}

	u, err := url.Parse(ovftool_uri)
	if err != nil {
		return nil, fmt.Errorf("Couldn't generate uri for ovftool: %s", err)
	}
	u.User = url.UserPassword(p.config.Username, p.config.Password)

	if p.config.ESXiHost != "" {
		q := u.Query()
		if ipv4Regex.MatchString(p.config.ESXiHost) {
			q.Add("ip", p.config.ESXiHost)
		} else if hostnameRegex.MatchString(p.config.ESXiHost) {
			q.Add("dns", p.config.ESXiHost)
		}
		u.RawQuery = q.Encode()
	}
	return u, nil
}

func getEncodedPassword(u *url.URL) (string, bool) {
	// filter password from all logging
	password, passwordSet := u.User.Password()
	if passwordSet && password != "" {
		encodedPassword := strings.Split(u.User.String(), ":")[1]
		return encodedPassword, true
	}
	return password, false
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packersdk.Ui, artifact packer.Artifact) (packer.Artifact, bool, bool, error) {
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

	ovftool_uri, err := p.generateURI()
	if err != nil {
		return nil, false, false, err
	}
	encodedPassword, isSet := getEncodedPassword(ovftool_uri)
	if isSet {
		packer.LogSecretFilter.Set(encodedPassword)
	}

	args, err := p.BuildArgs(source, ovftool_uri.String())
	if err != nil {
		ui.Message(fmt.Sprintf("Failed: %s\n", err))
	}

	ui.Message(fmt.Sprintf("Uploading %s to vSphere", source))

	log.Printf("Starting ovftool with parameters: %s", strings.Join(args, " "))

	ui.Message("Validating Username and Password with dry-run")
	err = p.ValidateOvfTool(args, ovftool)
	if err != nil {
		return nil, false, false, err
	}

	// Validation has passed, so run for real.
	ui.Message("Calling OVFtool to upload vm")
	commandAndArgs := []string{ovftool}
	commandAndArgs = append(commandAndArgs, args...)
	comm := &shelllocal.Communicator{
		ExecuteCommand: commandAndArgs,
	}
	flattenedCmd := strings.Join(commandAndArgs, " ")
	cmd := &packer.RemoteCmd{Command: flattenedCmd}
	log.Printf("[INFO] (vsphere): starting ovftool command: %s", flattenedCmd)
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return nil, false, false, fmt.Errorf(
			"Error uploading virtual machine: Please see output above for more information.")
	}

	artifact = NewArtifact(p.config.Datastore, p.config.VMFolder, p.config.VMName, artifact.Files())

	return artifact, false, false, nil
}

func (p *PostProcessor) ValidateOvfTool(args []string, ofvtool string) error {
	args = append([]string{"--verifyOnly"}, args...)
	var out bytes.Buffer
	cmdCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, ovftool, args...)
	cmd.Stdout = &out

	// Need to manually close stdin or else the ofvtool call will hang
	// forever in a situation where the user has provided an invalid
	// password or username
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	defer stdin.Close()

	if err := cmd.Run(); err != nil {
		outString := out.String()
		if strings.Contains(outString, "Enter login information for") {
			err = fmt.Errorf("Error performing OVFtool dry run; the username " +
				"or password you provided to ovftool is likely invalid.")
			return err
		}
		return nil
	}
	return nil
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

	args = append(args, source)
	args = append(args, ovftool_uri)

	return args, nil
}
