package classic

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateIPReservation struct{}

func (s *stepCreateIPReservation) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	config := state.Get("config").(*Config)
	client := state.Get("client").(*compute.ComputeClient)
	iprClient := client.IPReservations()
	// TODO: add optional Name and Tags

	ipresName := fmt.Sprintf("ipres_%s_%s", config.ImageName, uuid.TimeOrderedUUID())
	ui.Message(fmt.Sprintf("Creating temporary IP reservation: %s", ipresName))

	IPInput := &compute.CreateIPReservationInput{
		ParentPool: compute.PublicReservationPool,
		Permanent:  true,
		Name:       ipresName,
	}
	ipRes, err := iprClient.CreateIPReservation(IPInput)

	if err != nil {
		err := fmt.Errorf("Error creating IP Reservation: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	state.Put("instance_ip", ipRes.IP)
	state.Put("ipres_name", ipresName)
	return multistep.ActionContinue
}

func (s *stepCreateIPReservation) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Cleaning up IP reservations...")
	client := state.Get("client").(*compute.ComputeClient)

	ipResName := state.Get("ipres_name").(string)
	input := compute.DeleteIPReservationInput{Name: ipResName}
	ipClient := client.IPReservations()
	err := ipClient.DeleteIPReservation(&input)
	if err != nil {
		fmt.Printf("error deleting IP reservation: %s", err.Error())
	}

}
