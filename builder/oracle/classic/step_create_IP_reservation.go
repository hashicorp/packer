package classic

import (
	"fmt"
	"log"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
)

type stepCreateIPReservation struct{}

func (s *stepCreateIPReservation) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating IP reservation...")
	config := state.Get("config").(*Config)
	client := state.Get("client").(*compute.ComputeClient)
	iprClient := client.IPReservations()
	// TODO: add optional Name and Tags
	IPInput := &compute.CreateIPReservationInput{
		ParentPool: compute.PublicReservationPool,
		Permanent:  true,
		Name:       fmt.Sprintf("ipres_%s", config.ImageName),
	}
	ipRes, err := iprClient.CreateIPReservation(IPInput)
	if err != nil {
		log.Printf("Error creating IP Reservation: %s", err)
		return multistep.ActionHalt
	}
	state.Put("instance_ip", ipRes.IP)
	log.Printf("debug: ipRes is %#v", ipRes)
	return multistep.ActionContinue
}

func (s *stepCreateIPReservation) Cleanup(state multistep.StateBag) {
	// Nothing to do
}
