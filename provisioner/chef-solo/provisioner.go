// This package implements a provisioner for Packer that uses
// Chef to provision the remote machine, specifically with chef-solo (that is,
// without a Chef server).
package chefsolo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/iochan"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	RemoteStagingPath    = "/tmp/provision/chef-solo"
	RemoteFileCachePath  = "/tmp/provision/chef-solo"
	RemoteCookbookPath   = "/tmp/provision/chef-solo/cookbooks"
	DefaultCookbooksPath = "cookbooks"
)

var Ui packer.Ui

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	CookbooksPaths []string `mapstructure:"cookbooks_paths"`
	InstallCommand string   `mapstructure:"install_command"`
	Recipes        []string
	Json           map[string]interface{}
	PreventSudo    bool `mapstructure:"prevent_sudo"`
	SkipInstall    bool `mapstructure:"skip_install"`

	tpl *packer.ConfigTemplate
}

type Provisioner struct {
	config Config
}

type ExecuteRecipeTemplate struct {
	SoloRbPath string
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

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)

	if p.config.CookbooksPaths == nil {
		p.config.CookbooksPaths = []string{DefaultCookbooksPath}
	}

	if p.config.InstallCommand == "" {
		p.config.InstallCommand = "curl -L https://www.opscode.com/chef/install.sh | {{if .Sudo}}sudo {{end}}bash"
	}

	if p.config.Recipes == nil {
		p.config.Recipes = make([]string, 0)
	}

	if p.config.Json != nil {
		if _, err := json.Marshal(p.config.Json); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Bad JSON: %s", err))
		}
	} else {
		p.config.Json = make(map[string]interface{})
	}

	for _, path := range p.config.CookbooksPaths {
		pFileInfo, err := os.Stat(path)

		if err != nil || !pFileInfo.IsDir() {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Bad cookbook path '%s': %s", path, err))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	var err error
	Ui = ui

	if !p.config.SkipInstall {
		if err := p.installChef(ui, comm); err != nil {
			return fmt.Errorf("Error installing Chef: %s", err)
		}
	}

	err = CreateRemoteDirectory(RemoteCookbookPath, comm)
	if err != nil {
		return fmt.Errorf("Error creating remote staging directory: %s", err)
	}

	soloRbPath, err := CreateSoloRb(p.config.CookbooksPaths, comm)
	if err != nil {
		return fmt.Errorf("Error creating Chef Solo configuration file: %s", err)
	}

	jsonPath, err := CreateAttributesJson(p.config.Json, p.config.Recipes, comm)
	if err != nil {
		return fmt.Errorf("Error uploading JSON attributes file: %s", err)
	}

	// Upload all cookbooks
	for _, path := range p.config.CookbooksPaths {
		ui.Say(fmt.Sprintf("Copying cookbook path: %s", path))
		err = UploadLocalDirectory(path, comm)
		if err != nil {
			return fmt.Errorf("Error uploading cookbooks: %s", err)
		}
	}

	// Execute requested recipes
	ui.Say("Beginning Chef Solo run")

	// Compile the command
	var command bytes.Buffer
	t := template.Must(template.New("chef-run").Parse("{{if .Sudo}}sudo {{end}}chef-solo --no-color -c {{.SoloRbPath}} -j {{.JsonPath}}"))
	t.Execute(&command, &ExecuteRecipeTemplate{soloRbPath, jsonPath, !p.config.PreventSudo})

	err = executeCommand(command.String(), comm)
	if err != nil {
		return fmt.Errorf("Error running Chef Solo: %s", err)
	}

	return nil
}

func UploadLocalDirectory(localDir string, comm packer.Communicator) (err error) {
	visitPath := func(path string, f os.FileInfo, err error) (err2 error) {
		var remotePath = RemoteCookbookPath + "/" + path
		if f.IsDir() {
			// Make remote directory
			err = CreateRemoteDirectory(remotePath, comm)
			if err != nil {
				return err
			}
		} else {
			// Upload file to existing directory
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("Error opening file: %s", err)
			}

			err = comm.Upload(remotePath, file)
			if err != nil {
				return fmt.Errorf("Error uploading file: %s", err)
			}
		}
		return
	}

	log.Printf("Uploading directory %s", localDir)
	err = filepath.Walk(localDir, visitPath)
	if err != nil {
		return fmt.Errorf("Error uploading cookbook %s: %s", localDir, err)
	}

	return nil
}

func CreateRemoteDirectory(path string, comm packer.Communicator) (err error) {
	log.Printf("Creating remote directory: %s ", path)

	var copyCommand = []string{"mkdir -p", path}

	var cmd packer.RemoteCmd
	cmd.Command = strings.Join(copyCommand, " ")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// Start the command
	if err := comm.Start(&cmd); err != nil {
		return fmt.Errorf("Unable to create remote directory %s: %d", path, err)
	}

	// Wait for it to complete
	cmd.Wait()

	return
}

func CreateSoloRb(cookbooksPaths []string, comm packer.Communicator) (str string, err error) {
	Ui.Say("Creating Chef configuration file...")

	remotePath := RemoteStagingPath + "/solo.rb"

	tf, err := ioutil.TempFile("", "packer-chef-solo-rb")
	if err != nil {
		return "", fmt.Errorf("Error preparing Chef solo.rb: %s", err)
	}

	// Write our contents to it
	writer := bufio.NewWriter(tf)

	var cookbooksPathsFull = make([]string, len(cookbooksPaths))
	for i, path := range cookbooksPaths {
		cookbooksPathsFull[i] = "\"" + RemoteCookbookPath + "/" + path + "\""
	}

	var contents bytes.Buffer
	var soloRbText = `
	file_cache_path "{{.FileCachePath}}"
	cookbook_path   [{{.CookbookPath}}]
`

	t := template.Must(template.New("soloRb").Parse(soloRbText))
	t.Execute(&contents, map[string]string{
		"FileCachePath": RemoteFileCachePath,
		"CookbookPath":  strings.Join(cookbooksPathsFull, ","),
	})

	if _, err := writer.WriteString(contents.String()); err != nil {
		return "", fmt.Errorf("Error preparing solo.rb: %s", err)
	}

	if err := writer.Flush(); err != nil {
		return "", fmt.Errorf("Error preparing solo.rb: %s", err)
	}

	name := tf.Name()
	tf.Close()
	f, err := os.Open(name)
	defer os.Remove(name)

	log.Printf("Chef configuration file contents: %s", contents)

	// Upload the Chef Solo configuration file to the cookbook directory.
	log.Printf("Uploading chef configuration file to %s", remotePath)
	err = comm.Upload(remotePath, f)
	if err != nil {
		return "", fmt.Errorf("Error uploading Chef Solo configuration file: %s", err)
	}

	return remotePath, nil
}

func CreateAttributesJson(jsonAttrs map[string]interface{}, recipes []string, comm packer.Communicator) (str string, err error) {
	Ui.Say("Creating and uploading Chef attributes file")
	remotePath := RemoteStagingPath + "/node.json"

	var formattedRecipes []string
	for _, value := range recipes {
		formattedRecipes = append(formattedRecipes, "recipe["+value+"]")
	}

	// Add Recipes to JSON
	if len(formattedRecipes) > 0 {
		log.Printf("Overriding node run list: %s", strings.Join(formattedRecipes, ", "))
		jsonAttrs["run_list"] = formattedRecipes
	}

	// Convert to JSON string
	jsonString, err := json.MarshalIndent(jsonAttrs, "", "  ")
	if err != nil {
		return "", fmt.Errorf("Error parsing JSON attributes: %s", err)
	}

	tf, err := ioutil.TempFile("", "packer-chef-solo-json")
	if err != nil {
		return "", fmt.Errorf("Error preparing Chef attributes file: %s", err)
	}
	defer os.Remove(tf.Name())

	// Write our contents to it
	writer := bufio.NewWriter(tf)
	if _, err := writer.WriteString(string(jsonString)); err != nil {
		return "", fmt.Errorf("Error preparing Chef attributes file: %s", err)
	}

	if err := writer.Flush(); err != nil {
		return "", fmt.Errorf("Error preparing Chef attributes file: %s", err)
	}

	jsonFile := tf.Name()
	tf.Close()

	log.Printf("Opening %s for reading", jsonFile)
	f, err := os.Open(jsonFile)
	if err != nil {
		return "", fmt.Errorf("Error opening JSON attributes file: %s", err)
	}

	log.Printf("Uploading %s => %s", jsonFile, remotePath)
	err = comm.Upload(remotePath, f)
	if err != nil {
		return "", fmt.Errorf("Error uploading JSON attributes file: %s", err)
	}

	return remotePath, nil
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

func executeCommand(command string, comm packer.Communicator) (err error) {
	// Setup the remote command
	stdout_r, stdout_w := io.Pipe()
	stderr_r, stderr_w := io.Pipe()

	var cmd packer.RemoteCmd
	cmd.Command = command
	cmd.Stdout = stdout_w
	cmd.Stderr = stderr_w

	log.Printf("Executing command: %s", cmd.Command)
	err = comm.Start(&cmd)
	if err != nil {
		return fmt.Errorf("Failed executing command: %s", err)
	}

	exitChan := make(chan int, 1)
	stdoutChan := iochan.DelimReader(stdout_r, '\n')
	stderrChan := iochan.DelimReader(stderr_r, '\n')

	go func() {
		defer stdout_w.Close()
		defer stderr_w.Close()

		cmd.Wait()
		exitChan <- cmd.ExitStatus
	}()

OutputLoop:
	for {
		select {
		case output := <-stderrChan:
			Ui.Message(strings.TrimSpace(output))
		case output := <-stdoutChan:
			Ui.Message(strings.TrimSpace(output))
		case exitStatus := <-exitChan:
			log.Printf("Chef Solo provisioner exited with status %d", exitStatus)

			if exitStatus != 0 {
				return fmt.Errorf("Command exited with non-zero exit status: %d", exitStatus)
			}

			break OutputLoop
		}
	}

	// Make sure we finish off stdout/stderr because we may have gotten
	// a message from the exit channel first.
	for output := range stdoutChan {
		Ui.Message(output)
	}

	for output := range stderrChan {
		Ui.Message(output)
	}

	return nil
}
