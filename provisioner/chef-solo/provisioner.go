// This package implements a provisioner for Packer that executes
// shell scripts within the remote machine.
package chefSolo

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/mitchellh/iochan"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/packer"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"
	"path/filepath"
)

const RemoteStagingPath = "/tmp/provision/chef-solo"
const RemoteFileCachePath = "/tmp/provision/chef-solo"
const RemoteCookbookPath = "/tmp/provision/chef-solo/cookbooks"
const DefaultCookbookPath = "cookbooks"

var Ui packer.Ui

type config struct {
	// An array of local paths of cookbooks to upload.
	CookbookPaths []string `mapstructure:"cookbook_paths"`

	// The local path of the cookbooks to upload.
	CookbookPath string `mapstructure:"cookbook_path"`

	// An array of recipes to run.
	RunList []string `mapstructure:"run_list"`

	// An array of environment variables that will be injected before
	// your command(s) are executed.
	JsonFile string `mapstructure:"json_file"`
}

type Provisioner struct {
	config config
}

type ExecuteRecipeTemplate struct {
	SoloRbPath string
	JsonPath string
	RunList string
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	errs := make([]error, 0)
	for _, raw := range raws {
		if err := mapstructure.Decode(raw, &p.config); err != nil {
			return err
		}
	}
	
	if p.config.CookbookPaths == nil {
		p.config.CookbookPaths = make([]string, 0)
	}

	if p.config.CookbookPath == "" {
		p.config.CookbookPath = DefaultCookbookPath
	}
	
	if len(p.config.CookbookPaths) > 0 && p.config.CookbookPath != "" {
		errs = append(errs, errors.New("Only one of cookbooks or cookbook can be specified."))
	}
	
	if len(p.config.CookbookPaths) == 0 {
		p.config.CookbookPaths = append(p.config.CookbookPaths, p.config.CookbookPath)
	}

	if p.config.RunList == nil {
		p.config.RunList = make([]string, 0)
	}

	if p.config.JsonFile != "" {
		if _, err := os.Stat(p.config.JsonFile); err != nil {
			errs = append(errs, fmt.Errorf("Bad JSON attributes file '%s': %s", p.config.JsonFile, err))
		}
	}

	for _, path := range p.config.CookbookPaths {
		pFileInfo, err := os.Stat(path)
		
		if err != nil || !pFileInfo.IsDir() {
			errs = append(errs, fmt.Errorf("Bad cookbook path '%s': %s", path, err))
		}
	}

	if len(errs) > 0 {
		return &packer.MultiError{errs}
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	cookbookPaths := make([]string, len(p.config.CookbookPaths))
	copy(cookbookPaths, p.config.CookbookPaths)
	
	Ui = ui
	
	// Generic setup for Chef runs
	err := InstallChefSolo(comm)
	if err != nil {
		return fmt.Errorf("Error installing Chef Solo: %s", err)
	}
	
	err = CreateRemoteDirectory(RemoteCookbookPath, comm)
	if err != nil {
		return fmt.Errorf("Error creating remote staging directory: %s", err)
	}
	
	soloRbPath, err := CreateSoloRb(p.config.CookbookPaths, comm)
	if err != nil {
		return fmt.Errorf("Error creating Chef Solo configuration file: %s", err)
	}
	
	jsonPath, err := CreateAttributesJson(p.config.JsonFile, comm)
	if err != nil {
		return fmt.Errorf("Error uploading JSON attributes file: %s", err)
	}
	
	// Upload all cookbooks
	for _, path := range cookbookPaths {
		ui.Say(fmt.Sprintf("Copying cookbook path: %s", path))
		err = UploadLocalDirectory(path, comm)
		if err != nil {
			return fmt.Errorf("Error uploading cookbooks: %s", err)
		}
	}
	
	// Execute requested recipes
	for _, recipe := range p.config.RunList {
		ui.Say(fmt.Sprintf("chef-solo running recipe: %s", recipe))
		// Compile the command
		var command bytes.Buffer
		t := template.Must(template.New("chef-run").Parse("sudo chef-solo --no-color -c {{.SoloRbPath}} -j {{.JsonPath}} -o {{.RunList}}"))
		t.Execute(&command, &ExecuteRecipeTemplate{soloRbPath, jsonPath, recipe})
		
		err = executeCommand(command.String(), comm)
		if err != nil {
			return fmt.Errorf("Error running recipe %s: %s", recipe, err)
		}
	}

	return nil
}

func UploadLocalDirectory(localDir string, comm packer.Communicator) (err error) {
	visitPath := func (path string, f os.FileInfo, err error) (err2 error) {
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
	
	err = filepath.Walk(localDir, visitPath)
	if err != nil {
		return fmt.Errorf("Error uploading cookbook %s: %s", localDir, err)
	}
	
	return nil
}

func CreateRemoteDirectory(path string, comm packer.Communicator) (err error) {
	//Ui.Say(fmt.Sprintf("Creating directory: %s", path))
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

func CreateSoloRb(cookbookPaths []string, comm packer.Communicator) (str string, err error) {
	Ui.Say(fmt.Sprintf("Creating Chef configuration file..."))
	
	remotePath := RemoteStagingPath + "/solo.rb"
	tf, err := ioutil.TempFile("", "packer-chef-solo-rb")
	if err != nil {
		return "", fmt.Errorf("Error preparing Chef solo.rb: %s", err)
	}
	
	// Write our contents to it
	writer := bufio.NewWriter(tf)
	
	// Messy, messy...
	cbPathsCat := "\"" + RemoteCookbookPath + "/" + strings.Join(cookbookPaths, "\",\"" + RemoteCookbookPath + "/") + "\""
	contents := "file_cache_path \"" + RemoteFileCachePath + "\"\ncookbook_path [" + cbPathsCat + "]\n"
	
	if _, err := writer.WriteString(contents); err != nil {
		return "", fmt.Errorf("Error preparing solo.rb: %s", err)
	}

	if err := writer.Flush(); err != nil {
		return "", fmt.Errorf("Error preparing solo.rb: %s", err)
	}
	
	name := tf.Name()
	tf.Close()
	f, err := os.Open(name)
	comm.Upload(remotePath, f)
	
	defer os.Remove(name)
	
	// Upload the Chef Solo configuration file to the cookbook directory.
	
	if err != nil {
		return "", fmt.Errorf("Error uploading Chef Solo configuration file: %s", err)
	}
	//executeCommand("sudo cat " + remotePath, comm)
	return remotePath, nil
}

func CreateAttributesJson(jsonFile string, comm packer.Communicator) (str string, err error) {
	Ui.Say(fmt.Sprintf("Uploading Chef attributes file %s", jsonFile))
	remotePath := RemoteStagingPath + "/node.json"
	
	// Create an empty JSON file if none given
	if jsonFile == "" {
		tf, err := ioutil.TempFile("", "packer-chef-solo-json")
		if err != nil {
			return "", fmt.Errorf("Error preparing Chef attributes file: %s", err)
		}
		defer os.Remove(tf.Name())

		// Write our contents to it
		writer := bufio.NewWriter(tf)
		if _, err := writer.WriteString("{}"); err != nil {
			return "", fmt.Errorf("Error preparing Chef attributes file: %s", err)
		}

		if err := writer.Flush(); err != nil {
			return "", fmt.Errorf("Error preparing Chef attributes file: %s", err)
		}
		
		jsonFile = tf.Name()
		tf.Close()
	}
	
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

func InstallChefSolo(comm packer.Communicator) (err error) {
	Ui.Say(fmt.Sprintf("Installing Chef Solo"))
	var installCommand = "curl -L https://www.opscode.com/chef/install.sh | sudo bash"
	err = executeCommand(installCommand, comm)
	if err != nil {
	  return fmt.Errorf("Unable to install Chef Solo: %d", err)
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
	
	//Ui.Say(fmt.Sprintf("Executing command: %s", cmd.Command))
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
