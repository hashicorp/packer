//go:generate mapstructure-to-hcl2 -type Config

// This package implements a provisioner for Packer that uses
// Chef to provision the remote machine, specifically with chef-solo (that is,
// without a Chef server).
package chefsolo

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
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/provisioner"
	"github.com/hashicorp/packer/template/interpolate"
)

type guestOSTypeConfig struct {
	executeCommand string
	installCommand string
	stagingDir     string
}

var guestOSTypeConfigs = map[string]guestOSTypeConfig{
	provisioner.UnixOSType: {
		executeCommand: "{{if .Sudo}}sudo {{end}}chef-solo --no-color -c {{.ConfigPath}} -j {{.JsonPath}}",
		installCommand: "curl -L https://omnitruck.chef.io/install.sh | {{if .Sudo}}sudo {{end}}bash -s --{{if .Version}} -v {{.Version}}{{end}}",
		stagingDir:     "/tmp/packer-chef-solo",
	},
	provisioner.WindowsOSType: {
		executeCommand: "c:/opscode/chef/bin/chef-solo.bat --no-color -c {{.ConfigPath}} -j {{.JsonPath}}",
		installCommand: "powershell.exe -Command \". { iwr -useb https://omnitruck.chef.io/install.ps1 } | iex; Install-Project{{if .Version}} -version {{.Version}}{{end}}\"",
		stagingDir:     "C:/Windows/Temp/packer-chef-solo",
	},
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	ChefEnvironment            string   `mapstructure:"chef_environment"`
	ChefLicense                string   `mapstructure:"chef_license"`
	ConfigTemplate             string   `mapstructure:"config_template"`
	CookbookPaths              []string `mapstructure:"cookbook_paths"`
	RolesPath                  string   `mapstructure:"roles_path"`
	DataBagsPath               string   `mapstructure:"data_bags_path"`
	EncryptedDataBagSecretPath string   `mapstructure:"encrypted_data_bag_secret_path"`
	EnvironmentsPath           string   `mapstructure:"environments_path"`
	ExecuteCommand             string   `mapstructure:"execute_command"`
	InstallCommand             string   `mapstructure:"install_command"`
	RemoteCookbookPaths        []string `mapstructure:"remote_cookbook_paths"`
	Json                       map[string]interface{}
	PreventSudo                bool     `mapstructure:"prevent_sudo"`
	RunList                    []string `mapstructure:"run_list"`
	SkipInstall                bool     `mapstructure:"skip_install"`
	StagingDir                 string   `mapstructure:"staging_directory"`
	GuestOSType                string   `mapstructure:"guest_os_type"`
	Version                    string   `mapstructure:"version"`

	ctx interpolate.Context
}

type Provisioner struct {
	config            Config
	guestOSTypeConfig guestOSTypeConfig
	guestCommands     *provisioner.GuestCommands
}

type ConfigTemplate struct {
	CookbookPaths              string
	DataBagsPath               string
	EncryptedDataBagSecretPath string
	RolesPath                  string
	EnvironmentsPath           string
	ChefEnvironment            string
	ChefLicense                string

	// Templates don't support boolean statements until Go 1.2. In the
	// mean time, we do this.
	// TODO(mitchellh): Remove when Go 1.2 is released
	HasDataBagsPath               bool
	HasEncryptedDataBagSecretPath bool
	HasRolesPath                  bool
	HasEnvironmentsPath           bool
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

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"execute_command",
				"install_command",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	if p.config.GuestOSType == "" {
		p.config.GuestOSType = provisioner.DefaultOSType
	}
	p.config.GuestOSType = strings.ToLower(p.config.GuestOSType)

	var ok bool
	p.guestOSTypeConfig, ok = guestOSTypeConfigs[p.config.GuestOSType]
	if !ok {
		return fmt.Errorf("Invalid guest_os_type: \"%s\"", p.config.GuestOSType)
	}

	p.guestCommands, err = provisioner.NewGuestCommands(p.config.GuestOSType, !p.config.PreventSudo)
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

	if p.config.SkipInstall == false && p.config.InstallCommand == p.guestOSTypeConfig.installCommand {
		if p.config.ChefLicense == "" {
			p.config.ChefLicense = "accept-silent"
		}
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

	for _, path := range p.config.CookbookPaths {
		pFileInfo, err := os.Stat(path)

		if err != nil || !pFileInfo.IsDir() {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Bad cookbook path '%s': %s", path, err))
		}
	}

	if p.config.RolesPath != "" {
		pFileInfo, err := os.Stat(p.config.RolesPath)

		if err != nil || !pFileInfo.IsDir() {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Bad roles path '%s': %s", p.config.RolesPath, err))
		}
	}

	if p.config.DataBagsPath != "" {
		pFileInfo, err := os.Stat(p.config.DataBagsPath)

		if err != nil || !pFileInfo.IsDir() {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Bad data bags path '%s': %s", p.config.DataBagsPath, err))
		}
	}

	if p.config.EncryptedDataBagSecretPath != "" {
		pFileInfo, err := os.Stat(p.config.EncryptedDataBagSecretPath)

		if err != nil || pFileInfo.IsDir() {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Bad encrypted data bag secret '%s': %s", p.config.EncryptedDataBagSecretPath, err))
		}
	}

	if p.config.EnvironmentsPath != "" {
		pFileInfo, err := os.Stat(p.config.EnvironmentsPath)

		if err != nil || !pFileInfo.IsDir() {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Bad environments path '%s': %s", p.config.EnvironmentsPath, err))
		}
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

func (p *Provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator) error {
	ui.Say("Provisioning with chef-solo")

	if !p.config.SkipInstall {
		if err := p.installChef(ui, comm, p.config.Version); err != nil {
			return fmt.Errorf("Error installing Chef: %s", err)
		}
	}

	if err := p.createDir(ui, comm, p.config.StagingDir); err != nil {
		return fmt.Errorf("Error creating staging directory: %s", err)
	}

	cookbookPaths := make([]string, 0, len(p.config.CookbookPaths))
	for i, path := range p.config.CookbookPaths {
		targetPath := fmt.Sprintf("%s/cookbooks-%d", p.config.StagingDir, i)
		if err := p.uploadDirectory(ui, comm, targetPath, path); err != nil {
			return fmt.Errorf("Error uploading cookbooks: %s", err)
		}

		cookbookPaths = append(cookbookPaths, targetPath)
	}

	rolesPath := ""
	if p.config.RolesPath != "" {
		rolesPath = fmt.Sprintf("%s/roles", p.config.StagingDir)
		if err := p.uploadDirectory(ui, comm, rolesPath, p.config.RolesPath); err != nil {
			return fmt.Errorf("Error uploading roles: %s", err)
		}
	}

	dataBagsPath := ""
	if p.config.DataBagsPath != "" {
		dataBagsPath = fmt.Sprintf("%s/data_bags", p.config.StagingDir)
		if err := p.uploadDirectory(ui, comm, dataBagsPath, p.config.DataBagsPath); err != nil {
			return fmt.Errorf("Error uploading data bags: %s", err)
		}
	}

	encryptedDataBagSecretPath := ""
	if p.config.EncryptedDataBagSecretPath != "" {
		encryptedDataBagSecretPath = fmt.Sprintf("%s/encrypted_data_bag_secret", p.config.StagingDir)
		if err := p.uploadFile(ui, comm, encryptedDataBagSecretPath, p.config.EncryptedDataBagSecretPath); err != nil {
			return fmt.Errorf("Error uploading encrypted data bag secret: %s", err)
		}
	}

	environmentsPath := ""
	if p.config.EnvironmentsPath != "" {
		environmentsPath = fmt.Sprintf("%s/environments", p.config.StagingDir)
		if err := p.uploadDirectory(ui, comm, environmentsPath, p.config.EnvironmentsPath); err != nil {
			return fmt.Errorf("Error uploading environments: %s", err)
		}
	}

	configPath, err := p.createConfig(ui, comm, cookbookPaths, rolesPath, dataBagsPath, encryptedDataBagSecretPath, environmentsPath, p.config.ChefEnvironment, p.config.ChefLicense)
	if err != nil {
		return fmt.Errorf("Error creating Chef config file: %s", err)
	}

	jsonPath, err := p.createJson(ui, comm)
	if err != nil {
		return fmt.Errorf("Error creating JSON attributes: %s", err)
	}

	if err := p.executeChef(ui, comm, configPath, jsonPath); err != nil {
		return fmt.Errorf("Error executing Chef: %s", err)
	}

	return nil
}

func (p *Provisioner) uploadDirectory(ui packer.Ui, comm packer.Communicator, dst string, src string) error {
	if err := p.createDir(ui, comm, dst); err != nil {
		return err
	}

	// Make sure there is a trailing "/" so that the directory isn't
	// created on the other side.
	if src[len(src)-1] != '/' {
		src = src + "/"
	}

	return comm.UploadDir(dst, src, nil)
}

func (p *Provisioner) uploadFile(ui packer.Ui, comm packer.Communicator, dst string, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	return comm.Upload(dst, f, nil)
}

func (p *Provisioner) createConfig(ui packer.Ui, comm packer.Communicator, localCookbooks []string, rolesPath string, dataBagsPath string, encryptedDataBagSecretPath string, environmentsPath string, chefEnvironment string, chefLicense string) (string, error) {
	ui.Message("Creating configuration file 'solo.rb'")

	cookbook_paths := make([]string, len(p.config.RemoteCookbookPaths)+len(localCookbooks))
	for i, path := range p.config.RemoteCookbookPaths {
		cookbook_paths[i] = fmt.Sprintf(`"%s"`, path)
	}

	for i, path := range localCookbooks {
		i = len(p.config.RemoteCookbookPaths) + i
		cookbook_paths[i] = fmt.Sprintf(`"%s"`, path)
	}

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

	p.config.ctx.Data = &ConfigTemplate{
		CookbookPaths:                 strings.Join(cookbook_paths, ","),
		RolesPath:                     rolesPath,
		DataBagsPath:                  dataBagsPath,
		EncryptedDataBagSecretPath:    encryptedDataBagSecretPath,
		EnvironmentsPath:              environmentsPath,
		HasRolesPath:                  rolesPath != "",
		HasDataBagsPath:               dataBagsPath != "",
		HasEncryptedDataBagSecretPath: encryptedDataBagSecretPath != "",
		HasEnvironmentsPath:           environmentsPath != "",
		ChefEnvironment:               chefEnvironment,
		ChefLicense:                   chefLicense,
	}
	configString, err := interpolate.Render(tpl, &p.config.ctx)
	if err != nil {
		return "", err
	}

	remotePath := filepath.ToSlash(filepath.Join(p.config.StagingDir, "solo.rb"))
	if err := comm.Upload(remotePath, bytes.NewReader([]byte(configString)), nil); err != nil {
		return "", err
	}

	return remotePath, nil
}

func (p *Provisioner) createJson(ui packer.Ui, comm packer.Communicator) (string, error) {
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
	remotePath := filepath.ToSlash(filepath.Join(p.config.StagingDir, "node.json"))
	if err := comm.Upload(remotePath, bytes.NewReader(jsonBytes), nil); err != nil {
		return "", err
	}

	return remotePath, nil
}

func (p *Provisioner) createDir(ui packer.Ui, comm packer.Communicator, dir string) error {
	ui.Message(fmt.Sprintf("Creating directory: %s", dir))
	ctx := context.TODO()

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

func (p *Provisioner) executeChef(ui packer.Ui, comm packer.Communicator, config string, json string) error {
	p.config.ctx.Data = &ExecuteTemplate{
		ConfigPath: config,
		JsonPath:   json,
		Sudo:       !p.config.PreventSudo,
	}
	command, err := interpolate.Render(p.config.ExecuteCommand, &p.config.ctx)
	if err != nil {
		return err
	}

	ui.Message(fmt.Sprintf("Executing Chef: %s", command))

	cmd := &packer.RemoteCmd{
		Command: command,
	}
	ctx := context.TODO()
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus() != 0 {
		return fmt.Errorf("Non-zero exit status: %d", cmd.ExitStatus())
	}

	return nil
}

func (p *Provisioner) installChef(ui packer.Ui, comm packer.Communicator, version string) error {
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

var DefaultConfigTemplate = `
chef_license 		"{{.ChefLicense}}"
cookbook_path 	[{{.CookbookPaths}}]
{{if .HasRolesPath}}
role_path		"{{.RolesPath}}"
{{end}}
{{if .HasDataBagsPath}}
data_bag_path	"{{.DataBagsPath}}"
{{end}}
{{if .HasEncryptedDataBagSecretPath}}
encrypted_data_bag_secret "{{.EncryptedDataBagSecretPath}}"
{{end}}
{{if .HasEnvironmentsPath}}
environment_path "{{.EnvironmentsPath}}"
environment "{{.ChefEnvironment}}"
{{end}}
`
