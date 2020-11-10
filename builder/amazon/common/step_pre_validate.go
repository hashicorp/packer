package common

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/hashicorp/packer/builder/amazon/common/awserrors"
	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// StepPreValidate provides an opportunity to pre-validate any configuration for
// the build before actually doing any time consuming work
//
type StepPreValidate struct {
	DestAmiName        string
	ForceDeregister    bool
	AMISkipBuildRegion bool
	VpcId              string
	SubnetId           string
	HasSubnetFilter    bool
}

func (s *StepPreValidate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	if accessConfig, ok := state.GetOk("access_config"); ok {
		accessconf := accessConfig.(*AccessConfig)
		if !accessconf.VaultAWSEngine.Empty() {
			// loop over the authentication a few times to give vault-created creds
			// time to become eventually-consistent
			ui.Say("You're using Vault-generated AWS credentials. It may take a " +
				"few moments for them to become available on AWS. Waiting...")
			err := retry.Config{
				Tries: 11,
				ShouldRetry: func(err error) bool {
					if awserrors.Matches(err, "AuthFailure", "") {
						log.Printf("Waiting for Vault-generated AWS credentials" +
							" to pass authentication... trying again.")
						return true
					}
					return false
				},
				RetryDelay: (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 30 * time.Second, Multiplier: 2}).Linear,
			}.Run(ctx, func(ctx context.Context) error {
				ec2conn, err := accessconf.NewEC2Connection()
				if err != nil {
					return err
				}
				_, err = listEC2Regions(ec2conn)
				return err
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

	if s.AMISkipBuildRegion {
		ui.Say("skip_build_region was set; not prevalidating AMI name")
		return multistep.ActionContinue
	}

	ec2conn := state.Get("ec2").(*ec2.EC2)

	// Validate VPC settings for non-default VPCs
	ui.Say("Prevalidating any provided VPC information")
	if err := s.checkVpc(ec2conn); err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Prevalidating AMI Name: %s", s.DestAmiName))
	req, resp := ec2conn.DescribeImagesRequest(&ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{{
			Name:   aws.String("name"),
			Values: []*string{aws.String(s.DestAmiName)},
		}}})
	req.RetryCount = 11

	if err := req.Send(); err != nil {
		err = fmt.Errorf("Error querying AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(resp.Images) > 0 {
		err := fmt.Errorf("Error: AMI Name: '%s' is used by an existing AMI: %s", *resp.Images[0].Name, *resp.Images[0].ImageId)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepPreValidate) checkVpc(conn ec2iface.EC2API) error {
	if s.VpcId == "" || (s.VpcId != "" && (s.SubnetId != "" || s.HasSubnetFilter)) {
		// Skip validation if:
		// * The user has not provided a VpcId.
		// * Both VpcId and SubnetId are provided; AWS API will error if something is wrong.
		// * Both VpcId and SubnetFilter are provided
		return nil
	}

	res, err := conn.DescribeVpcs(&ec2.DescribeVpcsInput{VpcIds: []*string{aws.String(s.VpcId)}})
	if awserrors.Matches(err, "InvalidVpcID.NotFound", "") || err != nil {
		return fmt.Errorf("Error retrieving VPC information for vpc_id %s: %s", s.VpcId, err)
	}

	if res != nil && len(res.Vpcs) == 1 && res.Vpcs[0] != nil {
		if isDefault := aws.BoolValue(res.Vpcs[0].IsDefault); !isDefault {
			return fmt.Errorf("Error: subnet_id or subnet_filter must be provided for non-default VPCs (%s)", s.VpcId)
		}
	}
	return nil
}

// Cleanup ...
func (s *StepPreValidate) Cleanup(multistep.StateBag) {}
