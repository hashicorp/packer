// This package implements a provisioner for Packer that executes
// Puppet within the remote machine
package puppet

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/iochan"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/packer"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	RemoteStagingPath   = "/tmp/provision/puppet"
	DefaultModulePath   = "modules"
	DefaultManifestPath = "manifests"
	DefaultManifestFile = "site.pp"
)

var Ui packer.Ui

type config struct {
	// An array of local paths of modules to upload.
	ModulePath string `mapstructure:"module_path"`

	// Path to the manifests
	ManifestPath string `mapstructure:"manifest_path"`

	// Manifest file
	ManifestFile string `mapstructure:"manifest_file"`

	// Option to avoid sudo use when executing commands. Defaults to false.
	PreventSudo bool `mapstructure:"prevent_sudo"`
}

type Provisioner struct {
	config config
}

type ExecuteManifestTemplate struct {
	Sudo       bool
	Modulepath string
	Manifest   string
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	errs := make([]error, 0)
	for _, raw := range raws {
		if err := mapstructure.Decode(raw, &p.config); err != nil {
			return err
		}
	}

	if p.config.ModulePath == "" {
		p.config.ModulePath = DefaultModulePath
	}

	if p.config.ManifestPath == "" {
		p.config.ManifestPath = DefaultManifestPath
	}

	if p.config.ManifestFile == "" {
		p.config.ManifestFile = DefaultManifestFile
	}

	if p.config.ModulePath != "" {
		pFileInfo, err := os.Stat(p.config.ModulePath)

		if err != nil || !pFileInfo.IsDir() {
			errs = append(errs, fmt.Errorf("Bad module path '%s': %s", p.config.ModulePath, err))
		}
	}

	if p.config.ManifestPath != "" {
		pFileInfo, err := os.Stat(p.config.ManifestPath)

		if err != nil || !pFileInfo.IsDir() {
			errs = append(errs, fmt.Errorf("Bad manifest path '%s': %s", p.config.ManifestPath, err))
		}
	}

	if p.config.ManifestFile != "" {
		path := filepath.Join(p.config.ManifestPath, p.config.ManifestFile)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("No manifest file '%s': %s", path, err))
		}
	}

	if len(errs) > 0 {
		return &packer.MultiError{errs}
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	var err error
	Ui = ui

	err = CreateRemoteDirectory(RemoteStagingPath, comm)
	if err != nil {
		return fmt.Errorf("Error creating remote staging directory: %s", err)
	}

	// Upload all modules
	ui.Say(fmt.Sprintf("Copying module path: %s", p.config.ModulePath))
	err = UploadLocalDirectory(p.config.ModulePath, comm)
	if err != nil {
		return fmt.Errorf("Error uploading modules: %s", err)
	}

	// Upload manifests
	ui.Say(fmt.Sprintf("Copying manifests: %s", p.config.ManifestPath))
	err = UploadLocalDirectory(p.config.ManifestPath, comm)
	if err != nil {
		return fmt.Errorf("Error uploading manifests: %s", err)
	}

	// Execute Puppet
	ui.Say("Beginning Puppet run")

	// Compile the command
	var command bytes.Buffer
	mpath := filepath.Join(RemoteStagingPath, p.config.ManifestPath)
	manifest := filepath.Join(mpath, p.config.ManifestFile)
	modulepath := filepath.Join(RemoteStagingPath, p.config.ModulePath)
	t := template.Must(template.New("puppet-run").Parse("{{if .Sudo}}sudo {{end}}puppet apply --verbose --modulepath={{.Modulepath}} {{.Manifest}}"))
	t.Execute(&command, &ExecuteManifestTemplate{!p.config.PreventSudo, modulepath, manifest})

	err = executeCommand(command.String(), comm)
	if err != nil {
		return fmt.Errorf("Error running Puppet: %s", err)
	}

	return nil
}

func (p *Provisioner) Cancel() {
	// Just hard quit. It isn't a big deal if what we're doing keeps
	// running on the other side.
	os.Exit(0)
}

func UploadLocalDirectory(localDir string, comm packer.Communicator) (err error) {
	visitPath := func(path string, f os.FileInfo, err error) (err2 error) {
		var remotePath = RemoteStagingPath + "/" + path
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
		return fmt.Errorf("Error uploading modules %s: %s", localDir, err)
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
			log.Printf("Puppet provisioner exited with status %d", exitStatus)

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
