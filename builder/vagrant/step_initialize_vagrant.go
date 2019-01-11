package vagrant

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepInitializeVagrant struct {
	BoxName      string
	BoxVersion   string
	Minimal      bool
	Template     string
	SourceBox    string
	OutputDir    string
	SyncedFolder string
}

var DEFAULT_TEMPLATE = `Vagrant.configure("2") do |config|
  config.vm.box = "{{.BoxName}}"
  {{ if ne .SyncedFolder "" -}}
  		config.vm.synced_folder "{{.SyncedFolder}}", "/vagrant"
  {{- else -}}
  		config.vm.synced_folder ".", "/vagrant", disabled: true
  {{- end}}
end`

type VagrantfileOptions struct {
	SyncedFolder string
	BoxName      string
}

func (s *StepInitializeVagrant) createInitializeCommand() (string, error) {
	tplPath := filepath.Join(s.OutputDir, "packer-vagrantfile-template.erb")
	templateFile, err := os.Create(tplPath)
	templateFile.Chmod(0777)
	if err != nil {
		retErr := fmt.Errorf("Error creating vagrantfile %s", err.Error())
		return "", retErr
	}

	var tpl *template.Template
	if s.Template == "" {
		// Generate vagrantfile template based on our default
		tpl = template.Must(template.New("VagrantTpl").Parse(DEFAULT_TEMPLATE))
	} else {
		// Read in the template from provided file.
		tpl, err = template.ParseFiles(s.Template)
		if err != nil {
			return "", err
		}
	}

	opts := &VagrantfileOptions{
		SyncedFolder: s.SyncedFolder,
		BoxName:      s.SourceBox,
	}

	err = tpl.Execute(templateFile, opts)
	if err != nil {
		return "", err
	}

	abspath, err := filepath.Abs(tplPath)
	if err != nil {
		return "", err
	}

	return abspath, nil
}
func (s *StepInitializeVagrant) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(VagrantDriver)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Initializing Vagrant in build directory...")

	// Prepare arguments
	initArgs := []string{}

	if s.BoxName != "" {
		initArgs = append(initArgs, s.BoxName)
	}

	initArgs = append(initArgs, s.SourceBox)

	if s.BoxVersion != "" {
		initArgs = append(initArgs, "--box-version", s.BoxVersion)
	}

	if s.Minimal {
		initArgs = append(initArgs, "-m")
	}

	tplPath, err := s.createInitializeCommand()
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	initArgs = append(initArgs, "--template", tplPath)

	os.Chdir(s.OutputDir)
	// Call vagrant using prepared arguments
	err = driver.Init(initArgs)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepInitializeVagrant) Cleanup(state multistep.StateBag) {
}
