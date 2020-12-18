package vagrant

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepCreateVagrantfile struct {
	Template     string
	OutputDir    string
	SyncedFolder string
	GlobalID     string
	SourceBox    string
	BoxName      string
	InsertKey    bool
}

var DEFAULT_TEMPLATE = `Vagrant.configure("2") do |config|
  config.vm.define "source", autostart: false do |source|
	source.vm.box = "{{.SourceBox}}"
	config.ssh.insert_key = {{.InsertKey}}
  end
  config.vm.define "output" do |output|
	output.vm.box = "{{.BoxName}}"
	output.vm.box_url = "file://package.box"
	config.ssh.insert_key = {{.InsertKey}}
  end
  {{ if ne .SyncedFolder "" -}}
  		config.vm.synced_folder "{{.SyncedFolder}}", "/vagrant"
  {{- else -}}
  		config.vm.synced_folder ".", "/vagrant", disabled: true
  {{- end}}
end`

type VagrantfileOptions struct {
	SyncedFolder string
	SourceBox    string
	BoxName      string
	InsertKey    bool
}

func (s *StepCreateVagrantfile) createVagrantfile() (string, error) {
	tplPath := filepath.Join(s.OutputDir, "Vagrantfile")
	templateFile, err := os.Create(tplPath)
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
		BoxName:      s.BoxName,
		SourceBox:    s.SourceBox,
		InsertKey:    s.InsertKey,
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

func (s *StepCreateVagrantfile) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	// Skip the initialize step if we're trying to launch from a global ID.
	if s.GlobalID != "" {
		ui.Say("Using a global-id; skipping Vagrant init in this directory...")
		return multistep.ActionContinue
	}

	ui.Say("Creating a Vagrantfile in the build directory...")
	vagrantfilePath, err := s.createVagrantfile()
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	log.Printf("Created vagrantfile at %s", vagrantfilePath)

	return multistep.ActionContinue
}

func (s *StepCreateVagrantfile) Cleanup(state multistep.StateBag) {
}
