package vminstance

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/packer/builder/zstack/zstacktype"

	"github.com/hashicorp/packer/helper/multistep"
)

type StepPreValidate struct {
}

func (s *StepPreValidate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	_, _, ui := GetCommonFromState(state)
	ui.Say("start pre validate zstack...")

	var zone *zstacktype.Zone
	var image *zstacktype.Image
	var bs *zstacktype.BackupStorage
	var err error

	if err = validateZStackVersion(state); err != nil {
		return halt(state, err, "")
	}

	if zone, err = validateZone(state); err != nil {
		return halt(state, err, "")
	}
	if _, err = validateL3Network(state); err != nil {
		return halt(state, err, "")
	}

	if image, err = validateImage(state); err != nil {
		return halt(state, err, "")
	}

	if bs, err = validateBackupStorage(image.BackupStorage, state); err != nil {
		return halt(state, err, "")
	}

	contains := false
	for _, v := range bs.Zone {
		if v == zone.Uuid {
			contains = true
		}
	}

	if !contains {
		return halt(state, fmt.Errorf("zone or image is null or image's backupStorage is not in the zone"), "")
	}

	return multistep.ActionContinue
}

func validateImage(state multistep.StateBag) (*zstacktype.Image, error) {
	driver, config, _ := GetCommonFromState(state)

	image, err := driver.QueryImage(config.Image)
	if err != nil {
		return nil, err
	} else if image == nil {
		return nil, fmt.Errorf("cannot find image from uuid: %s", config.Image)
	}
	state.Put(GuestOSType, image.OSType)
	state.Put(Platform, image.Platform)

	return image, nil
}

func validateL3Network(state multistep.StateBag) (*zstacktype.L3Network, error) {
	driver, config, _ := GetCommonFromState(state)

	l3, err := driver.QueryL3Network(config.L3Network)
	if err != nil {
		return nil, err
	} else if l3 == nil {
		return nil, fmt.Errorf("cannot find l3network from uuid: %s", config.Zone)
	}

	if l3.Category != "Public" {
		return nil, fmt.Errorf("Packer need Public L3Network to SSH, but got: %s", l3.Category)
	}

	return l3, nil
}

func validateZone(state multistep.StateBag) (*zstacktype.Zone, error) {
	driver, config, _ := GetCommonFromState(state)

	zone, err := driver.QueryZone(config.Zone)
	if err != nil {
		return nil, err
	} else if zone == nil {
		return nil, fmt.Errorf("cannot find zone from uuid: %s", config.Zone)
	}

	return zone, nil
}

func validateBackupStorage(uuid string, state multistep.StateBag) (*zstacktype.BackupStorage, error) {
	driver, _, _ := GetCommonFromState(state)

	bs, err := driver.QueryBackupStorage(uuid)
	if err != nil {
		return nil, err
	} else if bs == nil {
		return nil, fmt.Errorf("cannot find backupStorage from uuid: %s", uuid)
	}

	state.Put(BackupStorage, bs.Uuid)

	return bs, nil
}

func versionLessThan(v1, v2 string) bool {
	r1 := strings.Split(v1, ".")
	r2 := strings.Split(v2, ".")

	for k, v := range r2 {
		if k >= len(r1) {
			break
		}
		v1, _ := strconv.Atoi(r1[k])
		v2, _ := strconv.Atoi(v)
		if v1 > v2 {
			return false
		} else if v1 < v2 {
			return true
		}
	}
	return false
}

func validateZStackVersion(state multistep.StateBag) (err error) {
	driver, _, _ := GetCommonFromState(state)
	version, err := driver.GetZStackVersion()
	if err != nil {
		return
	}
	if version == "" {
		return fmt.Errorf("got empty version from api return")
	}
	if versionLessThan(version, "3.4.0") {
		return fmt.Errorf("zstack-packer only works on zstack version greater or equal to 3.4.0, current version is: %s", version)
	}
	return
}

func (s *StepPreValidate) Cleanup(state multistep.StateBag) {
	_, _, ui := GetCommonFromState(state)
	ui.Say("cleanup prevalidate executing...")
}
