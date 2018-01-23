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
	// TODO create overrides that allow savvy users to add the image to their
	// own security lists instead of ours
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Configuring security lists and rules...")
	config := state.Get("config").(*Config)
	client := state.Get("client").(*compute.ComputeClient)

	secListName := fmt.Sprintf("/Compute-%s/%s/Packer_SSH_Allow",
		config.IdentityDomain, config.Username)
	secListClient := client.SecurityLists()
	secListInput := compute.CreateSecurityListInput{
		Description: "Packer-generated security list to give packer ssh access",
		Name:        secListName,
	}
	_, err := secListClient.CreateSecurityList(&secListInput)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			err = fmt.Errorf("Error creating security security IP List to"+
				" allow Packer to connect to Oracle instance via SSH: %s", err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}
	secListURI := fmt.Sprintf("%s/seclist/Compute-%s/%s/Packer_SSH_Allow",
		config.APIEndpoint, config.IdentityDomain, config.Username)
	log.Printf("Megan secListURI is %s", secListURI)
	// DOCS NOTE: user must have Compute_Operations role
	// Create security rule that allows Packer to connect via SSH

	secRulesClient := client.SecRules()
	secRulesInput := compute.CreateSecRuleInput{
		Action:          "PERMIT",
		Application:     "/oracle/public/ssh",
		Description:     "Packer-generated security rule to allow ssh",
		DestinationList: fmt.Sprintf("seclist:%s", secListName),
		Name:            "Packer-allow-SSH-Rule",
		SourceList:      "seciplist:/oracle/public/public-internet",
	}

	_, err = secRulesClient.CreateSecRule(&secRulesInput)
	if err != nil {
		err = fmt.Errorf("Error creating security rule to allow Packer to connect to Oracle instance via SSH: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put("security_list", secListName)
	return multistep.ActionContinue
}

func (s *stepSecurity) Cleanup(state multistep.StateBag) {
	// Nothing to do
}
