// This package implements a provisioner for Packer that uses
// Chef to provision the remote Windows machine, specifically with chef-solo
package windowschefsolo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	ChefEnvironment            string   `mapstructure:"chef_environment"`
	ConfigTemplate             string   `mapstructure:"config_template"`
	CookbookPaths              []string `mapstructure:"cookbook_paths"`
	RolesPath                  string   `mapstructure:"roles_path"`
	DataBagsPath               string   `mapstructure:"data_bags_path"`
	EncryptedDataBagSecretPath string   `mapstructure:"encrypted_data_bag_secret_path"`
	EnvironmentsPath           string   `mapstructure:"environments_path"`
	ExecuteCommand             string   `mapstructure:"execute_command"`
	InstallCommand             string   `mapstructure:"install_command"`
	InstallUrl                 string   `mapstructure:"install_url"`
	RemoteCookbookPaths        []string `mapstructure:"remote_cookbook_paths"`
	Json                       map[string]interface{}
	RunList                    []string `mapstructure:"run_list"`
	SkipInstall                bool     `mapstructure:"skip_install"`
	StagingDir                 string   `mapstructure:"staging_directory"`
	User                       string   `mapstructure:"user"`
	Password                   string   `mapstructure:"password"`

	tpl *packer.ConfigTemplate
}

type Provisioner struct {
	config Config
}

type ConfigTemplate struct {
	CookbookPaths              string
	DataBagsPath               string
	EncryptedDataBagSecretPath string
	RolesPath                  string
	EnvironmentsPath           string
	ChefEnvironment            string

	// Templates don't support boolean statements until Go 1.2. In the
	// mean time, we do this.
	// TODO(mitchellh): Remove when Go 1.2 is released
	HasDataBagsPath               bool
	HasEncryptedDataBagSecretPath bool
	HasRolesPath                  bool
	HasEnvironmentsPath           bool
}

type InstallTemplate struct {
	InstallUrl string
	StagingDir string
}

type ExecuteTemplate struct {
	StagingDir string
	User       string
	Password   string
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
		p.config.ExecuteCommand = DefaultExecuteChefTemplate
	}

	if p.config.InstallCommand == "" {
		p.config.InstallCommand = DefaultInstallChefOmnibusTemplate
	}

	if p.config.InstallUrl == "" {
		p.config.InstallUrl = "http://www.opscode.com/chef/install.msi"
	}

	if p.config.RunList == nil {
		p.config.RunList = make([]string, 0)
	}

	if p.config.StagingDir == "" {
		p.config.StagingDir = "c:/windows/temp/chef"
	}

	if p.config.User == "" {
		p.config.User = "Administrator"
	}

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)

	templates := map[string]*string{
		"config_template":           &p.config.ConfigTemplate,
		"data_bags_path":            &p.config.DataBagsPath,
		"encrypted_data_bag_secret": &p.config.EncryptedDataBagSecretPath,
		"roles_path":                &p.config.RolesPath,
		"staging_dir":               &p.config.StagingDir,
		"environments_path":         &p.config.EnvironmentsPath,
		"chef_environment":          &p.config.ChefEnvironment,
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
	ui.Say("Provisioning with chef-solo")

	if err := p.createDir(ui, comm, p.config.StagingDir); err != nil {
		return fmt.Errorf("Error creating staging directory: %s", err)
	}

	if !p.config.SkipInstall {
		if err := p.installChef(ui, comm); err != nil {
			return fmt.Errorf("Error installing Chef: %s", err)
		}
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

	configPath, err := p.createConfig(ui, comm, cookbookPaths, rolesPath, dataBagsPath, encryptedDataBagSecretPath, environmentsPath, p.config.ChefEnvironment)
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

func (p *Provisioner) uploadFile(ui packer.Ui, comm packer.Communicator, dst string, src string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	return comm.Upload(dst, f)
}

func (p *Provisioner) createConfig(ui packer.Ui, comm packer.Communicator, localCookbooks []string, rolesPath string, dataBagsPath string, encryptedDataBagSecretPath string, environmentsPath string, chefEnvironment string) (string, error) {
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

	configString, err := p.config.tpl.Process(tpl, &ConfigTemplate{
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
	})
	if err != nil {
		return "", err
	}

	remotePath := filepath.Join(p.config.StagingDir, "solo.rb")
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
	remotePath := filepath.Join(p.config.StagingDir, "node.json")
	if err := comm.Upload(remotePath, bytes.NewReader(jsonBytes)); err != nil {
		return "", err
	}

	return remotePath, nil
}

func (p *Provisioner) createDir(ui packer.Ui, comm packer.Communicator, dir string) error {
	ui.Message(fmt.Sprintf("Creating directory: %s", dir))
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("mkdir -p '%s'", dir),
	}

	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Non-zero exit status.")
	}

	return nil
}

func (p *Provisioner) executeChef(ui packer.Ui, comm packer.Communicator, config string, json string) error {
	ui.Message("Executing Chef-Solo...")

	executeTemplate := &ExecuteTemplate{
		User: p.config.User,
		Password: p.config.Password,
		StagingDir: p.config.StagingDir,
	}

	// Create scheduled task xml
	log.Println("Creating cheftask.xml")
	chefTaskXmlString, err := p.config.tpl.Process(ChefTaskXmlTemplate, executeTemplate)
	if err != nil {
		return err
	}

	chefTaskXmlPath := filepath.Join(p.config.StagingDir, "cheftask.xml")
	if err := comm.Upload(chefTaskXmlPath, bytes.NewReader([]byte(chefTaskXmlString))); err != nil {
		return err
	}

	// Create scheduled task powershell script
	log.Println("Creating cheftask_schrun.ps1")
	chefTaskSchrunString, err := p.config.tpl.Process(DefaultChefTaskTemplate, executeTemplate)
	if err != nil {
		return err
	}

	chefTaskSchrunPath := filepath.Join(p.config.StagingDir, "cheftask_schrun.ps1")
	if err := comm.Upload(chefTaskSchrunPath, bytes.NewReader([]byte(chefTaskSchrunString))); err != nil {
		return err
	}

	// Create execution powershell script
	log.Println("Creating cheftask.ps1")
	chefTaskString, err := p.config.tpl.Process(p.config.ExecuteCommand, executeTemplate)
	if err != nil {
		return err
	}

	chefTaskPath := filepath.Join(p.config.StagingDir, "cheftask.ps1")
	if err := comm.Upload(chefTaskPath, bytes.NewReader([]byte(chefTaskString))); err != nil {
		return err
	}

	// Execute the chef run
	log.Println("Executing cheftask.ps1")
	command := fmt.Sprintf("powershell.exe -InputFormat none -File %s", chefTaskPath)
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

	installScript, err := p.config.tpl.Process(p.config.InstallCommand, &InstallTemplate{
		InstallUrl: p.config.InstallUrl,
		StagingDir: p.config.StagingDir,
	})
	if err != nil {
		return err
	}

	installChefPs1Path := filepath.Join(p.config.StagingDir, "install-chef.ps1")
	if err := comm.Upload(installChefPs1Path, bytes.NewReader([]byte(installScript))); err != nil {
		return err
	}

	installCommand := fmt.Sprintf("powershell.exe -InputFormat none -File %s", installChefPs1Path)
	cmd := &packer.RemoteCmd{Command: installCommand}
	if err := cmd.StartWithUi(comm, ui); err != nil {
		return err
	}

	if cmd.ExitStatus != 0 {
		return fmt.Errorf(
			"Install script exited with non-zero exit status %d", cmd.ExitStatus)
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

var DefaultInstallChefOmnibusTemplate = `
$msi_path = join-path "{{.StagingDir}}" "chef.msi"
(New-Object System.Net.WebClient).DownloadFile('{{.InstallUrl}}', $msi_path)
Start-Process 'msiexec' -ArgumentList "/qb /i $msi_path" -NoNewWindow -Wait
`

var DefaultExecuteChefTemplate = `
$chef_task_running_file = join-path "{{.StagingDir}}" "cheftask.running"
$chef_task_xml_file = join-path "{{.StagingDir}}" "cheftask.xml"
$chef_stdout_file = join-path "{{.StagingDir}}" "cheftask.log"
$chef_exit_code_file = join-path "{{.StagingDir}}" "cheftask.exitcode"

# kill the task so we can recreate it
schtasks /delete /tn "chef-solo" /f 2>&1 | out-null

# Ensure the chef task running file doesn't exist from a previous failure
if (Test-Path $chef_task_running_file) {
  del $chef_task_running_file
}

# schedule the task to run once in the far distant future
schtasks /create /tn 'chef-solo' /xml $chef_task_xml_file /ru '{{.User}}' /rp '{{.Password}}' | Out-Null

# start the scheduled task right now
schtasks /run /tn "chef-solo" | Out-Null

# wait for run_chef.ps1 to start or timeout after 1 minute
$timeoutSeconds = 60
$elapsedSeconds = 0
while ( (!(Test-Path $chef_task_running_file)) -and ($elapsedSeconds -lt $timeoutSeconds) ) {
  Start-Sleep -s 1
  $elapsedSeconds++
}

if ($elapsedSeconds -ge $timeoutSeconds) {
  Write-Error "Timed out waiting for chef scheduled task to start"
  exit -2
}

# read the entire file, but only write out new lines we haven't seen before
$numLinesRead = 0
$success = $TRUE
while (Test-Path $chef_task_running_file) {
  Start-Sleep -m 100
  
  if (Test-Path $chef_stdout_file) {
    $text = (get-content $chef_stdout_file)
    $numLines = ($text | Measure-Object -line).lines    
    $numLinesToRead = $numLines - $numLinesRead
    
    if ($numLinesToRead -gt 0) {
      $text | select -first $numLinesToRead -skip $numLinesRead | ForEach {
        Write-Host "$_"
      }
      $numLinesRead += $numLinesToRead
    }
  }
}

exit Get-Content $chef_exit_code_file
`

var ChefTaskXmlTemplate = `<?xml version="1.0" encoding="UTF-16"?>
<Task version="1.2" xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
  <RegistrationInfo>
    <Date>2013-06-22T12:00:00</Date>
    <Author>{{.User}}</Author>
  </RegistrationInfo>
  <Triggers/>
  <Principals>
    <Principal id="Author">
      <UserId>{{.User}}</UserId>
      <LogonType>Password</LogonType>
      <RunLevel>HighestAvailable</RunLevel>
    </Principal>
  </Principals>
  <Settings>
    <MultipleInstancesPolicy>IgnoreNew</MultipleInstancesPolicy>
    <DisallowStartIfOnBatteries>false</DisallowStartIfOnBatteries>
    <StopIfGoingOnBatteries>false</StopIfGoingOnBatteries>
    <AllowHardTerminate>true</AllowHardTerminate>
    <StartWhenAvailable>false</StartWhenAvailable>
    <RunOnlyIfNetworkAvailable>false</RunOnlyIfNetworkAvailable>
    <IdleSettings>
      <StopOnIdleEnd>true</StopOnIdleEnd>
      <RestartOnIdle>false</RestartOnIdle>
    </IdleSettings>
    <AllowStartOnDemand>true</AllowStartOnDemand>
    <Enabled>true</Enabled>
    <Hidden>false</Hidden>
    <RunOnlyIfIdle>false</RunOnlyIfIdle>
    <WakeToRun>false</WakeToRun>
    <ExecutionTimeLimit>PT2H</ExecutionTimeLimit>
    <Priority>4</Priority>
  </Settings>
  <Actions Context="Author">
    <Exec>
      <Command>powershell</Command>
      <Arguments>-file {{.StagingDir}}/cheftask_schrun.ps1</Arguments>
    </Exec>
  </Actions>
</Task>
`

var DefaultChefTaskTemplate = `
$exitCode = -1
Set-ExecutionPolicy Unrestricted -force;

$chef_task_running_file = join-path "{{.StagingDir}}" "cheftask.running"
$chef_task_xml_file = join-path "{{.StagingDir}}" "cheftask.xml"
$chef_stdout_file = join-path "{{.StagingDir}}" "cheftask.log"
$chef_stderr_file = join-path "{{.StagingDir}}" "cheftask.err.log"
$chef_exit_code_file = join-path "{{.StagingDir}}" "cheftask.exitcode"

$solo_rb_file = join-path "{{.StagingDir}}" "solo.rb"
$node_json_file = join-path "{{.StagingDir}}" "node.json"

Try
{
  "running" | Out-File $chef_task_running_file
  $process = (Start-Process "c:\opscode\chef\bin\chef-solo.bat" -ArgumentList "--no-color -c $solo_rb_file -j $node_json_file -l debug" -NoNewWindow -PassThru -Wait -RedirectStandardOutput $chef_stdout_file -RedirectStandardError $chef_stderr_file)
  $exitCode = $process.ExitCode
}
Finally
{
  $exitCode | Out-File $chef_exit_code_file
  if (Test-Path $chef_task_running_file) {
    del $chef_task_running_file
  }
}

exit $exitCode
`
