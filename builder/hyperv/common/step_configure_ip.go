package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"strings"
	"time"
)

type StepConfigureIp struct {
}

func (s *StepConfigureIp) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)

	errorMsg := "Error configuring ip address: %s"
	vmName := state.Get("vmName").(string)

	ui.Say("Configuring ip address...")

	count := 60
	var duration time.Duration = 1
	sleepTime := time.Minute * duration
	var ip string

	for count != 0 {
		cmdOut, err := driver.GetVirtualMachineNetworkAdapterAddress(vmName)
		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		ip = strings.TrimSpace(string(cmdOut))

		if ip != "False" {
			break
		}

		log.Println(fmt.Sprintf("Waiting for another %v minutes...", uint(duration)))
		time.Sleep(sleepTime)
		count--
	}

	if count == 0 {
		err := fmt.Errorf(errorMsg, "IP address assigned to the adapter is empty")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("ip address is " + ip)

	hostName, err := driver.GetHostName(ip)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("hostname is " + hostName)

	state.Put("ip", ip)
	state.Put("hostname", hostName)

	return multistep.ActionContinue
}

func (s *StepConfigureIp) Cleanup(state multistep.StateBag) {
	// do nothing
}
