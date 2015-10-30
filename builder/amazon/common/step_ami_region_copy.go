package common

import (
	"fmt"

	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepAMIRegionCopy struct {
	AccessConfig *AccessConfig
	Regions      []string
	Name         string
}

func (s *StepAMIRegionCopy) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)
	amis := state.Get("amis").(map[string]string)
	ami := amis[*ec2conn.Config.Region]

	if len(s.Regions) == 0 {
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Copying AMI (%s) to other regions...", ami))

	var lock sync.Mutex
	var wg sync.WaitGroup
	errs := new(packer.MultiError)
	for _, region := range s.Regions {
		if region == *ec2conn.Config.Region {
			ui.Message(fmt.Sprintf(
				"Avoiding copying AMI to duplicate region %s", region))
			continue
		}

		wg.Add(1)
		ui.Message(fmt.Sprintf("Copying to: %s", region))

		go func(region string) {
			defer wg.Done()
			id, err := amiRegionCopy(state, s.AccessConfig, s.Name, ami, region, *ec2conn.Config.Region)

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
func amiRegionCopy(state multistep.StateBag, config *AccessConfig, name string, imageId string,
	target string, source string) (string, error) {

	// Connect to the region where the AMI will be copied to
	awsConfig, err := config.Config()
	if err != nil {
		return "", err
	}
	awsConfig.Region = aws.String(target)

	sess := session.New(awsConfig)
	regionconn := ec2.New(sess)

	resp, err := regionconn.CopyImage(&ec2.CopyImageInput{
		SourceRegion:  &source,
		SourceImageId: &imageId,
		Name:          &name,
	})

	if err != nil {
		return "", fmt.Errorf("Error Copying AMI (%s) to region (%s): %s",
			imageId, target, err)
	}

	stateChange := StateChangeConf{
		Pending:   []string{"pending"},
		Target:    "available",
		Refresh:   AMIStateRefreshFunc(regionconn, *resp.ImageId),
		StepState: state,
	}

	if _, err := WaitForState(&stateChange); err != nil {
		return "", fmt.Errorf("Error waiting for AMI (%s) in region (%s): %s",
			*resp.ImageId, target, err)
	}

	return *resp.ImageId, nil
}
