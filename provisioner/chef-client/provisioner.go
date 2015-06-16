// This package implements a provisioner for Packer that uses
// Chef to provision the remote machine, specifically with chef-client (that is,
// with a Chef server).
package chefclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/common/uuid"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	ChefEnvironment      string `mapstructure:"chef_environment"`
	SslVerifyMode        string `mapstructure:"ssl_verify_mode"`
	ConfigTemplate       string `mapstructure:"config_template"`
	ExecuteCommand       string `mapstructure:"execute_command"`
	InstallCommand       string `mapstructure:"install_command"`
	Json                 map[string]interface{}
	NodeName             string   `mapstructure:"node_name"`
	PreventSudo          bool     `mapstructure:"prevent_sudo"`
	RunList              []string `mapstructure:"run_list"`
	ServerUrl            string   `mapstructure:"server_url"`
	SkipCleanClient      bool     `mapstructure:"skip_clean_client"`
	SkipCleanNode        bool     `mapstructure:"skip_clean_node"`
	SkipInstall          bool     `mapstructure:"skip_install"`
	StagingDir           string   `mapstructure:"staging_directory"`
	ClientKey            string   `mapstructure:"client_key"`
	ValidationKeyPath    string   `mapstructure:"validation_key_path"`
	ValidationClientName string   `mapstructure:"validation_client_name"`

	ctx interpolate.Context
}

type Provisioner struct {
	config Config
}

type ConfigTemplate struct {
	NodeName             string
	ServerUrl            string
	ClientKey            string
	ValidationKeyPath    string
	ValidationClientName string
	ChefEnvironment      string
	SslVerifyMode        string
}

type ExecuteTemplate struct {
	ConfigPath string
	JsonPath   string
	Sudo       bool
}

type InstallChefTemplate struct {
	Sudo bool
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate: true,
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

	if p.config.ExecuteCommand == "" {
		p.config.ExecuteCommand = "{{if .Sudo}}sudo {{end}}chef-client " +
			"--no-color -c {{.ConfigPath}} -j {{.JsonPath}}"
	}

	if p.config.InstallCommand == "" {
		p.config.InstallCommand = "curl -L " +
			"https://www.opscode.com/chef/install.sh | " +
			"{{if .Sudo}}sudo {{end}}bash"
	}

	if p.config.RunList == nil {
		p.config.RunList = make([]string, 0)
	}

	if p.config.StagingDir == "" {
		p.config.StagingDir = "/tmp/packer-chef-client"
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

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {

	nodeName := p.config.NodeName
	if nodeName == "" {
		nodeName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}
	remoteValidationKeyPath := ""
	serverUrl := p.config.ServerUrl

	if !p.config.SkipInstall {
		if err := p.installChef(ui, comm); err != nil {
			return fmt.Errorf("Error installing Chef: %s", err)
		}
	}

	if err := p.createDir(ui, comm, p.config.StagingDir); err != nil {
		return fmt.Errorf("Error creating staging directory: %s", err)
	}

	if p.config.ClientKey == "" {
		p.config.ClientKey = fmt.Sprintf("%s/client.pem", p.config.StagingDir)
	}

	if p.config.ValidationKeyPath != "" {
		remoteValidationKeyPath = fmt.Sprintf("%s/validation.pem", p.config.StagingDir)
		if err := p.copyValidationKey(ui, comm, remoteValidationKeyPath); err != nil {
			return fmt.Errorf("Error copying validation key: %s", err)
		}
	}

	configPath, err := p.createConfig(
		ui, comm, nodeName, serverUrl, p.config.ClientKey, remoteValidationKeyPath, p.config.ValidationClientName, p.config.ChefEnvironment, p.config.SslVerifyMode)
	if err != nil {
		return fmt.Errorf("Error creating Chef config file: %s", err)
	}

	jsonPath, err := p.createJson(ui, comm)
	if err != nil {
		return fmt.Errorf("Error creating JSON attributes: %s", err)
	}

	err = p.executeChef(ui, comm, configPath, jsonPath)
	if !p.config.SkipCleanNode {
		if err2 := p.cleanNode(ui, comm, nodeName); err2 != nil {
			return fmt.Errorf("Error cleaning up chef node: %s", err2)
		}
	}

	if !p.config.SkipCleanClient {
		if err2 := p.cleanClient(ui, comm, nodeName); err2 != nil {
			return fmt.Errorf("Error cleaning up chef client: %s", err2)
		}
	}

	if err != nil {
		return fmt.Errorf("Error executing Chef: %s", err)
	}

	if err := p.removeDir(ui, comm, p.config.StagingDir); err != nil {
		return fmt.Errorf("Error removing /etc/chef directory: %s", err)
	}

	return nil
}

func (p *Provisioner) Cancel() {
	// Just hard quit. It isn't a big deal if what we're doing keeps
	// running on the other side.
	os.Exit(0)
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

func (p *Provisioner) createConfig(ui packer.Ui, comm packer.Communicator, nodeName string, serverUrl string, clientKey string, remoteKeyPath string, validationClientName string, chefEnvironment string, sslVerifyMode string) (string, error) {
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

	ctx := p.config.ctx
	ctx.Data = &ConfigTemplate{
		NodeName:             nodeName,
		ServerUrl:            serverUrl,
		ClientKey:            clientKey,
		ValidationKeyPath:    remoteKeyPath,
		ValidationClientName: validationClientName,
		ChefEnvironment:      chefEnvironment,
		SslVerifyMode:        sslVerifyMode,
	}
	configString, err := interpolate.Render(tpl, &ctx)
	if err != nil {
		return "", err
	}

	remotePath := filepath.ToSlash(filepath.Join(p.config.StagingDir, "client.rb"))
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
	remotePath := filepath.ToSlash(filepath.Join(p.config.StagingDir, "first-boot.json"))
	if err := comm.Upload(remotePath, bytes.NewReader(jsonBytes), nil); err != nil {
		return "", err
	}

	return remotePath, nil
}

func (p *Provisioner) createDir(ui packer.Ui, comm packer.Communicator, dir string) error {
	ui.Message(fmt.Sprintf("Creating directory: %s", dir))

	mkdirCmd := fmt.Sprintf("mkdir -p '%s'", dir)
	if !p.config.PreventSudo {
		mkdirCmd = "sudo " + mkdirCmd
	}

	cmd := &packer.RemoteCmd{Command: mkdirCmd}
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Non-zero exit status. See output above for more info.")
	}

	// Chmod the directory to 0777 just so that we can access it as our user
	mkdirCmd = fmt.Sprintf("chmod 0777 '%s'", dir)
	if !p.config.PreventSudo {
		mkdirCmd = "sudo " + mkdirCmd
	}
	cmd = &packer.RemoteCmd{Command: mkdirCmd}
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Non-zero exit status. See output above for more info.")
	}

	return nil
}

func (p *Provisioner) cleanNode(ui packer.Ui, comm packer.Communicator, node string) error {
	ui.Say("Cleaning up chef node...")
	args := []string{"node", "delete", node}
	if err := p.knifeExec(ui, comm, node, args); err != nil {
		return fmt.Errorf("Failed to cleanup node: %s", err)
	}

	return nil
}

func (p *Provisioner) cleanClient(ui packer.Ui, comm packer.Communicator, node string) error {
	ui.Say("Cleaning up chef client...")
	args := []string{"client", "delete", node}
	if err := p.knifeExec(ui, comm, node, args); err != nil {
		return fmt.Errorf("Failed to cleanup client: %s", err)
	}

	return nil
}

func (p *Provisioner) knifeExec(ui packer.Ui, comm packer.Communicator, node string, args []string) error {
	flags := []string{
		"-y",
		"-s", fmt.Sprintf("'%s'", p.config.ServerUrl),
		"-k", fmt.Sprintf("'%s'", p.config.ClientKey),
		"-u", fmt.Sprintf("'%s'", node),
	}

	cmdText := fmt.Sprintf(
		"knife %s %s", strings.Join(args, " "), strings.Join(flags, " "))
	if !p.config.PreventSudo {
		cmdText = "sudo " + cmdText
	}

	cmd := &packer.RemoteCmd{Command: cmdText}
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus != 0 {
		return fmt.Errorf(
			"Non-zero exit status. See output above for more info.\n\n"+
				"Command: %s",
			cmdText)
	}

	return nil
}

func (p *Provisioner) removeDir(ui packer.Ui, comm packer.Communicator, dir string) error {
	ui.Message(fmt.Sprintf("Removing directory: %s", dir))

	rmCmd := fmt.Sprintf("rm -rf '%s'", dir)
	if !p.config.PreventSudo {
		rmCmd = "sudo " + rmCmd
	}

	cmd := &packer.RemoteCmd{
		Command: rmCmd,
	}

	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
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

	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Non-zero exit status: %d", cmd.ExitStatus)
	}

	return nil
}

func (p *Provisioner) installChef(ui packer.Ui, comm packer.Communicator) error {
	ui.Message("Installing Chef...")

	p.config.ctx.Data = &InstallChefTemplate{
		Sudo: !p.config.PreventSudo,
	}
	command, err := interpolate.Render(p.config.InstallCommand, &p.config.ctx)
	if err != nil {
		return err
	}

	cmd := &packer.RemoteCmd{Command: command}
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus != 0 {
		return fmt.Errorf(
			"Install script exited with non-zero exit status %d", cmd.ExitStatus)
	}

	return nil
}

func (p *Provisioner) copyValidationKey(ui packer.Ui, comm packer.Communicator, remotePath string) error {
	ui.Message("Uploading validation key...")

	// First upload the validation key to a writable location
	f, err := os.Open(p.config.ValidationKeyPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := comm.Upload(remotePath, f, nil); err != nil {
		return err
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
log_level        :info
log_location     STDOUT
chef_server_url  "{{.ServerUrl}}"
client_key       "{{.ClientKey}}"
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
{{if ne .SslVerifyMode ""}}
ssl_verify_mode :{{.SslVerifyMode}}
{{end}}
`
