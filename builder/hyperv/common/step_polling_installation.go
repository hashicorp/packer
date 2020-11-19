package common

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

const port string = "13000"

type StepPollingInstallation struct {
}

func (s *StepPollingInstallation) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	errorMsg := "Error polling VM: %s"
	vmIp := state.Get("ip").(string)

	ui.Say("Start polling VM to check the installation is complete...")
	host := "'" + vmIp + "'," + port

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock {function foo(){try{$client=New-Object System.Net.Sockets.TcpClient(")
	blockBuffer.WriteString(host)
	blockBuffer.WriteString(") -ErrorAction SilentlyContinue;if($client -eq $null){return $false}}catch{return $false}return $true} foo}")

	count := 60
	var duration time.Duration = 20
	sleepTime := time.Second * duration

	var res string

	for count > 0 {
		log.Println(fmt.Sprintf("Connecting vm (%s)...", host))
		cmd := exec.Command("powershell", blockBuffer.String())
		cmdOut, err := cmd.Output()
		if err != nil {
			err := fmt.Errorf(errorMsg, err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		res = strings.TrimSpace(string(cmdOut))

		if res != "False" {
			ui.Say("Signal was received from the VM")
			// Sleep before starting provision
			time.Sleep(time.Second * 30)
			break
		}

		log.Println(fmt.Sprintf("Slipping for more %v seconds...", uint(duration)))
		time.Sleep(sleepTime)
		count--
	}

	if count == 0 {
		err := fmt.Errorf(errorMsg, "a signal from vm was not received in a given time period ")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say("The installation complete")

	return multistep.ActionContinue
}

func (s *StepPollingInstallation) Cleanup(state multistep.StateBag) {

}
