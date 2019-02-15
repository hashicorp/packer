package vagrant

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepCreateVagrantfile_Impl(t *testing.T) {
	var raw interface{}
	raw = new(StepCreateVagrantfile)
	if _, ok := raw.(multistep.Step); !ok {
		t.Fatalf("initialize should be a step")
	}
}

func TestCreateFile(t *testing.T) {
	testy := StepCreateVagrantfile{
		OutputDir: "./",
		SourceBox: "bananas",
	}
	templatePath, err := testy.createVagrantfile()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer os.Remove(templatePath)
	contents, err := ioutil.ReadFile(templatePath)
	actual := string(contents)
	expected := `Vagrant.configure("2") do |config|
  config.vm.box = "bananas"
  config.vm.synced_folder ".", "/vagrant", disabled: true
end`
	if ok := strings.Compare(actual, expected); ok != 0 {
		t.Fatalf("EXPECTED: \n%s\n\n RECEIVED: \n%s\n\n", expected, actual)
	}
}

func TestCreateFile_customSync(t *testing.T) {
	testy := StepCreateVagrantfile{
		OutputDir:    "./",
		SyncedFolder: "myfolder/foldertimes",
	}
	templatePath, err := testy.createVagrantfile()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer os.Remove(templatePath)
	contents, err := ioutil.ReadFile(templatePath)
	actual := string(contents)
	expected := `Vagrant.configure("2") do |config|
  config.vm.box = ""
  config.vm.synced_folder "myfolder/foldertimes", "/vagrant"
end`
	if ok := strings.Compare(actual, expected); ok != 0 {
		t.Fatalf("EXPECTED: \n%s\n\n RECEIVED: \n%s\n\n", expected, actual)
	}
}
