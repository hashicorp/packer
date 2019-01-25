package vagrant

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepInitialize_Impl(t *testing.T) {
	var raw interface{}
	raw = new(StepInitializeVagrant)
	if _, ok := raw.(multistep.Step); !ok {
		t.Fatalf("initialize should be a step")
	}
}

func TestCreateFile(t *testing.T) {
	testy := StepInitializeVagrant{
		OutputDir: "./",
		SourceBox: "bananas",
	}
	templatePath, err := testy.createInitializeCommand()
	if err != nil {
		t.Fatalf(err.Error())
	}
	contents, err := ioutil.ReadFile(templatePath)
	actual := string(contents)
	expected := `Vagrant.configure("2") do |config|
  config.vm.box = "bananas"
  config.vm.synced_folder ".", "/vagrant", disabled: true
end`
	if ok := strings.Compare(actual, expected); ok != 0 {
		t.Fatalf("EXPECTED: \n%s\n\n RECEIVED: \n%s\n\n", expected, actual)
	}
	os.Remove(templatePath)
}

func TestCreateFile_customSync(t *testing.T) {
	testy := StepInitializeVagrant{
		OutputDir:    "./",
		SyncedFolder: "myfolder/foldertimes",
	}
	templatePath, err := testy.createInitializeCommand()
	defer os.Remove(templatePath)
	if err != nil {
		t.Fatalf(err.Error())
	}
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

func TestPrepInitArgs(t *testing.T) {
	type testArgs struct {
		Step     StepInitializeVagrant
		Expected []string
	}
	initTests := []testArgs{
		{
			Step: StepInitializeVagrant{
				SourceBox: "my_source_box.box",
			},
			Expected: []string{"my_source_box.box", "--template"},
		},
		{
			Step: StepInitializeVagrant{
				SourceBox: "my_source_box",
				BoxName:   "My Box",
			},
			Expected: []string{"My Box", "my_source_box", "--template"},
		},
		{
			Step: StepInitializeVagrant{
				SourceBox:  "my_source_box",
				BoxName:    "My Box",
				BoxVersion: "42",
			},
			Expected: []string{"My Box", "my_source_box", "--box-version", "42", "--template"},
		},
		{
			Step: StepInitializeVagrant{
				SourceBox: "my_source_box",
				BoxName:   "My Box",
				Minimal:   true,
			},
			Expected: []string{"My Box", "my_source_box", "-m", "--template"},
		},
	}
	for _, initTest := range initTests {
		initArgs, err := initTest.Step.prepInitArgs()
		defer os.Remove(initArgs[len(initArgs)-1])
		if err != nil {
			t.Fatalf(err.Error())
		}
		for i, val := range initTest.Expected {
			if strings.Compare(initArgs[i], val) != 0 {
				t.Fatalf("expected %#v but received %#v", initTest.Expected, initArgs[:len(initArgs)-1])
			}
		}
	}
}
