package common

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	retry "github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepPreValidate provides an opportunity to pre-validate any configuration for
// the build before actually doing any time consuming work
//
type StepPreValidate struct {
	DestAmiName     string
	ForceDeregister bool
}

func (s *StepPreValidate) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	if accessConfig, ok := state.GetOk("access_config"); ok {
		accessconf := accessConfig.(*AccessConfig)
		if !accessconf.VaultAWSEngine.Empty() {
			// loop over the authentication a few times to give vault-created creds
			// time to become eventually-consistent
			ui.Say("You're using Vault-generated AWS credentials. It may take a " +
				"few moments for them to become available on AWS. Waiting...")
			err := retry.Retry(0.2, 30, 11, func(_ uint) (bool, error) {
				ec2conn, err := accessconf.NewEC2Connection()
				if err != nil {
					return true, err
				}
				_, err = listEC2Regions(ec2conn)
				if err != nil {
					if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "AuthFailure" {
						log.Printf("Waiting for Vault-generated AWS credentials" +
							" to pass authentication... trying again.")
						return false, nil
					}
					return true, err
				}
				return true, nil
			})

			if err != nil {
				state.Put("error", fmt.Errorf("Was unable to Authenticate to AWS using Vault-"+
					"Generated Credentials within the retry timeout."))
				return multistep.ActionHalt
			}
		}

		if amiConfig, ok := state.GetOk("ami_config"); ok {
			amiconf := amiConfig.(*AMIConfig)
			if !amiconf.AMISkipRegionValidation {
				regionsToValidate := append(amiconf.AMIRegions, accessconf.RawRegion)
				err := accessconf.ValidateRegion(regionsToValidate...)
				if err != nil {
					state.Put("error", fmt.Errorf("error validating regions: %v", err))
					return multistep.ActionHalt
				}
			}
		}
	}

	if s.ForceDeregister {
		ui.Say("Force Deregister flag found, skipping prevalidating AMI Name")
		return multistep.ActionContinue
	}

	ec2conn := state.Get("ec2").(*ec2.EC2)

	ui.Say(fmt.Sprintf("Prevalidating AMI Name: %s", s.DestAmiName))
	resp, err := ec2conn.DescribeImages(&ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{{
			Name:   aws.String("name"),
			Values: []*string{aws.String(s.DestAmiName)},
		}}})

	if err != nil {
		err := fmt.Errorf("Error querying AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(resp.Images) > 0 {
		err := fmt.Errorf("Error: name conflicts with an existing AMI: %s", *resp.Images[0].ImageId)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepPreValidate) Cleanup(multistep.StateBag) {}
