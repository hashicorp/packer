package iso

import (
	"fmt"
	"github.com/mitchellh/multistep"
	vboxcommon "github.com/mitchellh/packer/builder/virtualbox/common"
	"github.com/mitchellh/packer/packer"
	"os/exec"
	"log"
	"regexp"
)


func extpackStatus() string {

//	Get extpacks info
	extpacksCmd := exec.Command("vboxmanage", "list", "extpacks")
	extpacksOut, err := extpacksCmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	extpacksOutString := string(extpacksOut)
	fmt.Printf("%s\n", extpacksOutString)

// 	Regexp to confirm if extpack is installed (Search for name)
	reExtpacksInstalled, err := regexp.Compile("Oracle VM VirtualBox Extension Pack")
	if err != nil {
		log.Fatal(err)
	}
	var extpacksInstalled bool = reExtpacksInstalled.MatchString(extpacksOutString)
	if extpacksInstalled == false {
		var returnString string = "Oracle VM VirtualBox Extension Pack is is not installed. Vrde may not " +
		"work as expected."
		return returnString
	}

//	Get the installed version of virtual box
	versionCmd := exec.Command("vboxmanage", "--version")
	versionOut, err := versionCmd.Output()
	if err != nil {
		log.Fatal(err)
	}

//	Use regexp to strip extraneous string
	reVersion, err := regexp.Compile("[0-9]+\\.[0-9]+\\.[0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	versionOutMatch := reVersion.Find(versionOut)
	versionOutString := string(versionOutMatch)

//	Regexp to confirm if the extpack version matches the VirtualBox version
	reExtpacksMatchVersion, err := regexp.Compile(versionOutString)
	var extpacksMatchVersion bool = reExtpacksMatchVersion.MatchString(extpacksOutString)
	if extpacksMatchVersion == false {
		var returnString string = "The Oracle VM VirtualBox Extension Pack version does not match " +
		"the\nVirtualBox version. Vrde may not work as expected."
		return returnString
	}

	var returnString string = ""
	return returnString

}

// These steps enable the use of vrde, consequently allowing users to
// connect to remote and local virtual machines using software complient
// with Remote Desktop Protocol
type stepSetVrde struct{}

func (s *stepSetVrde) Run(state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*config)
	driver := state.Get("driver").(vboxcommon.Driver)
	ui := state.Get("ui").(packer.Ui)
	vmName := state.Get("vmName").(string)

	var vrdeValue string

//	The vboxmanage --vrde flag  takes a string, so we must pass bool
//  to string
	if config.Vrde == true {
		vrdeValue = "on"
	} else if config.Vrde == false {
		vrdeValue = "off"
	}

	command := []string{
		"modifyvm", vmName,
		"--vrde", vrdeValue,
	}

	var extpackStatusString string = extpackStatus()
	ui.Say(fmt.Sprintf("%s", extpackStatusString))
	ui.Say(fmt.Sprintf("Setting vrde to %s", vrdeValue))
	err := driver.VBoxManage(command...)
	if err != nil {
		err := fmt.Errorf("Error setting vrde: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepSetVrde) Cleanup(state multistep.StateBag) {}
