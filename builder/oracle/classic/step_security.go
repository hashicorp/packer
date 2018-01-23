package classic

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepSecurity struct{}

func (s *stepSecurity) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Configuring security lists and rules to enable SSH access...")

	config := state.Get("config").(*Config)
	client := state.Get("client").(*compute.ComputeClient)

	secListName := fmt.Sprintf("/Compute-%s/%s/Packer_SSH_Allow_%s",
		config.IdentityDomain, config.Username, config.ImageName)
	secListClient := client.SecurityLists()
	secListInput := compute.CreateSecurityListInput{
		Description: "Packer-generated security list to give packer ssh access",
		Name:        secListName,
	}
	_, err := secListClient.CreateSecurityList(&secListInput)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			err = fmt.Errorf("Error creating security List to"+
				" allow Packer to connect to Oracle instance via SSH: %s", err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}
	// DOCS NOTE: user must have Compute_Operations role
	// Create security rule that allows Packer to connect via SSH
	secRulesClient := client.SecRules()
	secRulesInput := compute.CreateSecRuleInput{
		Action:          "PERMIT",
		Application:     "/oracle/public/ssh",
		Description:     "Packer-generated security rule to allow ssh",
		DestinationList: fmt.Sprintf("seclist:%s", secListName),
		Name:            fmt.Sprintf("Packer-allow-SSH-Rule_%s", config.ImageName),
		SourceList:      config.SSHSourceList,
	}

	secRuleName := fmt.Sprintf("/Compute-%s/%s/Packer-allow-SSH-Rule_%s",
		config.IdentityDomain, config.Username, config.ImageName)
	_, err = secRulesClient.CreateSecRule(&secRulesInput)
	if err != nil {
		log.Printf(err.Error())
		if !strings.Contains(err.Error(), "already exists") {
			err = fmt.Errorf("Error creating security rule to"+
				" allow Packer to connect to Oracle instance via SSH: %s", err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}
	state.Put("security_rule_name", secRuleName)
	state.Put("security_list", secListName)
	return multistep.ActionContinue
}

func (s *stepSecurity) Cleanup(state multistep.StateBag) {
	client := state.Get("client").(*compute.ComputeClient)
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Deleting the packer-generated security rules and lists...")
	// delete security list that Packer generated
	secListName := state.Get("security_list").(string)
	secListClient := client.SecurityLists()
	input := compute.DeleteSecurityListInput{Name: secListName}
	err := secListClient.DeleteSecurityList(&input)
	if err != nil {
		ui.Say(fmt.Sprintf("Error deleting the packer-generated security list %s; "+
			"please delete manually. (error : %s)", secListName, err.Error()))
	}
	// delete security rules that Packer generated
	secRuleName := state.Get("security_rule_name").(string)
	secRulesClient := client.SecRules()
	ruleInput := compute.DeleteSecRuleInput{Name: secRuleName}
	err = secRulesClient.DeleteSecRule(&ruleInput)
	if err != nil {
		ui.Say(fmt.Sprintf("Error deleting the packer-generated security rule %s; "+
			"please delete manually. (error: %s)", secRuleName, err.Error()))
	}
	return
}
