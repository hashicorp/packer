package common

import (
	"fmt"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"strings"
	"time"
)

type StepAMIRegionCopy struct {
	Regions        []string
	AMICopyTimeout time.Duration
}

type amiCopyResult struct {
	region, imageId string
	err             error
}

func (s *StepAMIRegionCopy) Run(state multistep.StateBag) multistep.StepAction {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)
	amis := state.Get("amis").(map[string]string)
	ami := amis[ec2conn.Region.Name]
	result := make(chan amiCopyResult)

	if len(s.Regions) == 0 {
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Copying AMI (%s) to other regions...", ami))
	for _, region := range s.Regions {
		go func(region string) {
			ui.Message(fmt.Sprintf("Copying to: %s", region))

			// Connect to the region where the AMI will be copied to
			conn := ec2.New(ec2conn.Auth, aws.Regions[region])
			resp, err := conn.CopyImage(&ec2.CopyImage{
				SourceRegion:  ec2conn.Region.Name,
				SourceImageId: ami,
			})

			if err != nil {
				err := fmt.Errorf("Error Copying AMI (%s) to region (%s): %s", ami, region, err)
				result <- amiCopyResult{err: err}
				return
			}

			stateChange := StateChangeConf{
				Conn:      conn,
				Pending:   []string{"pending"},
				Target:    "available",
				Refresh:   AMIStateRefreshFunc(conn, resp.ImageId),
				StepState: state,
			}

			ui.Say(fmt.Sprintf("Waiting for AMI (%s) in region (%s) to become ready...", resp.ImageId, region))
			if _, err := WaitForState(&stateChange); err != nil {
				err := fmt.Errorf("Error waiting for AMI (%s) in region (%s): %s", resp.ImageId, region, err)
				result <- amiCopyResult{err: err}
				return
			}

			result <- amiCopyResult{region: region, imageId: resp.ImageId}
		}(region)
	}

	errs := make([]error, 0)
	timeout := time.After(s.AMICopyTimeout)

	for i := 0; i < len(s.Regions); i++ {
		select {
		case r := <-result:
			if r.err != nil {
				ui.Error(r.err.Error())
				errs = append(errs, r.err)
			} else {
				ui.Say(fmt.Sprintf("AMI (%s) in region (%s) is now ready", r.imageId, r.region))
				amis[r.region] = r.imageId
			}
		case <-timeout:
			errs = append(errs, fmt.Errorf("Timed out copying AMI (%s) to other regions", ami))
			break
		}
	}

	if len(errs) > 0 {
		errors := make([]string, len(errs))
		for i, err := range errs {
			errors[i] = fmt.Sprintf("* %s", err)
		}
		state.Put("error", fmt.Errorf(
			"The following %d error(s) occurred copying AMI (%s) other regions:\n\n%s",
			len(errs), ami, strings.Join(errors, "\n"),
		))
		return multistep.ActionHalt
	}

	state.Put("amis", amis)
	return multistep.ActionContinue
}

func (s *StepAMIRegionCopy) Cleanup(state multistep.StateBag) {
	// No cleanup...
}
