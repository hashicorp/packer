package classic

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type stepSecurity struct {
	CommType        string
	SecurityListKey string
	secListName     string
	secRuleName     string
}

func (s *stepSecurity) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)
	runID := state.Get("run_id").(string)
	client := state.Get("client").(*compute.Client)

	commType := ""
	if s.CommType == "ssh" {
		commType = "SSH"
	} else if s.CommType == "winrm" {
		commType = "WINRM"
	}
	secListName := fmt.Sprintf("Packer_%s_Allow_%s", commType, runID)

	if _, ok := state.GetOk(secListName); ok {
		log.Println("SecList created in earlier step, continuing")
		// copy sec list name to proper key
		state.Put(s.SecurityListKey, secListName)
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Configuring security lists and rules to enable %s access...", commType))
	log.Println(secListName)

	secListClient := client.SecurityLists()
	secListInput := compute.CreateSecurityListInput{
		Description: fmt.Sprintf("Packer-generated security list to give packer %s access", commType),
		Name:        config.Identifier(secListName),
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
		application = fmt.Sprintf("packer_winRM_%s", runID)
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
	secRuleName := fmt.Sprintf("Packer-allow-%s-Rule_%s", commType, runID)
	log.Println(secRuleName)
	secRulesInput := compute.CreateSecRuleInput{
		Action:          "PERMIT",
		Application:     application,
		Description:     "Packer-generated security rule to allow ssh/winrm",
		DestinationList: "seclist:" + config.Identifier(secListName),
		Name:            config.Identifier(secRuleName),
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
	state.Put(s.SecurityListKey, secListName)
	state.Put(secListName, true)
	s.secListName = secListName
	s.secRuleName = secRuleName
	return multistep.ActionContinue
}

func (s *stepSecurity) Cleanup(state multistep.StateBag) {
	if s.secListName == "" || s.secRuleName == "" {
		return
	}

	client := state.Get("client").(*compute.Client)
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)

	ui.Say("Deleting temporary rules and lists...")

	// delete security rules that Packer generated
	secRulesClient := client.SecRules()
	ruleInput := compute.DeleteSecRuleInput{
		Name: config.Identifier(s.secRuleName),
	}
	err := secRulesClient.DeleteSecRule(&ruleInput)
	if err != nil {
		ui.Say(fmt.Sprintf("Error deleting the packer-generated security rule %s; "+
			"please delete manually. (error: %s)", s.secRuleName, err.Error()))
	}

	// delete security list that Packer generated
	secListClient := client.SecurityLists()
	input := compute.DeleteSecurityListInput{Name: config.Identifier(s.secListName)}
	err = secListClient.DeleteSecurityList(&input)
	if err != nil {
		ui.Say(fmt.Sprintf("Error deleting the packer-generated security list %s; "+
			"please delete manually. (error : %s)", s.secListName, err.Error()))
	}

	// Some extra cleanup if we used the winRM communicator
	if s.CommType == "winrm" {
		// Delete the packer-generated application
		application, ok := state.GetOk("winrm_application")
		if !ok {
			return
		}
		applicationClient := client.SecurityApplications()
		deleteApplicationInput := compute.DeleteSecurityApplicationInput{
			Name: config.Identifier(application.(string)),
		}
		err = applicationClient.DeleteSecurityApplication(&deleteApplicationInput)
		if err != nil {
			ui.Say(fmt.Sprintf("Error deleting the packer-generated winrm security application %s; "+
				"please delete manually. (error : %s)", application.(string), err.Error()))
		}
	}

}
