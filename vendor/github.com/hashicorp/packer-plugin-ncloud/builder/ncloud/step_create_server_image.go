package ncloud

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/server"
	"github.com/NaverCloudPlatform/ncloud-sdk-go-v2/services/vserver"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

const (
	ServerImageStatusCreated = "CREAT"
)

type StepCreateServerImage struct {
	Conn              *NcloudAPIClient
	CreateServerImage func(serverInstanceNo string) (*server.MemberServerImage, error)
	Say               func(message string)
	Error             func(e error)
	Config            *Config
}

func NewStepCreateServerImage(conn *NcloudAPIClient, ui packersdk.Ui, config *Config) *StepCreateServerImage {
	var step = &StepCreateServerImage{
		Conn:   conn,
		Say:    func(message string) { ui.Say(message) },
		Error:  func(e error) { ui.Error(e.Error()) },
		Config: config,
	}

	if config.SupportVPC {
		step.CreateServerImage = step.createVpcServerImage
	} else {
		step.CreateServerImage = step.createClassicServerImage
	}

	return step
}

func (s *StepCreateServerImage) createClassicServerImage(serverInstanceNo string) (*server.MemberServerImage, error) {
	// Can't create server image when status of server instance is stopping (not stopped)
	if err := waiterClassicServerInstanceStatus(s.Conn, serverInstanceNo, ServerInstanceStatusStopped, 1*time.Minute); err != nil {
		return nil, err
	}

	reqParams := &server.CreateMemberServerImageRequest{
		MemberServerImageName:        &s.Config.ServerImageName,
		MemberServerImageDescription: &s.Config.ServerImageDescription,
		ServerInstanceNo:             &serverInstanceNo,
	}

	memberServerImageList, err := s.Conn.server.V2Api.CreateMemberServerImage(reqParams)
	if err != nil {
		return nil, err
	}

	serverImage := memberServerImageList.MemberServerImageList[0]

	s.Say(fmt.Sprintf("Server Image[%s:%s] is creating...", *serverImage.MemberServerImageName, *serverImage.MemberServerImageNo))

	if err := waiterClassicMemberServerImageStatus(s.Conn, *serverImage.MemberServerImageNo, ServerImageStatusCreated, 6*time.Hour); err != nil {
		return nil, errors.New("TIMEOUT : Server Image is not created")
	}

	s.Say(fmt.Sprintf("Server Image[%s:%s] is created", *serverImage.MemberServerImageName, *serverImage.MemberServerImageNo))

	return serverImage, nil
}

func (s *StepCreateServerImage) createVpcServerImage(serverInstanceNo string) (*server.MemberServerImage, error) {
	// Can't create server image when status of server instance is stopping (not stopped)
	if err := waiterVpcServerInstanceStatus(s.Conn, serverInstanceNo, ServerInstanceStatusStopped, 1*time.Minute); err != nil {
		return nil, err
	}

	reqParams := &vserver.CreateMemberServerImageInstanceRequest{
		MemberServerImageName:        &s.Config.ServerImageName,
		MemberServerImageDescription: &s.Config.ServerImageDescription,
		ServerInstanceNo:             &serverInstanceNo,
	}

	memberServerImageList, err := s.Conn.vserver.V2Api.CreateMemberServerImageInstance(reqParams)
	if err != nil {
		return nil, err
	}

	serverImage := memberServerImageList.MemberServerImageInstanceList[0]

	s.Say(fmt.Sprintf("Server Image[%s:%s] is creating...", *serverImage.MemberServerImageName, *serverImage.MemberServerImageInstanceNo))

	if err := waiterVpcMemberServerImageStatus(s.Conn, *serverImage.MemberServerImageInstanceNo, ServerImageStatusCreated, 6*time.Hour); err != nil {
		return nil, errors.New("TIMEOUT : Server Image is not created")
	}

	s.Say(fmt.Sprintf("Server Image[%s:%s] is created", *serverImage.MemberServerImageName, *serverImage.MemberServerImageInstanceNo))

	result := &server.MemberServerImage{
		MemberServerImageNo:   serverImage.MemberServerImageInstanceNo,
		MemberServerImageName: serverImage.MemberServerImageName,
	}

	if serverImage.MemberServerImageInstanceStatus != nil {
		result.MemberServerImageStatus = &server.CommonCode{
			Code:     serverImage.MemberServerImageInstanceStatus.Code,
			CodeName: serverImage.MemberServerImageInstanceStatus.CodeName,
		}
	}

	return result, nil
}

func (s *StepCreateServerImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	s.Say("Create Server Image")

	serverInstanceNo := state.Get("instance_no").(string)

	serverImage, err := s.CreateServerImage(serverInstanceNo)
	if err == nil {
		state.Put("member_server_image", serverImage)
	}

	return processStepResult(err, s.Error, state)
}

func (*StepCreateServerImage) Cleanup(multistep.StateBag) {
}
