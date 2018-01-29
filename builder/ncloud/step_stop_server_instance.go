package ncloud

import (
	"context"
	"fmt"
	"log"
	"time"

	ncloud "github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepStopServerInstance struct {
	Conn               *ncloud.Conn
	StopServerInstance func(serverInstanceNo string) error
	Say                func(message string)
	Error              func(e error)
}

func NewStepStopServerInstance(conn *ncloud.Conn, ui packer.Ui) *StepStopServerInstance {
	var step = &StepStopServerInstance{
		Conn:  conn,
		Say:   func(message string) { ui.Say(message) },
		Error: func(e error) { ui.Error(e.Error()) },
	}

	step.StopServerInstance = step.stopServerInstance

	return step
}

func (s *StepStopServerInstance) stopServerInstance(serverInstanceNo string) error {
	reqParams := new(ncloud.RequestStopServerInstances)
	reqParams.ServerInstanceNoList = []string{serverInstanceNo}

	serverInstanceList, err := s.Conn.StopServerInstances(reqParams)
	if err != nil {
		return err
	}

	s.Say(fmt.Sprintf("Server Instance is stopping. Server InstanceNo is %s", serverInstanceList.ServerInstanceList[0].ServerInstanceNo))
	log.Println("Server Instance information : ", serverInstanceList.ServerInstanceList[0])

	if err := waiterServerInstanceStatus(s.Conn, serverInstanceNo, "NSTOP", 5*time.Minute); err != nil {
		return err
	}

	s.Say(fmt.Sprintf("Server Instance stopped. Server InstanceNo is %s", serverInstanceList.ServerInstanceList[0].ServerInstanceNo))

	return nil
}

func (s *StepStopServerInstance) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	s.Say("Stop Server Instance")

	var serverInstanceNo = state.Get("InstanceNo").(string)

	err := s.StopServerInstance(serverInstanceNo)

	return processStepResult(err, s.Error, state)
}

func (*StepStopServerInstance) Cleanup(multistep.StateBag) {
}
