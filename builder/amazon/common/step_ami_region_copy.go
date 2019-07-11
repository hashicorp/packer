package common

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type StepAMIRegionCopy struct {
	AccessConfig      *AccessConfig
	Regions           []string
	AMIKmsKeyId       string
	RegionKeyIds      map[string]string
	EncryptBootVolume *bool // nil means preserve
	Name              string
	OriginalRegion    string

	toDelete           string
	getRegionConn      func(*AccessConfig, string) (ec2iface.EC2API, error)
	AMISkipBuildRegion bool
}

func (s *StepAMIRegionCopy) DeduplicateRegions() {
	// Deduplicates regions by looping over the list of regions and storing
	// the regions as keys in a map. This saves users from accidentally copying
	// regions twice if they've added a region to a map twice.

	RegionMap := map[string]bool{}
	RegionSlice := []string{}

	for _, r := range s.Regions {
		RegionMap[r] = true
	}

	// Now print all those keys into the region slice again
	for k, _ := range RegionMap {
		RegionSlice = append(RegionSlice, k)
	}

	s.Regions = RegionSlice
}

func (s *StepAMIRegionCopy) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	amis := state.Get("amis").(map[string]string)
	snapshots := state.Get("snapshots").(map[string][]string)

	ami := amis[s.OriginalRegion]
	// Always copy back into original region to preserve the ami name
	if s.EncryptBootVolume != nil || s.AMISkipBuildRegion {
		// if we haven't specificed encryption and we aren't skipping the save
		// to the build region, we don't have anything to delete
		s.toDelete = ami
	}

	if s.EncryptBootVolume != nil {
		if !s.AMISkipBuildRegion {
			s.Regions = append(s.Regions, s.OriginalRegion)
		}
		// Now that we've added OriginalRegion, best to make sure there aren't
		// any duplicates hanging around; duplicates will waste time.
		s.DeduplicateRegions()

		if *s.EncryptBootVolume {
			// encrypt_boot is true, so we have to copy the temporary
			// AMI with required encryption setting.
			// temp image was created by stepCreateAMI.
			if s.RegionKeyIds == nil {
				s.RegionKeyIds = make(map[string]string)
			}

			// Make sure the kms_key_id for the original region is in the map
			if _, ok := s.RegionKeyIds[s.OriginalRegion]; !ok {
				s.RegionKeyIds[s.OriginalRegion] = s.AMIKmsKeyId
			}
		}
	}

	if len(s.Regions) == 0 {
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Copying/Encrypting AMI (%s) to other regions...", ami))

	var lock sync.Mutex
	var wg sync.WaitGroup
	var regKeyID string
	errs := new(packer.MultiError)

	wg.Add(len(s.Regions))
	for _, region := range s.Regions {
		ui.Message(fmt.Sprintf("Copying to: %s", region))

		if s.EncryptBootVolume != nil && *s.EncryptBootVolume {
			regKeyID = s.RegionKeyIds[region]
		}

		go func(region string) {
			defer wg.Done()
			id, snapshotIds, err := s.amiRegionCopy(ctx, state, s.AccessConfig, s.Name, ami, region, s.OriginalRegion, regKeyID, s.EncryptBootVolume)
			lock.Lock()
			defer lock.Unlock()
			amis[region] = id
			snapshots[region] = snapshotIds
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
	ec2conn := state.Get("ec2").(*ec2.EC2)
	ui := state.Get("ui").(packer.Ui)

	if len(s.toDelete) == 0 {
		return
	}

	// Delete the unencrypted amis and snapshots
	ui.Say("Deregistering the AMI and deleting unencrypted temporary " +
		"AMIs and snapshots")

	resp, err := ec2conn.DescribeImages(&ec2.DescribeImagesInput{
		ImageIds: []*string{&s.toDelete},
	})

	if err != nil {
		err := fmt.Errorf("Error describing AMI: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return
	}

	// Deregister image by name.
	for _, i := range resp.Images {
		_, err := ec2conn.DeregisterImage(&ec2.DeregisterImageInput{
			ImageId: i.ImageId,
		})

		if err != nil {
			err := fmt.Errorf("Error deregistering existing AMI: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return
		}
		ui.Say(fmt.Sprintf("Deregistered AMI id: %s", *i.ImageId))

		// Delete snapshot(s) by image
		for _, b := range i.BlockDeviceMappings {
			if b.Ebs != nil && aws.StringValue(b.Ebs.SnapshotId) != "" {
				_, err := ec2conn.DeleteSnapshot(&ec2.DeleteSnapshotInput{
					SnapshotId: b.Ebs.SnapshotId,
				})

				if err != nil {
					err := fmt.Errorf("Error deleting existing snapshot: %s", err)
					state.Put("error", err)
					ui.Error(err.Error())
					return
				}
				ui.Say(fmt.Sprintf("Deleted snapshot: %s", *b.Ebs.SnapshotId))
			}
		}
	}
}

func getRegionConn(config *AccessConfig, target string) (ec2iface.EC2API, error) {
	// Connect to the region where the AMI will be copied to
	session, err := config.Session()
	if err != nil {
		return nil, fmt.Errorf("Error getting region connection for copy: %s", err)
	}

	regionconn := ec2.New(session.Copy(&aws.Config{
		Region: aws.String(target),
	}))

	return regionconn, nil
}

// amiRegionCopy does a copy for the given AMI to the target region and
// returns the resulting ID and snapshot IDs, or error.
func (s *StepAMIRegionCopy) amiRegionCopy(ctx context.Context, state multistep.StateBag, config *AccessConfig, name, imageId,
	target, source, keyId string, encrypt *bool) (string, []string, error) {
	snapshotIds := []string{}

	if s.getRegionConn == nil {
		s.getRegionConn = getRegionConn
	}

	regionconn, err := s.getRegionConn(config, target)
	if err != nil {
		return "", snapshotIds, err
	}
	resp, err := regionconn.CopyImage(&ec2.CopyImageInput{
		SourceRegion:  &source,
		SourceImageId: &imageId,
		Name:          &name,
		Encrypted:     encrypt,
		KmsKeyId:      aws.String(keyId),
	})

	if err != nil {
		return "", snapshotIds, fmt.Errorf("Error Copying AMI (%s) to region (%s): %s",
			imageId, target, err)
	}

	// Wait for the image to become ready
	if err := WaitUntilAMIAvailable(ctx, regionconn, *resp.ImageId); err != nil {
		return "", snapshotIds, fmt.Errorf("Error waiting for AMI (%s) in region (%s): %s",
			*resp.ImageId, target, err)
	}

	// Getting snapshot IDs out of the copied AMI
	describeImageResp, err := regionconn.DescribeImages(&ec2.DescribeImagesInput{ImageIds: []*string{resp.ImageId}})
	if err != nil {
		return "", snapshotIds, fmt.Errorf("Error describing copied AMI (%s) in region (%s): %s",
			imageId, target, err)
	}

	for _, blockDeviceMapping := range describeImageResp.Images[0].BlockDeviceMappings {
		if blockDeviceMapping.Ebs != nil && blockDeviceMapping.Ebs.SnapshotId != nil {
			snapshotIds = append(snapshotIds, *blockDeviceMapping.Ebs.SnapshotId)
		}
	}

	return *resp.ImageId, snapshotIds, nil
}
