package vagrant

import (
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

func TestStepAdd_Impl(t *testing.T) {
	var raw interface{}
	raw = new(StepAddBox)
	if _, ok := raw.(multistep.Step); !ok {
		t.Fatalf("initialize should be a step")
	}
}

func TestPrepAddArgs(t *testing.T) {
	type testArgs struct {
		Step     StepAddBox
		Expected []string
	}
	addTests := []testArgs{
		{
			Step: StepAddBox{
				SourceBox: "my_source_box.box",
				BoxName:   "AWESOME BOX",
			},
			Expected: []string{"AWESOME BOX", "my_source_box.box"},
		},
		{
			Step: StepAddBox{
				SourceBox: "my_source_box",
				BoxName:   "AWESOME BOX",
			},
			Expected: []string{"my_source_box"},
		},
		{
			Step: StepAddBox{
				BoxVersion:   "eleventyone",
				CACert:       "adfasdf",
				CAPath:       "adfasdf",
				DownloadCert: "adfasdf",
				Clean:        true,
				Force:        true,
				Insecure:     true,
				Provider:     "virtualbox",
				SourceBox:    "bananabox.box",
				BoxName:      "bananas",
			},
			Expected: []string{"bananas", "bananabox.box", "--box-version", "eleventyone", "--cacert", "adfasdf", "--capath", "adfasdf", "--cert", "adfasdf", "--clean", "--force", "--insecure", "--provider", "virtualbox"},
		},
	}
	for _, addTest := range addTests {
		addArgs := addTest.Step.generateAddArgs()
		for i, val := range addTest.Expected {
			if strings.Compare(addArgs[i], val) != 0 {
				t.Fatalf("expected %#v but received %#v", addTest.Expected, addArgs)
			}
		}
	}
}
