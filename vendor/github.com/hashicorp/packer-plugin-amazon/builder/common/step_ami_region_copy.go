package common

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
)

type StepAMIRegionCopy struct {
	AccessConfig      *AccessConfig
	Regions           []string
	AMIKmsKeyId       string
	RegionKeyIds      map[string]string
	EncryptBootVolume config.Trilean // nil means preserve
	Name              string
	OriginalRegion    string

	toDelete           string
	getRegionConn      func(*AccessConfig, string) (ec2iface.EC2API, error)
	AMISkipCreateImage bool
	AMISkipBuildRegion bool
}

func (s *StepAMIRegionCopy) DeduplicateRegions(intermediary bool) {
	// Deduplicates regions by looping over the list of regions and storing
	// the regions as keys in a map. This saves users from accidentally copying
	// regions twice if they've added a region to a map twice.

	RegionMap := map[string]bool{}
	RegionSlice := []string{}

	// Original build region may or may not be present in the Regions list, so
	// let's make absolutely sure it's in our map.
	RegionMap[s.OriginalRegion] = true
	for _, r := range s.Regions {
		RegionMap[r] = true
	}

	if !intermediary || s.AMISkipBuildRegion {
		// We don't want to copy back into the original region if we aren't
		// using an intermediary image, so remove the original region from our
		// map.

		// We also don't want to copy back into the original region if the
		// intermediary image is because we're skipping the build region.
		delete(RegionMap, s.OriginalRegion)

	}

	// Now print all those keys into the region slice again
	for k := range RegionMap {
		RegionSlice = append(RegionSlice, k)
	}

	s.Regions = RegionSlice
}

func (s *StepAMIRegionCopy) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	if s.AMISkipCreateImage {
		ui.Say("Skipping AMI region copy...")
		return multistep.ActionContinue
	}

	amis := state.Get("amis").(map[string]string)
	snapshots := state.Get("snapshots").(map[string][]string)
	intermediary, _ := state.Get("intermediary_image").(bool)

	s.DeduplicateRegions(intermediary)
	ami := amis[s.OriginalRegion]

	// Make a note to delete the intermediary AMI if necessary.
	if intermediary {
		s.toDelete = ami
	}

	if s.EncryptBootVolume.True() {
		// encrypt_boot is true, so we have to copy the temporary
		// AMI with required encryption setting.
		// temp image was created by stepCreateAMI.
		if s.RegionKeyIds == nil {
			s.RegionKeyIds = make(map[string]string)
		}

		// Make sure the kms_key_id for the original region is in the map, as
		// long as the AMIKmsKeyId isn't being defaulted.
		if s.AMIKmsKeyId != "" {
			if _, ok := s.RegionKeyIds[s.OriginalRegion]; !ok {
				s.RegionKeyIds[s.OriginalRegion] = s.AMIKmsKeyId
			}
		} else {
			if regionKey, ok := s.RegionKeyIds[s.OriginalRegion]; ok {
				s.AMIKmsKeyId = regionKey
			}
		}
	}

	if len(s.Regions) == 0 {
		return multistep.ActionContinue
	}

	ui.Say(fmt.Sprintf("Copying/Encrypting AMI (%s) to other regions...", ami))

	var lock sync.Mutex
	var wg sync.WaitGroup
	errs := new(packersdk.MultiError)
	wg.Add(len(s.Regions))
	for _, region := range s.Regions {
		var regKeyID string
		ui.Message(fmt.Sprintf("Copying to: %s", region))

		if s.EncryptBootVolume.True() {
			// Encrypt is true, explicitly
			regKeyID = s.RegionKeyIds[region]
		} else {
			// Encrypt is nil or false; Make sure region key is empty
			regKeyID = ""
		}

		go func(region string) {
			defer wg.Done()
			id, snapshotIds, err := s.amiRegionCopy(ctx, state, s.AccessConfig,
				s.Name, ami, region, s.OriginalRegion, regKeyID,
				s.EncryptBootVolume.ToBoolPointer())
			lock.Lock()
			defer lock.Unlock()
			amis[region] = id
			snapshots[region] = snapshotIds
			if err != nil {
				errs = packersdk.MultiErrorAppend(errs, err)
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
	ui := state.Get("ui").(packersdk.Ui)

	if len(s.toDelete) == 0 {
		return
	}

	// Delete the unencrypted amis and snapshots
	ui.Say("Deregistering the AMI and deleting unencrypted temporary " +
		"AMIs and snapshots")
	err := DestroyAMIs([]*string{&s.toDelete}, ec2conn)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return
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
	if err := s.AccessConfig.PollingConfig.WaitUntilAMIAvailable(ctx, regionconn, *resp.ImageId); err != nil {
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
