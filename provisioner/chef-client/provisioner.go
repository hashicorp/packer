// This package implements a provisioner for Packer that uses
// Chef to provision the remote machine, specifically with chef-client (that is,
// with a Chef server).
package chefclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"os/exec"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	ConfigTemplate      string   `mapstructure:"config_template"`
	CookbookPaths       []string `mapstructure:"cookbook_paths"`
	RolesPath           string   `mapstructure:"roles_path"`
	NodeName	    string   `mapstructure:"node_name"`
	ServerUrl	    string   `mapstructure:"chef_server_url"`
	DataBagsPath        string   `mapstructure:"data_bags_path"`
	ExecuteCommand      string   `mapstructure:"execute_command"`
	InstallCommand      string   `mapstructure:"install_command"`
	ValidationCommand   string   `mapstructure:"validation_command"` 
        ClientCommand	    string   `mapstructure:"client_command"`
	RemoteCookbookPaths []string `mapstructure:"remote_cookbook_paths"`
	Json                map[string]interface{}
	PreventSudo         bool     `mapstructure:"prevent_sudo"`
	RunList             []string `mapstructure:"run_list"`
	SkipInstall         bool     `mapstructure:"skip_install"`
	StagingDir          string   `mapstructure:"staging_directory"`

	tpl *packer.ConfigTemplate
}

type Provisioner struct {
	config Config
}

type ConfigTemplate struct {
	CookbookPaths string
	DataBagsPath  string
	RolesPath     string
	NodeName      string
	ServerUrl     string

	// Templates don't support boolean statements until Go 1.2. In the
	// mean time, we do this.
	// TODO(mitchellh): Remove when Go 1.2 is released
	HasDataBagsPath bool
	HasRolesPath    bool
	HasNodeName     bool
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
	md, err := common.DecodeConfig(&p.config, raws...)
	if err != nil {
		return err
	}

	p.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	p.config.tpl.UserVars = p.config.PackerUserVars

	if p.config.ExecuteCommand == "" {
		p.config.ExecuteCommand = "{{if .Sudo}}sudo {{end}}chef-client --no-color -c {{.ConfigPath}} -j {{.JsonPath}}"
	}

	if p.config.InstallCommand == "" {
		p.config.InstallCommand = "curl -L https://www.opscode.com/chef/install.sh | {{if .Sudo}}sudo {{end}}bash  -s -- -v 10.26.0"
	}

	if p.config.ValidationCommand == "" {
		p.config.ValidationCommand = "{{if .Sudo}}sudo {{end}} mv /tmp/validation.pem /etc/chef/validation.pem"
	}

	if p.config.ClientCommand == "" {
               // p.config.ClientCommand = "{{if .Sudo}}sudo {{end}} mv /tmp/client.rb /etc/chef/client.rb"
        }


	if p.config.RunList == nil {
		p.config.RunList = make([]string, 0)
	}

	if p.config.StagingDir == "" {
		p.config.StagingDir = "/tmp/packer-chef-client"
	}

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)

	templates := map[string]*string{
		"config_template": &p.config.ConfigTemplate,
		"data_bags_path":  &p.config.DataBagsPath,
		"roles_path":      &p.config.RolesPath,
		"node_name":	   &p.config.NodeName,
		"staging_dir":     &p.config.StagingDir,
		"chef_server_url": &p.config.ServerUrl,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = p.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	sliceTemplates := map[string][]string{
		"cookbook_paths":        p.config.CookbookPaths,
		"remote_cookbook_paths": p.config.RemoteCookbookPaths,
		"run_list":              p.config.RunList,
	}

	for n, slice := range sliceTemplates {
		for i, elem := range slice {
			var err error
			slice[i], err = p.config.tpl.Process(elem, nil)
			if err != nil {
				errs = packer.MultiErrorAppend(
					errs, fmt.Errorf("Error processing %s[%d]: %s", n, i, err))
			}
		}
	}

	validates := map[string]*string{
		"execute_command": &p.config.ExecuteCommand,
		"install_command": &p.config.InstallCommand,
		"validation_command": &p.config.ValidationCommand,
		"client_command": &p.config.ClientCommand,
	}

	for n, ptr := range validates {
		if err := p.config.tpl.Validate(*ptr); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error parsing %s: %s", n, err))
		}
	}

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

	// Process the user variables within the JSON and set the JSON.
	// Do this early so that we can validate and show errors.
	p.config.Json, err = p.processJsonUserVars()
	if err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Error processing user variables in JSON: %s", err))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	if !p.config.SkipInstall {
                if err := p.installChef(ui, comm); err != nil {
                        return fmt.Errorf("Error installing Chef: %s", err)
                }
        }
	if err := p.moveClient(ui, comm);err != nil {
                return fmt.Errorf("Error moving client.rb: %s",err)
        }


	if err := p.createDir(ui, comm, p.config.StagingDir); err != nil {
		return fmt.Errorf("Error creating staging directory: %s", err)
	}

	if err := p.createHints(ui, comm);err != nil {
		return fmt.Errorf("Error creating ohai hints file and directory: %s", err)
	}

        if err := p.moveValidation(ui, comm);err != nil {
                return fmt.Errorf("Error moving validation.pem: %s", err)
        }

	cookbookPaths := make([]string, 0, len(p.config.CookbookPaths))

	nodeName := ""
	if p.config.NodeName != "" {
		nodeName = fmt.Sprintf("%s",p.config.NodeName)
	}

	serverUrl := ""
	if p.config.ServerUrl != "" {
		serverUrl = fmt.Sprintf("%s",p.config.ServerUrl)
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

	configPath, err := p.createConfig(ui, comm, cookbookPaths, rolesPath, dataBagsPath, nodeName, serverUrl)
	if err != nil {
		return fmt.Errorf("Error creating Chef config file: %s", err)
	}

	jsonPath, err := p.createJson(ui, comm)
	if err != nil {
		return fmt.Errorf("Error creating JSON attributes: %s", err)
	}

	if err := p.executeChef(ui, comm, configPath, jsonPath); err != nil {
                if err2 := p.cleanNode(ui, comm, p.config.NodeName); err2 != nil {
                        return fmt.Errorf("Error cleaning up chef node: %s", err2)
                }

                if err2 := p.cleanClient(ui, comm, p.config.NodeName); err2 != nil {
                        return fmt.Errorf("Error cleaning up chef client: %s", err2)
                }

		return fmt.Errorf("Error executing Chef: %s", err)
	}

	if err := p.cleanNode(ui, comm, p.config.NodeName); err != nil {
		return fmt.Errorf("Error cleaning up chef node: %s", err)
	}

	if err := p.cleanClient(ui, comm, p.config.NodeName); err != nil {
		return fmt.Errorf("Error cleaning up chef client: %s", err)
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

func (p *Provisioner) createConfig(ui packer.Ui, comm packer.Communicator, localCookbooks []string, rolesPath string, dataBagsPath string, nodeName string, serverUrl string) (string, error) {
	ui.Message("Creating configuration file 'client.rb'")

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

	configString, err := p.config.tpl.Process(tpl, &ConfigTemplate{
		CookbookPaths:   strings.Join(cookbook_paths, ","),
		RolesPath:       rolesPath,
		DataBagsPath:    dataBagsPath,
		NodeName:	 nodeName,
		ServerUrl:	 serverUrl,
		HasRolesPath:    rolesPath != "",
		HasDataBagsPath: dataBagsPath != "",
		HasNodeName:	 nodeName != "",
	})
	if err != nil {
		return "", err
	}

	remotePath := filepath.Join(p.config.StagingDir, "client.rb")
	if err := comm.Upload(remotePath, bytes.NewReader([]byte(configString))); err != nil {
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
	remotePath := filepath.Join(p.config.StagingDir, "first-boot.json")
	if err := comm.Upload(remotePath, bytes.NewReader(jsonBytes)); err != nil {
		return "", err
	}

	return remotePath, nil
}

func (p *Provisioner) createDir(ui packer.Ui, comm packer.Communicator, dir string) error {
	ui.Message(fmt.Sprintf("Creating directory: %s", dir))
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("sudo mkdir -p '%s' && sudo chown ubuntu '%s'", dir,dir),
	}

	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Non-zero exit status.")
	}

	return nil
}

func (p *Provisioner) createHints(ui packer.Ui, comm packer.Communicator) error {
	ui.Message(fmt.Sprintf("Creating directory: /etc/chef/ohai/hints"))
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("mkdir -p /etc/chef/ohai/hints; echo '{}' > /etc/chef/ohai/hints/ec2.json"),
	}

	if err := cmd.StartWithUi(comm,ui); err != nil{
		return err
	}
	
	return nil
}

func (p *Provisioner) cleanNode(ui packer.Ui, comm packer.Communicator, node string) error {
        ui.Say("Cleaning up chef node...")
        app := "knife node delete -y "+node
        
	cmd := exec.Command("sh", "-c", app)
        out, err := cmd.Output()

	ui.Message(fmt.Sprintf("%s",out))

	if err != nil{
		return err
	}

	return nil
}

func (p *Provisioner) cleanClient(ui packer.Ui, comm packer.Communicator, node string) error {
	ui.Say("Cleaning up chef client...")
	app := "knife client delete -y "+node

        cmd := exec.Command("sh", "-c", app)
        out, err := cmd.Output()

	ui.Message(fmt.Sprintf("%s",out))

        if err != nil {
                return err
        }

	return nil
}

func (p *Provisioner) removeDir(ui packer.Ui, comm packer.Communicator, dir string) error {
	ui.Message(fmt.Sprintf("Removing directory: %s", dir))
	        cmd := &packer.RemoteCmd{
                Command: fmt.Sprintf("sudo rm -rf %s",dir),
        }

        if err := cmd.StartWithUi(comm,ui); err != nil{
                return err
        }

        return nil
}

func (p *Provisioner) executeChef(ui packer.Ui, comm packer.Communicator, config string, json string) error {
	command, err := p.config.tpl.Process(p.config.ExecuteCommand, &ExecuteTemplate{
		ConfigPath: config,
		JsonPath:   json,
		Sudo:       !p.config.PreventSudo,
	})
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

	command, err := p.config.tpl.Process(p.config.InstallCommand, &InstallChefTemplate{
		Sudo: !p.config.PreventSudo,
	})
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

func (p *Provisioner) moveValidation(ui packer.Ui, comm packer.Communicator) error {
        ui.Message("Moving validation.pem...")

        command, err := p.config.tpl.Process(p.config.ValidationCommand, &InstallChefTemplate{
                Sudo: !p.config.PreventSudo,
        })
        if err != nil {
                return err
        }

        cmd := &packer.RemoteCmd{Command: command}
        if err := cmd.StartWithUi(comm, ui); err != nil {
                return err
        }

        if cmd.ExitStatus != 0 {
                return fmt.Errorf(
                        "Move script exited with non-zero exit status %d", cmd.ExitStatus)
        }

        return nil
}

func (p *Provisioner) moveClient(ui packer.Ui, comm packer.Communicator) error {
        ui.Message("Moving client.rb...")

        command, err := p.config.tpl.Process(p.config.ClientCommand, &InstallChefTemplate{
                Sudo: !p.config.PreventSudo,
        })
        if err != nil {
                return err
        }

        cmd := &packer.RemoteCmd{Command: command}
        if err := cmd.StartWithUi(comm, ui); err != nil {
                return err
        }

        if cmd.ExitStatus != 0 {
                return fmt.Errorf(
                        "Move script exited with non-zero exit status %d", cmd.ExitStatus)
        }

        return nil
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
	for k, v := range p.config.tpl.UserVars {
		originalUserVars[k] = v
	}

	// Make sure we reset them no matter what
	defer func() {
		p.config.tpl.UserVars = originalUserVars
	}()

	// Make the current user variables JSON string safe.
	for k, v := range p.config.tpl.UserVars {
		v = strings.Replace(v, `\`, `\\`, -1)
		v = strings.Replace(v, `"`, `\"`, -1)
		p.config.tpl.UserVars[k] = v
	}

	// Process the bytes with the template processor
	jsonBytesProcessed, err := p.config.tpl.Process(string(jsonBytes), nil)
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
validation_client_name "chef-validator"
{{if .HasNodeName}}
node_name "{{.NodeName}}"
{{end}}
`



//cookbook_path 	[{{.CookbookPaths}}]
//{{if .HasRolesPath}}
//role_path		"{{.RolesPath}}"
//{{end}}
//{{if .HasDataBagsPath}}
//data_bag_path	"{{.DataBagsPath}}"
//{{end}}
//`

