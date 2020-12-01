package cloudstack

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/uuid"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

type stepCreateSecurityGroup struct {
	tempSG string
}

func (s *stepCreateSecurityGroup) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*cloudstack.CloudStackClient)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	if len(config.SecurityGroups) > 0 {
		state.Put("security_groups", config.SecurityGroups)
		return multistep.ActionContinue
	}

	if !config.CreateSecurityGroup {
		return multistep.ActionContinue
	}

	ui.Say("Creating temporary Security Group...")

	p := client.SecurityGroup.NewCreateSecurityGroupParams(
		fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID()),
	)
	p.SetDescription("Temporary SG created by Packer")
	if config.Project != "" {
		p.SetProjectid(config.Project)
	}

	sg, err := client.SecurityGroup.CreateSecurityGroup(p)
	if err != nil {
		err := fmt.Errorf("Failed to create security group: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.tempSG = sg.Id
	state.Put("security_groups", []string{sg.Id})

	// Create Ingress rule
	i := client.SecurityGroup.NewAuthorizeSecurityGroupIngressParams()
	i.SetCidrlist(config.CIDRList)
	i.SetProtocol("TCP")
	i.SetSecuritygroupid(sg.Id)
	i.SetStartport(config.Comm.Port())
	i.SetEndport(config.Comm.Port())
	if config.Project != "" {
		i.SetProjectid(config.Project)
	}

	_, err = client.SecurityGroup.AuthorizeSecurityGroupIngress(i)
	if err != nil {
		err := fmt.Errorf("Failed to authorize security group ingress rule: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

// Cleanup any resources that may have been created during the Run phase.
func (s *stepCreateSecurityGroup) Cleanup(state multistep.StateBag) {
	client := state.Get("client").(*cloudstack.CloudStackClient)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packersdk.Ui)

	if s.tempSG == "" {
		return
	}

	ui.Say(fmt.Sprintf("Cleanup temporary security group: %s ...", s.tempSG))
	p := client.SecurityGroup.NewDeleteSecurityGroupParams()
	p.SetId(s.tempSG)
	if config.Project != "" {
		p.SetProjectid(config.Project)
	}

	if _, err := client.SecurityGroup.DeleteSecurityGroup(p); err != nil {
		ui.Error(err.Error())
		ui.Error(fmt.Sprintf("Error deleting security group: %s. Please destroy it manually.\n", s.tempSG))
	}
}
