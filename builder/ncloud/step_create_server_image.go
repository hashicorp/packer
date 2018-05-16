package ncloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	ncloud "github.com/NaverCloudPlatform/ncloud-sdk-go/sdk"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepCreateServerImage struct {
	Conn              *ncloud.Conn
	CreateServerImage func(serverInstanceNo string) (*ncloud.ServerImage, error)
	Say               func(message string)
	Error             func(e error)
	Config            *Config
}

func NewStepCreateServerImage(conn *ncloud.Conn, ui packer.Ui, config *Config) *StepCreateServerImage {
	var step = &StepCreateServerImage{
		Conn:   conn,
		Say:    func(message string) { ui.Say(message) },
		Error:  func(e error) { ui.Error(e.Error()) },
		Config: config,
	}

	step.CreateServerImage = step.createServerImage

	return step
}

func (s *StepCreateServerImage) createServerImage(serverInstanceNo string) (*ncloud.ServerImage, error) {
	// Can't create server image when status of server instance is stopping (not stopped)
	if err := waiterServerInstanceStatus(s.Conn, serverInstanceNo, "NSTOP", 1*time.Minute); err != nil {
		return nil, err
	}

	reqParams := new(ncloud.RequestCreateServerImage)
	reqParams.MemberServerImageName = s.Config.ServerImageName
	reqParams.MemberServerImageDescription = s.Config.ServerImageDescription
	reqParams.ServerInstanceNo = serverInstanceNo

	memberServerImageList, err := s.Conn.CreateMemberServerImage(reqParams)
	if err != nil {
		return nil, err
	}

	serverImage := memberServerImageList.MemberServerImageList[0]

	s.Say(fmt.Sprintf("Server Image[%s:%s] is creating...", serverImage.MemberServerImageName, serverImage.MemberServerImageNo))

	if err := waiterMemberServerImageStatus(s.Conn, serverImage.MemberServerImageNo, "CREAT", 6*time.Hour); err != nil {
		return nil, errors.New("TIMEOUT : Server Image is not created")
	}

	s.Say(fmt.Sprintf("Server Image[%s:%s] is created", serverImage.MemberServerImageName, serverImage.MemberServerImageNo))

	return &serverImage, nil
}

func (s *StepCreateServerImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	s.Say("Create Server Image")

	serverInstanceNo := state.Get("InstanceNo").(string)

	serverImage, err := s.CreateServerImage(serverInstanceNo)
	if err == nil {
		state.Put("memberServerImage", serverImage)
	}

	return processStepResult(err, s.Error, state)
}

func (*StepCreateServerImage) Cleanup(multistep.StateBag) {
}
