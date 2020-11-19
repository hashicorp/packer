package ncloud

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepStopServerInstance struct {
	Conn               *NcloudAPIClient
	StopServerInstance func(serverInstanceNo string) error
	Say                func(message string)
	Error              func(e error)
}

func NewStepStopServerInstance(conn *NcloudAPIClient, ui packersdk.Ui) *StepStopServerInstance {
	var step = &StepStopServerInstance{
		Conn:  conn,
		Say:   func(message string) { ui.Say(message) },
		Error: func(e error) { ui.Error(e.Error()) },
	}

	step.StopServerInstance = step.stopServerInstance

	return step
}

func (s *StepStopServerInstance) stopServerInstance(serverInstanceNo string) error {
	reqParams := new(server.StopServerInstancesRequest)
	reqParams.ServerInstanceNoList = []*string{&serverInstanceNo}

	serverInstanceList, err := s.Conn.server.V2Api.StopServerInstances(reqParams)
	if err != nil {
		return err
	}

	s.Say(fmt.Sprintf("Server Instance is stopping. Server InstanceNo is %s", *serverInstanceList.ServerInstanceList[0].ServerInstanceNo))
	log.Println("Server Instance information : ", serverInstanceList.ServerInstanceList[0])

	if err := waiterServerInstanceStatus(s.Conn, serverInstanceNo, "NSTOP", 5*time.Minute); err != nil {
		return err
	}

	s.Say(fmt.Sprintf("Server Instance stopped. Server InstanceNo is %s", *serverInstanceList.ServerInstanceList[0].ServerInstanceNo))

	return nil
}

func (s *StepStopServerInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.Say("Stop Server Instance")

	var serverInstanceNo = state.Get("InstanceNo").(string)

	err := s.StopServerInstance(serverInstanceNo)

	return processStepResult(err, s.Error, state)
}

func (*StepStopServerInstance) Cleanup(multistep.StateBag) {
}
