package ncloud

import (
	"context"
	"errors"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type StepTerminateServerInstance struct {
	Conn                    *NcloudAPIClient
	TerminateServerInstance func(serverInstanceNo string) error
	Say                     func(message string)
	Error                   func(e error)
}

func NewStepTerminateServerInstance(conn *NcloudAPIClient, ui packersdk.Ui) *StepTerminateServerInstance {
	var step = &StepTerminateServerInstance{
		Conn:  conn,
		Say:   func(message string) { ui.Say(message) },
		Error: func(e error) { ui.Error(e.Error()) },
	}

	step.TerminateServerInstance = step.terminateServerInstance

	return step
}

func (s *StepTerminateServerInstance) terminateServerInstance(serverInstanceNo string) error {
	reqParams := new(server.TerminateServerInstancesRequest)
	reqParams.ServerInstanceNoList = []*string{&serverInstanceNo}

	_, err := s.Conn.server.V2Api.TerminateServerInstances(reqParams)
	if err != nil {
		return err
	}

	c1 := make(chan error, 1)

	go func() {
		reqParams := new(server.GetServerInstanceListRequest)
		reqParams.ServerInstanceNoList = []*string{&serverInstanceNo}

		for {

			serverInstanceList, err := s.Conn.server.V2Api.GetServerInstanceList(reqParams)
			if err != nil {
				c1 <- err
				return
			} else if *serverInstanceList.TotalRows == 0 {
				c1 <- nil
				return
			}

			time.Sleep(time.Second * 3)
		}
	}()

	select {
	case res := <-c1:
		return res
	case <-time.After(time.Second * 60):
		return errors.New("TIMEOUT : Can't terminate server instance")
	}
}

func (s *StepTerminateServerInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.Say("Terminate Server Instance")

	var serverInstanceNo = state.Get("InstanceNo").(string)

	err := s.TerminateServerInstance(serverInstanceNo)

	return processStepResult(err, s.Error, state)
}

func (*StepTerminateServerInstance) Cleanup(multistep.StateBag) {
}
