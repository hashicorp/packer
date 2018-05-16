package classic

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepSecurity struct{}

func (s *stepSecurity) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	commType := ""
	if config.Comm.Type == "ssh" {
		commType = "SSH"
	} else if config.Comm.Type == "winrm" {
		commType = "WINRM"
	}

	ui.Say(fmt.Sprintf("Configuring security lists and rules to enable %s access...", commType))

	client := state.Get("client").(*compute.ComputeClient)
	runUUID := uuid.TimeOrderedUUID()

	namePrefix := fmt.Sprintf("/Compute-%s/%s/", config.IdentityDomain, config.Username)
	secListName := fmt.Sprintf("Packer_%s_Allow_%s_%s", commType, config.ImageName, runUUID)
	secListClient := client.SecurityLists()
	secListInput := compute.CreateSecurityListInput{
		Description: fmt.Sprintf("Packer-generated security list to give packer %s access", commType),
		Name:        namePrefix + secListName,
	}
	_, err := secListClient.CreateSecurityList(&secListInput)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			err = fmt.Errorf("Error creating security List to"+
				" allow Packer to connect to Oracle instance via %s: %s", commType, err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}
	// DOCS NOTE: user must have Compute_Operations role
	// Create security rule that allows Packer to connect via SSH or winRM
	var application string
	if commType == "SSH" {
		application = "/oracle/public/ssh"
	} else if commType == "WINRM" {
		// Check to see whether a winRM security application is already defined
		applicationClient := client.SecurityApplications()
		application = fmt.Sprintf("packer_winRM_%s", runUUID)
		applicationInput := compute.CreateSecurityApplicationInput{
			Description: "Allows Packer to connect to instance via winRM",
			DPort:       "5985-5986",
			Name:        application,
			Protocol:    "TCP",
		}
		_, err = applicationClient.CreateSecurityApplication(&applicationInput)
		if err != nil {
			err = fmt.Errorf("Error creating security application to"+
				" allow Packer to connect to Oracle instance via %s: %s", commType, err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
		state.Put("winrm_application", application)
	}
	secRulesClient := client.SecRules()
	secRuleName := fmt.Sprintf("Packer-allow-%s-Rule_%s_%s", commType,
		config.ImageName, runUUID)
	secRulesInput := compute.CreateSecRuleInput{
		Action:          "PERMIT",
		Application:     application,
		Description:     "Packer-generated security rule to allow ssh/winrm",
		DestinationList: "seclist:" + namePrefix + secListName,
		Name:            namePrefix + secRuleName,
		SourceList:      config.SSHSourceList,
	}

	_, err = secRulesClient.CreateSecRule(&secRulesInput)
	if err != nil {
		err = fmt.Errorf("Error creating security rule to"+
			" allow Packer to connect to Oracle instance: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	state.Put("security_rule_name", secRuleName)
	state.Put("security_list", secListName)
	return multistep.ActionContinue
}

func (s *stepSecurity) Cleanup(state multistep.StateBag) {
	secRuleName, ok := state.GetOk("security_rule_name")
	if !ok {
		return
	}
	secListName, ok := state.GetOk("security_list")
	if !ok {
		return
	}

	client := state.Get("client").(*compute.ComputeClient)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)

	ui.Say("Deleting temporary rules and lists...")

	namePrefix := fmt.Sprintf("/Compute-%s/%s/", config.IdentityDomain, config.Username)
	// delete security rules that Packer generated
	secRulesClient := client.SecRules()
	ruleInput := compute.DeleteSecRuleInput{Name: namePrefix + secRuleName.(string)}
	err := secRulesClient.DeleteSecRule(&ruleInput)
	if err != nil {
		ui.Say(fmt.Sprintf("Error deleting the packer-generated security rule %s; "+
			"please delete manually. (error: %s)", secRuleName.(string), err.Error()))
	}

	// delete security list that Packer generated
	secListClient := client.SecurityLists()
	input := compute.DeleteSecurityListInput{Name: namePrefix + secListName.(string)}
	err = secListClient.DeleteSecurityList(&input)
	if err != nil {
		ui.Say(fmt.Sprintf("Error deleting the packer-generated security list %s; "+
			"please delete manually. (error : %s)", secListName.(string), err.Error()))
	}

	// Some extra cleanup if we used the winRM communicator
	if config.Comm.Type == "winrm" {
		// Delete the packer-generated application
		application, ok := state.GetOk("winrm_application")
		if !ok {
			return
		}
		applicationClient := client.SecurityApplications()
		deleteApplicationInput := compute.DeleteSecurityApplicationInput{
			Name: namePrefix + application.(string),
		}
		err = applicationClient.DeleteSecurityApplication(&deleteApplicationInput)
		if err != nil {
			ui.Say(fmt.Sprintf("Error deleting the packer-generated winrm security application %s; "+
				"please delete manually. (error : %s)", application.(string), err.Error()))
		}
	}

}
