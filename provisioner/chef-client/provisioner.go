//go:generate mapstructure-to-hcl2 -type Config

// This package implements a provisioner for Packer that uses
// Chef to provision the remote machine, specifically with chef-client (that is,
// with a Chef server).
package chefclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/guestexec"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer/packer-plugin-sdk/uuid"
)

type guestOSTypeConfig struct {
	executeCommand string
	installCommand string
	knifeCommand   string
	stagingDir     string
}

var guestOSTypeConfigs = map[string]guestOSTypeConfig{
	guestexec.UnixOSType: {
		executeCommand: "{{if .Sudo}}sudo {{end}}chef-client --no-color -c {{.ConfigPath}} -j {{.JsonPath}}",
		installCommand: "curl -L https://omnitruck.chef.io/install.sh | {{if .Sudo}}sudo {{end}}bash -s --{{if .Version}} -v {{.Version}}{{end}}",
		knifeCommand:   "{{if .Sudo}}sudo {{end}}knife {{.Args}} {{.Flags}}",
		stagingDir:     "/tmp/packer-chef-client",
	},
	guestexec.WindowsOSType: {
		executeCommand: "c:/opscode/chef/bin/chef-client.bat --no-color -c {{.ConfigPath}} -j {{.JsonPath}}",
		installCommand: "powershell.exe -Command \". { iwr -useb https://omnitruck.chef.io/install.ps1 } | iex; Install-Project{{if .Version}} -version {{.Version}}{{end}}\"",
		knifeCommand:   "c:/opscode/chef/bin/knife.bat {{.Args}} {{.Flags}}",
		stagingDir:     "C:/Windows/Temp/packer-chef-client",
	},
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Json map[string]interface{}

	ChefEnvironment            string   `mapstructure:"chef_environment"`
	ChefLicense                string   `mapstructure:"chef_license"`
	ClientKey                  string   `mapstructure:"client_key"`
	ConfigTemplate             string   `mapstructure:"config_template"`
	ElevatedUser               string   `mapstructure:"elevated_user"`
	ElevatedPassword           string   `mapstructure:"elevated_password"`
	EncryptedDataBagSecretPath string   `mapstructure:"encrypted_data_bag_secret_path"`
	ExecuteCommand             string   `mapstructure:"execute_command"`
	GuestOSType                string   `mapstructure:"guest_os_type"`
	InstallCommand             string   `mapstructure:"install_command"`
	KnifeCommand               string   `mapstructure:"knife_command"`
	NodeName                   string   `mapstructure:"node_name"`
	PolicyGroup                string   `mapstructure:"policy_group"`
	PolicyName                 string   `mapstructure:"policy_name"`
	PreventSudo                bool     `mapstructure:"prevent_sudo"`
	RunList                    []string `mapstructure:"run_list"`
	ServerUrl                  string   `mapstructure:"server_url"`
	SkipCleanClient            bool     `mapstructure:"skip_clean_client"`
	SkipCleanNode              bool     `mapstructure:"skip_clean_node"`
	SkipCleanStagingDirectory  bool     `mapstructure:"skip_clean_staging_directory"`
	SkipInstall                bool     `mapstructure:"skip_install"`
	SslVerifyMode              string   `mapstructure:"ssl_verify_mode"`
	TrustedCertsDir            string   `mapstructure:"trusted_certs_dir"`
	StagingDir                 string   `mapstructure:"staging_directory"`
	ValidationClientName       string   `mapstructure:"validation_client_name"`
	ValidationKeyPath          string   `mapstructure:"validation_key_path"`
	Version                    string   `mapstructure:"version"`

	ctx interpolate.Context
}

type Provisioner struct {
	config            Config
	communicator      packer.Communicator
	guestOSTypeConfig guestOSTypeConfig
	guestCommands     *guestexec.GuestCommands
	generatedData     map[string]interface{}
}

type ConfigTemplate struct {
	ChefEnvironment            string
	ChefLicense                string
	ClientKey                  string
	EncryptedDataBagSecretPath string
	NodeName                   string
	PolicyGroup                string
	PolicyName                 string
	ServerUrl                  string
	SslVerifyMode              string
	TrustedCertsDir            string
	ValidationClientName       string
	ValidationKeyPath          string
}

type ExecuteTemplate struct {
	ConfigPath string
	JsonPath   string
	Sudo       bool
}

type InstallChefTemplate struct {
	Sudo    bool
	Version string
}

type KnifeTemplate struct {
	Sudo  bool
	Flags string
	Args  string
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         "chef-client",
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"execute_command",
				"install_command",
				"knife_command",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	if p.config.GuestOSType == "" {
		p.config.GuestOSType = guestexec.DefaultOSType
	}
	p.config.GuestOSType = strings.ToLower(p.config.GuestOSType)

	var ok bool
	p.guestOSTypeConfig, ok = guestOSTypeConfigs[p.config.GuestOSType]
	if !ok {
		return fmt.Errorf("Invalid guest_os_type: \"%s\"", p.config.GuestOSType)
	}

	p.guestCommands, err = guestexec.NewGuestCommands(p.config.GuestOSType, !p.config.PreventSudo)
	if err != nil {
		return fmt.Errorf("Invalid guest_os_type: \"%s\"", p.config.GuestOSType)
	}

	if p.config.ExecuteCommand == "" {
		p.config.ExecuteCommand = p.guestOSTypeConfig.executeCommand
	}

	if p.config.InstallCommand == "" {
		p.config.InstallCommand = p.guestOSTypeConfig.installCommand
	}

	if p.config.RunList == nil {
		p.config.RunList = make([]string, 0)
	}

	if p.config.StagingDir == "" {
		p.config.StagingDir = p.guestOSTypeConfig.stagingDir
	}

	if p.config.KnifeCommand == "" {
		p.config.KnifeCommand = p.guestOSTypeConfig.knifeCommand
	}

	var errs *packer.MultiError
	if p.config.ConfigTemplate != "" {
		fi, err := os.Stat(p.config.ConfigTemplate)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Bad config template path: %s", err))
		} else if fi.IsDir() {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Config template path must be a file: %s", err))
		}
	}

	if p.config.ServerUrl == "" {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("server_url must be set"))
	}

	if p.config.SkipInstall == false && p.config.InstallCommand == p.guestOSTypeConfig.installCommand {
		if p.config.ChefLicense == "" {
			p.config.ChefLicense = "accept-silent"
		}
	}

	if p.config.EncryptedDataBagSecretPath != "" {
		pFileInfo, err := os.Stat(p.config.EncryptedDataBagSecretPath)

		if err != nil || pFileInfo.IsDir() {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Bad encrypted data bag secret '%s': %s", p.config.EncryptedDataBagSecretPath, err))
		}
	}

	if (p.config.PolicyName != "") != (p.config.PolicyGroup != "") {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("If either policy_name or policy_group are set, they must both be set."))
	}

	jsonValid := true
	for k, v := range p.config.Json {
		p.config.Json[k], err = p.deepJsonFix(k, v)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing JSON: %s", err))
			jsonValid = false
		}
	}

	if jsonValid {
		// Process the user variables within the JSON and set the JSON.
		// Do this early so that we can validate and show errors.
		p.config.Json, err = p.processJsonUserVars()
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing user variables in JSON: %s", err))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *Provisioner) Provision(ctx context.Context, ui packersdk.Ui, comm packer.Communicator, generatedData map[string]interface{}) error {
	p.generatedData = generatedData
	p.communicator = comm

	nodeName := p.config.NodeName
	if nodeName == "" {
		nodeName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}
	remoteValidationKeyPath := ""
	serverUrl := p.config.ServerUrl

	if !p.config.SkipInstall {
		if err := p.installChef(ui, comm, p.config.Version); err != nil {
			return fmt.Errorf("Error installing Chef: %s", err)
		}
	}

	if err := p.createDir(ui, comm, p.config.StagingDir); err != nil {
		return fmt.Errorf("Error creating staging directory: %s", err)
	}

	if p.config.ClientKey == "" {
		p.config.ClientKey = fmt.Sprintf("%s/client.pem", p.config.StagingDir)
	}

	encryptedDataBagSecretPath := ""
	if p.config.EncryptedDataBagSecretPath != "" {
		encryptedDataBagSecretPath = fmt.Sprintf("%s/encrypted_data_bag_secret", p.config.StagingDir)
		if err := p.uploadFile(ui,
			comm,
			encryptedDataBagSecretPath,
			p.config.EncryptedDataBagSecretPath); err != nil {
			return fmt.Errorf("Error uploading encrypted data bag secret: %s", err)
		}
	}

	if p.config.ValidationKeyPath != "" {
		path, err := packer.ExpandUser(p.config.ValidationKeyPath)
		if err != nil {
			return fmt.Errorf("Error while expanding a tilde in the validation key: %s", err)
		}
		remoteValidationKeyPath = fmt.Sprintf("%s/validation.pem", p.config.StagingDir)
		if err := p.uploadFile(ui, comm, remoteValidationKeyPath, path); err != nil {
			return fmt.Errorf("Error copying validation key: %s", err)
		}
	}

	configPath, err := p.createConfig(
		ui,
		comm,
		nodeName,
		serverUrl,
		p.config.ClientKey,
		p.config.ChefLicense,
		encryptedDataBagSecretPath,
		remoteValidationKeyPath,
		p.config.ValidationClientName,
		p.config.ChefEnvironment,
		p.config.PolicyGroup,
		p.config.PolicyName,
		p.config.SslVerifyMode,
		p.config.TrustedCertsDir)
	if err != nil {
		return fmt.Errorf("Error creating Chef config file: %s", err)
	}

	jsonPath, err := p.createJson(ui, comm)
	if err != nil {
		return fmt.Errorf("Error creating JSON attributes: %s", err)
	}

	err = p.executeChef(ui, comm, configPath, jsonPath)

	if !(p.config.SkipCleanNode && p.config.SkipCleanClient) {

		knifeConfigPath, knifeErr := p.createKnifeConfig(
			ui, comm, nodeName, serverUrl, p.config.ClientKey, p.config.SslVerifyMode, p.config.TrustedCertsDir)

		if knifeErr != nil {
			return fmt.Errorf("Error creating knife config on node: %s", knifeErr)
		}

		if !p.config.SkipCleanNode {
			if err := p.cleanNode(ui, comm, nodeName, knifeConfigPath); err != nil {
				return fmt.Errorf("Error cleaning up chef node: %s", err)
			}
		}

		if !p.config.SkipCleanClient {
			if err := p.cleanClient(ui, comm, nodeName, knifeConfigPath); err != nil {
				return fmt.Errorf("Error cleaning up chef client: %s", err)
			}
		}
	}

	if err != nil {
		return fmt.Errorf("Error executing Chef: %s", err)
	}

	if !p.config.SkipCleanStagingDirectory {
		if err := p.removeDir(ui, comm, p.config.StagingDir); err != nil {
			return fmt.Errorf("Error removing %s: %s", p.config.StagingDir, err)
		}
	}

	return nil
}

func (p *Provisioner) uploadFile(ui packersdk.Ui, comm packer.Communicator, remotePath string, localPath string) error {
	ui.Message(fmt.Sprintf("Uploading %s...", localPath))

	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return comm.Upload(remotePath, f, nil)
}

func (p *Provisioner) createConfig(
	ui packersdk.Ui,
	comm packer.Communicator,
	nodeName string,
	serverUrl string,
	clientKey string,
	chefLicense string,
	encryptedDataBagSecretPath,
	remoteKeyPath string,
	validationClientName string,
	chefEnvironment string,
	policyGroup string,
	policyName string,
	sslVerifyMode string,
	trustedCertsDir string) (string, error) {

	ui.Message("Creating configuration file 'client.rb'")

	// Read the template
	tpl := DefaultConfigTemplate
	if p.config.ConfigTemplate != "" {
		f, err := os.Open(p.config.ConfigTemplate)
		if err != nil {
			return "", err
		}
		defer f.Close()

		tplBytes, err := ioutil.ReadAll(f)
		if err != nil {
			return "", err
		}

		tpl = string(tplBytes)
	}

	ictx := p.config.ctx
	ictx.Data = &ConfigTemplate{
		NodeName:                   nodeName,
		ServerUrl:                  serverUrl,
		ClientKey:                  clientKey,
		ChefLicense:                chefLicense,
		ValidationKeyPath:          remoteKeyPath,
		ValidationClientName:       validationClientName,
		ChefEnvironment:            chefEnvironment,
		PolicyGroup:                policyGroup,
		PolicyName:                 policyName,
		SslVerifyMode:              sslVerifyMode,
		TrustedCertsDir:            trustedCertsDir,
		EncryptedDataBagSecretPath: encryptedDataBagSecretPath,
	}
	configString, err := interpolate.Render(tpl, &ictx)
	if err != nil {
		return "", err
	}

	remotePath := filepath.ToSlash(filepath.Join(p.config.StagingDir, "client.rb"))
	if err := comm.Upload(remotePath, bytes.NewReader([]byte(configString)), nil); err != nil {
		return "", err
	}

	return remotePath, nil
}

func (p *Provisioner) createKnifeConfig(ui packersdk.Ui, comm packer.Communicator, nodeName string, serverUrl string, clientKey string, sslVerifyMode string, trustedCertsDir string) (string, error) {
	ui.Message("Creating configuration file 'knife.rb'")

	// Read the template
	tpl := DefaultKnifeTemplate

	ictx := p.config.ctx
	ictx.Data = &ConfigTemplate{
		NodeName:        nodeName,
		ServerUrl:       serverUrl,
		ClientKey:       clientKey,
		SslVerifyMode:   sslVerifyMode,
		TrustedCertsDir: trustedCertsDir,
	}
	configString, err := interpolate.Render(tpl, &ictx)
	if err != nil {
		return "", err
	}

	remotePath := filepath.ToSlash(filepath.Join(p.config.StagingDir, "knife.rb"))
	if err := comm.Upload(remotePath, bytes.NewReader([]byte(configString)), nil); err != nil {
		return "", err
	}

	return remotePath, nil
}

func (p *Provisioner) createJson(ui packersdk.Ui, comm packer.Communicator) (string, error) {
	ui.Message("Creating JSON attribute file")

	jsonData := make(map[string]interface{})

	// Copy the configured JSON
	for k, v := range p.config.Json {
		jsonData[k] = v
	}

	// Set the run list if it was specified
	if len(p.config.RunList) > 0 {
		jsonData["run_list"] = p.config.RunList
	}

	jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return "", err
	}

	// Upload the bytes
	remotePath := filepath.ToSlash(filepath.Join(p.config.StagingDir, "first-boot.json"))
	if err := comm.Upload(remotePath, bytes.NewReader(jsonBytes), nil); err != nil {
		return "", err
	}

	return remotePath, nil
}

func (p *Provisioner) createDir(ui packersdk.Ui, comm packer.Communicator, dir string) error {
	ctx := context.TODO()
	ui.Message(fmt.Sprintf("Creating directory: %s", dir))

	cmd := &packer.RemoteCmd{Command: p.guestCommands.CreateDir(dir)}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus() != 0 {
		return fmt.Errorf("Non-zero exit status. See output above for more info.")
	}

	// Chmod the directory to 0777 just so that we can access it as our user
	cmd = &packer.RemoteCmd{Command: p.guestCommands.Chmod(dir, "0777")}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus() != 0 {
		return fmt.Errorf("Non-zero exit status. See output above for more info.")
	}

	return nil
}

func (p *Provisioner) cleanNode(ui packersdk.Ui, comm packer.Communicator, node string, knifeConfigPath string) error {
	ui.Say("Cleaning up chef node...")
	args := []string{"node", "delete", node}
	if err := p.knifeExec(ui, comm, node, knifeConfigPath, args); err != nil {
		return fmt.Errorf("Failed to cleanup node: %s", err)
	}

	return nil
}

func (p *Provisioner) cleanClient(ui packersdk.Ui, comm packer.Communicator, node string, knifeConfigPath string) error {
	ui.Say("Cleaning up chef client...")
	args := []string{"client", "delete", node}
	if err := p.knifeExec(ui, comm, node, knifeConfigPath, args); err != nil {
		return fmt.Errorf("Failed to cleanup client: %s", err)
	}

	return nil
}

func (p *Provisioner) knifeExec(ui packersdk.Ui, comm packer.Communicator, node string, knifeConfigPath string, args []string) error {
	flags := []string{
		"-y",
		"-c", knifeConfigPath,
	}
	ctx := context.TODO()

	p.config.ctx.Data = &KnifeTemplate{
		Sudo:  !p.config.PreventSudo,
		Flags: strings.Join(flags, " "),
		Args:  strings.Join(args, " "),
	}

	command, err := interpolate.Render(p.config.KnifeCommand, &p.config.ctx)
	if err != nil {
		return err
	}

	cmd := &packer.RemoteCmd{Command: command}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus() != 0 {
		return fmt.Errorf(
			"Non-zero exit status. See output above for more info.\n\n"+
				"Command: %s",
			command)
	}

	return nil
}

func (p *Provisioner) removeDir(ui packersdk.Ui, comm packer.Communicator, dir string) error {
	ui.Message(fmt.Sprintf("Removing directory: %s", dir))
	ctx := context.TODO()

	cmd := &packer.RemoteCmd{Command: p.guestCommands.RemoveDir(dir)}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}

	return nil
}

func (p *Provisioner) executeChef(ui packersdk.Ui, comm packer.Communicator, config string, json string) error {
	p.config.ctx.Data = &ExecuteTemplate{
		ConfigPath: config,
		JsonPath:   json,
		Sudo:       !p.config.PreventSudo,
	}
	ctx := context.TODO()

	command, err := interpolate.Render(p.config.ExecuteCommand, &p.config.ctx)
	if err != nil {
		return err
	}

	if p.config.ElevatedUser != "" {
		command, err = guestexec.GenerateElevatedRunner(command, p)
		if err != nil {
			return err
		}
	}

	ui.Message(fmt.Sprintf("Executing Chef: %s", command))

	cmd := &packer.RemoteCmd{
		Command: command,
	}

	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus() != 0 {
		return fmt.Errorf("Non-zero exit status: %d", cmd.ExitStatus())
	}

	return nil
}

func (p *Provisioner) installChef(ui packersdk.Ui, comm packer.Communicator, version string) error {
	ui.Message("Installing Chef...")
	ctx := context.TODO()

	p.config.ctx.Data = &InstallChefTemplate{
		Sudo:    !p.config.PreventSudo,
		Version: version,
	}
	command, err := interpolate.Render(p.config.InstallCommand, &p.config.ctx)
	if err != nil {
		return err
	}

	ui.Message(command)

	cmd := &packer.RemoteCmd{Command: command}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus() != 0 {
		return fmt.Errorf(
			"Install script exited with non-zero exit status %d", cmd.ExitStatus())
	}

	return nil
}

func (p *Provisioner) deepJsonFix(key string, current interface{}) (interface{}, error) {
	if current == nil {
		return nil, nil
	}

	switch c := current.(type) {
	case []interface{}:
		val := make([]interface{}, len(c))
		for i, v := range c {
			var err error
			val[i], err = p.deepJsonFix(fmt.Sprintf("%s[%d]", key, i), v)
			if err != nil {
				return nil, err
			}
		}

		return val, nil
	case []uint8:
		return string(c), nil
	case map[interface{}]interface{}:
		val := make(map[string]interface{})
		for k, v := range c {
			ks, ok := k.(string)
			if !ok {
				return nil, fmt.Errorf("%s: key is not string", key)
			}

			var err error
			val[ks], err = p.deepJsonFix(
				fmt.Sprintf("%s.%s", key, ks), v)
			if err != nil {
				return nil, err
			}
		}

		return val, nil
	default:
		return current, nil
	}
}

func (p *Provisioner) processJsonUserVars() (map[string]interface{}, error) {
	jsonBytes, err := json.Marshal(p.config.Json)
	if err != nil {
		// This really shouldn't happen since we literally just unmarshalled
		panic(err)
	}

	// Copy the user variables so that we can restore them later, and
	// make sure we make the quotes JSON-friendly in the user variables.
	originalUserVars := make(map[string]string)
	for k, v := range p.config.ctx.UserVariables {
		originalUserVars[k] = v
	}

	// Make sure we reset them no matter what
	defer func() {
		p.config.ctx.UserVariables = originalUserVars
	}()

	// Make the current user variables JSON string safe.
	for k, v := range p.config.ctx.UserVariables {
		v = strings.Replace(v, `\`, `\\`, -1)
		v = strings.Replace(v, `"`, `\"`, -1)
		p.config.ctx.UserVariables[k] = v
	}

	// Process the bytes with the template processor
	p.config.ctx.Data = nil
	jsonBytesProcessed, err := interpolate.Render(string(jsonBytes), &p.config.ctx)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonBytesProcessed), &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (p *Provisioner) Communicator() packer.Communicator {
	return p.communicator
}

func (p *Provisioner) ElevatedUser() string {
	return p.config.ElevatedUser
}

func (p *Provisioner) ElevatedPassword() string {
	// Replace ElevatedPassword for winrm users who used this feature
	p.config.ctx.Data = p.generatedData

	elevatedPassword, _ := interpolate.Render(p.config.ElevatedPassword, &p.config.ctx)

	return elevatedPassword
}

var DefaultConfigTemplate = `
log_level        :info
log_location     STDOUT
chef_server_url  "{{.ServerUrl}}"
client_key       "{{.ClientKey}}"
chef_license     "{{.ChefLicense}}"
{{if ne .EncryptedDataBagSecretPath ""}}
encrypted_data_bag_secret "{{.EncryptedDataBagSecretPath}}"
{{end}}
{{if ne .ValidationClientName ""}}
validation_client_name "{{.ValidationClientName}}"
{{else}}
validation_client_name "chef-validator"
{{end}}
{{if ne .ValidationKeyPath ""}}
validation_key "{{.ValidationKeyPath}}"
{{end}}
node_name "{{.NodeName}}"
{{if ne .ChefEnvironment ""}}
environment "{{.ChefEnvironment}}"
{{end}}
{{if ne .PolicyGroup ""}}
policy_group "{{.PolicyGroup}}"
{{end}}
{{if ne .PolicyName ""}}
policy_name "{{.PolicyName}}"
{{end}}
{{if ne .SslVerifyMode ""}}
ssl_verify_mode :{{.SslVerifyMode}}
{{end}}
{{if ne .TrustedCertsDir ""}}
trusted_certs_dir "{{.TrustedCertsDir}}"
{{end}}
`

var DefaultKnifeTemplate = `
log_level        :info
log_location     STDOUT
chef_server_url  "{{.ServerUrl}}"
client_key       "{{.ClientKey}}"
node_name "{{.NodeName}}"
{{if ne .SslVerifyMode ""}}
ssl_verify_mode :{{.SslVerifyMode}}
{{end}}
{{if ne .TrustedCertsDir ""}}
trusted_certs_dir "{{.TrustedCertsDir}}"
{{end}}
`
