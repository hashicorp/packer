package common

import (
	"fmt"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"sync"
)

type StepAMIRegionCopy struct {
	Regions []string
}

func (s *StepAMIRegionCopy) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)
	amis := state.Get("amis").(map[string]string)
	ami := amis[ec2conn.Region.Name]

	if len(s.Regions) == 0 {
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Copying AMI (%s) to other regions...", ami))

	var lock sync.Mutex
	var wg sync.WaitGroup
	errs := new(packer.MultiError)
	for _, region := range s.Regions {
		wg.Add(1)
		ui.Message(fmt.Sprintf("Copying to: %s", region))

		go func(region string) {
			defer wg.Done()
			id, err := amiRegionCopy(state, ec2conn.Auth, ami,
				aws.Regions[region], ec2conn.Region)

			lock.Lock()
			defer lock.Unlock()
			amis[region] = id
			if err != nil {
				errs = packer.MultiErrorAppend(errs, err)
			}
		}(region)
	}

	// TODO(mitchellh): Wait but also allow for cancels to go through...
	ui.Message("Waiting for all copies to complete...")
	wg.Wait()

	// If there were errors, show them
	if len(errs.Errors) > 0 {
		state.Put("error", errs)
		ui.Error(errs.Error())
		return multistep.ActionHalt
	}

	state.Put("amis", amis)
	return multistep.ActionContinue
}

func (s *StepAMIRegionCopy) Cleanup(state multistep.StateBag) {
	// No cleanup...
}

// amiRegionCopy does a copy for the given AMI to the target region and
// returns the resulting ID or error.
func amiRegionCopy(state multistep.StateBag, auth aws.Auth, imageId string,
	target aws.Region, source aws.Region) (string, error) {

	// Connect to the region where the AMI will be copied to
	regionconn := ec2.New(auth, target)
	resp, err := regionconn.CopyImage(&ec2.CopyImage{
		SourceRegion:  source.Name,
		SourceImageId: imageId,
	})

	if err != nil {
		return "", fmt.Errorf("Error Copying AMI (%s) to region (%s): %s",
			imageId, target, err)
	}

	stateChange := StateChangeConf{
		Pending:   []string{"pending"},
		Target:    "available",
		Refresh:   AMIStateRefreshFunc(regionconn, resp.ImageId),
		StepState: state,
	}

	if _, err := WaitForState(&stateChange); err != nil {
		return "", fmt.Errorf("Error waiting for AMI (%s) in region (%s): %s",
			resp.ImageId, target, err)
	}

	return resp.ImageId, nil
}
