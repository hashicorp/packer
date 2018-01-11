package ncloud

import (
	"errors"
	"fmt"
	"log"
	"time"

	ncloud "github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type StepDeletePublicIPInstance struct {
	Conn                   *ncloud.Conn
	DeletePublicIPInstance func(publicIPInstanceNo string) error
	Say                    func(message string)
	Error                  func(e error)
}

func NewStepDeletePublicIPInstance(conn *ncloud.Conn, ui packer.Ui) *StepDeletePublicIPInstance {
	var step = &StepDeletePublicIPInstance{
		Conn:  conn,
		Say:   func(message string) { ui.Say(message) },
		Error: func(e error) { ui.Error(e.Error()) },
	}

	step.DeletePublicIPInstance = step.deletePublicIPInstance

	return step
}

func (s *StepDeletePublicIPInstance) deletePublicIPInstance(publicIPInstanceNo string) error {
	reqParams := new(ncloud.RequestDeletePublicIPInstances)
	reqParams.PublicIPInstanceNoList = []string{publicIPInstanceNo}

	c1 := make(chan error, 1)

	go func() {
		for {
			resp, err := s.Conn.DeletePublicIPInstances(reqParams)
			if err != nil && (resp.ReturnCode == 24073 || resp.ReturnCode == 25032) {
				// error code : 24073 : Unable to destroy the server since a public IP is associated with the server. First, please disassociate a public IP from the server.
				// error code : 25032 : You may not delete sk since (other) user is changing the target official IP settings.
				log.Println(resp.ReturnCode, resp.ReturnMessage)
			} else if err != nil {
				c1 <- fmt.Errorf("error code: %d, error message: %s", resp.ReturnCode, resp.ReturnMessage)
				return
			} else if err == nil {
				s.Say(fmt.Sprintf("Public IP Instance [%s] is deleted.", publicIPInstanceNo))
				c1 <- nil
				return
			}

			time.Sleep(time.Second * 5)
		}
	}()

	select {
	case res := <-c1:
		return res
	case <-time.After(time.Second * 60):
		return errors.New("TIMEOUT : Can't delete server instance")
	}
}

func (s *StepDeletePublicIPInstance) Run(state multistep.StateBag) multistep.StepAction {
	s.Say("Delete Public IP Instance")

	publicIPInstance := state.Get("PublicIPInstance").(*ncloud.PublicIPInstance)

	err := s.DeletePublicIPInstance(publicIPInstance.PublicIPInstanceNo)

	return processStepResult(err, s.Error, state)
}

func (*StepDeletePublicIPInstance) Cleanup(multistep.StateBag) {
}
