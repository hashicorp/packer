package vminstance

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/packer/builder/zstack/zstacktype"
	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/packer"
)

const (
	inv_str = "inventories"
)

// implement of Driver
type driverZStack struct {
	client *zClient
}

func NewDriverZStack(config Config, ui packer.Ui) (Driver, error) {
	client := Client(config.AccessKey, config.KeySecret, config.BaseUrl)
	client.ui = ui
	client.replace = config.PollReplaceStr
	client.state_timeout = config.stateTimeout
	client.create_timeout = config.createTimeout
	client.packer_tag = !config.SkipPackerSystemTag
	return &driverZStack{
		client: client,
	}, nil
}

func addSystemTags(tags []string, args ...string) []string {
	for _, v := range args {
		tags = append(tags, v)
	}
	return tags
}

func (d *driverZStack) ExportImage(image zstacktype.Image) (string, error) {
	d.client.ui.Say(fmt.Sprintf("start export image: %s", image.Uuid))

	params := map[string]interface{}{
		"backupStorageUuid": image.BackupStorage,
		"imageUuid":         image.Uuid,
	}

	result, err := d.client.SendRequest(fmt.Sprintf("/v1/backup-storage/%s/actions", image.BackupStorage), "PUT", "exportImageFromBackupStorage", params)
	if err != nil {
		return "", err
	}
	path := result["imageUrl"].(string)

	return path, nil
}

func (d *driverZStack) CreateDataVolumeFromSize(initial zstacktype.CreateDataVolume) (*zstacktype.DataVolume, error) {
	d.client.ui.Say(fmt.Sprintf("start create data volume from size: %v", initial.Size))
	params := map[string]interface{}{
		"name":               initial.Name,
		"description":        "auto created by packer-zstack",
		"primaryStorageUuid": initial.PrimaryStorage,
		"diskSize":           initial.Size,
	}

	var systemtags []string
	if d.client.packer_tag {
		systemtags = append(systemtags, "packer")
	}
	systemtags = append(systemtags, fmt.Sprintf("localStorage::hostUuid::%s", initial.Host))

	params["systemTags"] = systemtags

	result, err := d.client.Send("/v1/volumes/data", "POST", params)
	if err != nil {
		return nil, err
	}
	v := result["inventory"].(map[string]interface{})

	size, _ := v["size"].(uint64)
	volume := &zstacktype.DataVolume{
		Uuid:   fmt.Sprintf("%v", v["uuid"]),
		Name:   fmt.Sprintf("%v", v["name"]),
		Status: fmt.Sprintf("%v", v["status"]),
		Type:   fmt.Sprintf("%v", v["type"]),
		Size:   size,
	}

	return volume, err
}

func (d *driverZStack) CreateDataVolumeFromImage(initial zstacktype.CreateDataVolumeFromImage) (*zstacktype.DataVolume, error) {
	d.client.ui.Say(fmt.Sprintf("start create data volume from image: %s", initial.Uuid))
	params := map[string]interface{}{
		"name":               initial.Name,
		"description":        "auto created by packer-zstack",
		"primaryStorageUuid": initial.PrimaryStorage,
		"imageUuid":          initial.Uuid,
		"hostUuid":           initial.Host,
	}

	if d.client.packer_tag {
		systemtags := []string{"packer"}
		params["systemTags"] = systemtags
	}

	result, err := d.client.Send(fmt.Sprintf("/v1/volumes/data/from/data-volume-templates/%s", initial.Uuid), "POST", params)
	if err != nil {
		return nil, err
	}
	v := result["inventory"].(map[string]interface{})

	size, _ := v["size"].(uint64)
	volume := &zstacktype.DataVolume{
		Uuid:   fmt.Sprintf("%v", v["uuid"]),
		Name:   fmt.Sprintf("%v", v["name"]),
		Status: fmt.Sprintf("%v", v["status"]),
		Type:   fmt.Sprintf("%v", v["type"]),
		Size:   size,
	}

	return volume, err
}

func (d *driverZStack) CreateDataVolumeImage(initial zstacktype.CreateVolumeImage) (*zstacktype.Image, error) {
	d.client.ui.Say(fmt.Sprintf("start create image from dataVolume: %s", initial.DataVolume))
	params := map[string]interface{}{
		"name":               initial.Name,
		"description":        "auto created by packer-zstack",
		"volumeUuid":         initial.DataVolume,
		"system":             false,
		"backupStorageUuids": []string{initial.BackupStorage},
	}

	if d.client.packer_tag {
		systemtags := []string{"packer"}
		params["systemTags"] = systemtags
	}

	result, err := d.client.Send(fmt.Sprintf("/v1/images/data-volume-templates/from/volumes/%s", initial.DataVolume), "POST", params)
	if err != nil {
		return nil, err
	}
	i := result["inventory"].(map[string]interface{})
	image := &zstacktype.Image{
		Uuid:          fmt.Sprintf("%v", i["uuid"]),
		Name:          fmt.Sprintf("%v", i["name"]),
		Status:        fmt.Sprintf("%v", i["status"]),
		BackupStorage: initial.BackupStorage,
	}

	return image, err
}

func (d *driverZStack) CreateImage(initial zstacktype.CreateImage) (*zstacktype.Image, error) {
	d.client.ui.Say(fmt.Sprintf("start create image from rootvolume: %s", initial.RootVolume))
	params := map[string]interface{}{
		"name":               initial.Name,
		"description":        "auto created by packer-zstack",
		"guestOsType":        initial.GusetOsType,
		"rootVolumeUuid":     initial.RootVolume,
		"platform":           initial.Platform,
		"system":             false,
		"backupStorageUuids": []string{initial.BackupStorage},
	}

	if d.client.packer_tag {
		systemtags := []string{"packer"}
		params["systemTags"] = systemtags
	}

	result, err := d.client.Send(fmt.Sprintf("/v1/images/root-volume-templates/from/volumes/%s", initial.RootVolume), "POST", params)
	if err != nil {
		return nil, err
	}
	i := result["inventory"].(map[string]interface{})
	image := &zstacktype.Image{
		Uuid:          fmt.Sprintf("%v", i["uuid"]),
		Name:          fmt.Sprintf("%v", i["name"]),
		Status:        fmt.Sprintf("%v", i["status"]),
		BackupStorage: initial.BackupStorage,
	}

	return image, err
}

func (d *driverZStack) StopVmInstance(uuid string) error {
	d.client.ui.Say(fmt.Sprintf("start stop vminstance: %s", uuid))

	_, err := d.client.SendRequest(fmt.Sprintf("/v1/vm-instances/%s/actions", uuid), "PUT", "stopVmInstance", nil)

	return err
}

func (d *driverZStack) AttachDataVolume(uuid, vmUuid string) (string, error) {
	d.client.ui.Say(fmt.Sprintf("start attach data volume: %s", uuid))

	params := map[string]interface{}{
		"vmInstanceUuid": vmUuid,
		"volumeUuid":     uuid,
	}

	result, err := d.client.Send(fmt.Sprintf("/v1/volumes/%s/vm-instances/%s", uuid, vmUuid), "POST", params)
	if err != nil {
		return "", err
	}
	v := result["inventory"].(map[string]interface{})
	return fmt.Sprintf("%v", v["deviceId"]), err
}

func (d *driverZStack) DetachDataVolume(uuid string) error {
	d.client.ui.Say(fmt.Sprintf("start detach data volume: %s", uuid))

	_, err := d.client.Send(fmt.Sprintf("/v1/volumes/%s/vm-instances", uuid), "DELETE", nil)
	return err
}

func (d *driverZStack) DeleteDataVolume(uuid string) error {
	d.client.ui.Say(fmt.Sprintf("start delete data volume: %s", uuid))
	params := map[string]interface{}{
		"deleteMode": "Enforcing",
	}

	_, err := d.client.Send(fmt.Sprintf("/v1/volumes/%s", uuid), "DELETE", params)
	if err != nil {
		return err
	}
	_, err = d.client.SendRequest(fmt.Sprintf("/v1/volumes/%s/actions", uuid), "PUT", "expungeDataVolume", nil)

	return err
}

func (d *driverZStack) DeleteVmInstance(uuid string) error {
	d.client.ui.Say(fmt.Sprintf("start delete vminstance: %s", uuid))
	params := map[string]interface{}{
		"deleteMode": "Enforcing",
	}

	_, err := d.client.Send(fmt.Sprintf("/v1/vm-instances/%s", uuid), "DELETE", params)
	if err != nil {
		return err
	}
	_, err = d.client.SendRequest(fmt.Sprintf("/v1/vm-instances/%s/actions", uuid), "PUT", "expungeVmInstance", nil)

	return err
}

func getIp(inventory interface{}, l3uuid string) (ip string, err error) {
	nics, err := toSlice(inventory)
	if err != nil {
		return "", err
	}
	if len(nics) == 0 {
		return "", fmt.Errorf("no vmNics found for vm")
	}
	for _, v := range nics {
		nic := v.(map[string]interface{})
		if nic["l3NetworkUuid"].(string) == l3uuid {
			return nic["ip"].(string), nil
		}
	}
	return "", fmt.Errorf("no ip found with l3NetworkUuid[%s]", l3uuid)
}

func (d *driverZStack) CreateVmInstance(initial zstacktype.CreateVm) (*zstacktype.VmInstance, error) {
	d.client.ui.Say(fmt.Sprintf("start create vminstance from image: %s", initial.Image))
	params := map[string]interface{}{
		"imageUuid":            initial.Image,
		"instanceOfferingUuid": initial.InstanceOffering,
		"l3NetworkUuids":       []string{initial.L3},
		"name":                 initial.Name,
		"description":          "auto created by packer-zstack",
	}

	var systemtags []string
	systemtags = addSystemTags(systemtags, "cdroms::Empty::None::None", "cleanTraffic::false")
	// systemtags = addSystemTags(systemtags, "ssh::")
	if initial.Sshkey != "" {
		systemtags = addSystemTags(systemtags, fmt.Sprintf("sshkey::%s", initial.Sshkey))
	}

	if d.client.packer_tag {
		systemtags = addSystemTags(systemtags, "packer")
	}

	if initial.UserData != "" {
		userData := initial.UserData
		if userData[len(userData)-1] == '\n' {
			userData = userData[:len(userData)-1]
		}
		if _, err := base64.StdEncoding.DecodeString(userData); err != nil {
			log.Printf("[DEBUG] base64 encoding user data...")
			userData = base64.StdEncoding.EncodeToString([]byte(userData))
		}
		log.Printf("[DEBUG] userdata: %s", userData)
		systemtags = addSystemTags(systemtags, fmt.Sprintf("userdata::%s", userData))
	}

	params["systemTags"] = systemtags

	result, err := d.client.Send("/v1/vm-instances", "POST", params)
	if err != nil {
		return nil, err
	}
	v := result["inventory"].(map[string]interface{})
	ip, err := getIp(v["vmNics"], initial.L3)
	if err != nil {
		return nil, err
	}

	vm := &zstacktype.VmInstance{
		Uuid:       fmt.Sprintf("%v", v["uuid"]),
		Name:       fmt.Sprintf("%v", v["name"]),
		State:      fmt.Sprintf("%v", v["state"]),
		RootVolume: fmt.Sprintf("%v", v["rootVolumeUuid"]),
		PublicIp:   ip,
		Host:       fmt.Sprintf("%v", v["hostUuid"]),
	}

	return vm, err
}

func (d *driverZStack) QueryZone(uuid string) (zone *zstacktype.Zone, err error) {
	rsp, err := d.client.Query("/v1/zones", map[string]string{"uuid": uuid})
	if err != nil {
		return nil, err
	}
	inventories, err := toSlice(rsp[inv_str])
	if err != nil {
		return nil, err
	}
	if len(inventories) == 0 {
		return nil, nil
	}
	z := inventories[0].(map[string]interface{})
	zone = &zstacktype.Zone{
		Name: fmt.Sprintf("%v", z["name"]),
		Uuid: fmt.Sprintf("%v", z["uuid"]),
	}
	return
}

func (d *driverZStack) QueryVolume(uuid string) (volume *zstacktype.DataVolume, err error) {
	rsp, err := d.client.Query("/v1/volumes", map[string]string{"uuid": uuid})
	if err != nil {
		return nil, err
	}
	inventories, err := toSlice(rsp[inv_str])
	if err != nil {
		return nil, err
	}
	if len(inventories) == 0 {
		return nil, nil
	}
	v := inventories[0].(map[string]interface{})
	size, _ := v["size"].(uint64)
	volume = &zstacktype.DataVolume{
		Name:           fmt.Sprintf("%v", v["name"]),
		Uuid:           fmt.Sprintf("%v", v["uuid"]),
		Status:         fmt.Sprintf("%v", v["status"]),
		Size:           size,
		Type:           fmt.Sprintf("%v", v["type"]),
		DeviceId:       fmt.Sprintf("%v", v["deviceId"]),
		PrimaryStorage: fmt.Sprintf("%v", v["primaryStorageUuid"]),
	}
	return
}

func getCidr(ranges interface{}) (cidr []string) {
	if ranges == nil {
		return nil
	}
	var cidrs []string
	rs, _ := toSlice(ranges)
	for _, r := range rs {
		cidrs = append(cidrs, r.(map[string]interface{})["networkCidr"].(string))
	}
	return cidrs
}

func (d *driverZStack) QueryL3Network(uuid string) (l3 *zstacktype.L3Network, err error) {
	rsp, err := d.client.Query("/v1/l3-networks", map[string]string{"uuid": uuid})
	if err != nil {
		return nil, err
	}
	inventories, err := toSlice(rsp[inv_str])
	if err != nil {
		return nil, err
	}
	if len(inventories) == 0 {
		return nil, nil
	}
	n := inventories[0].(map[string]interface{})
	l3 = &zstacktype.L3Network{
		Name:     fmt.Sprintf("%v", n["name"]),
		Uuid:     fmt.Sprintf("%v", n["uuid"]),
		Category: fmt.Sprintf("%v", n["category"]),
		Cidr:     getCidr(n["ipRanges"]),
	}
	return
}

func (d *driverZStack) QueryImage(uuid string) (image *zstacktype.Image, err error) {
	rsp, err := d.client.Query("/v1/images", map[string]string{"uuid": uuid})
	if err != nil {
		return nil, err
	}
	inventories, err := toSlice(rsp[inv_str])
	if err != nil {
		return nil, err
	}
	if len(inventories) == 0 {
		return nil, nil
	}
	i := inventories[0].(map[string]interface{})
	bss, err := toSlice(i["backupStorageRefs"])
	bs := bss[0].(map[string]interface{})

	image = &zstacktype.Image{
		Name:          fmt.Sprintf("%v", i["name"]),
		Uuid:          fmt.Sprintf("%v", i["uuid"]),
		BackupStorage: fmt.Sprintf("%v", bs["backupStorageUuid"]),
		Type:          fmt.Sprintf("%v", i["type"]),
		OSType:        fmt.Sprintf("%v", i["guestOsType"]),
		Platform:      fmt.Sprintf("%v", i["platform"]),
		Status:        fmt.Sprintf("%v", i["status"]),
	}
	return
}

func (d *driverZStack) QueryBackupStorage(uuid string) (bs *zstacktype.BackupStorage, err error) {
	rsp, err := d.client.Query("/v1/backup-storage", map[string]string{"uuid": uuid})
	if err != nil {
		return nil, err
	}
	inventories, err := toSlice(rsp[inv_str])
	if err != nil {
		return nil, err
	}
	if len(inventories) == 0 {
		return nil, nil
	}
	b := inventories[0].(map[string]interface{})

	zones, err := toSlice(b["attachedZoneUuids"])
	var zoneUuids []string
	for _, zone := range zones {
		zoneUuids = append(zoneUuids, fmt.Sprintf("%v", zone))
	}

	bs = &zstacktype.BackupStorage{
		Name: fmt.Sprintf("%v", b["name"]),
		Uuid: fmt.Sprintf("%v", b["uuid"]),
		Zone: zoneUuids,
	}
	return
}

func (d *driverZStack) QueryVm(uuid string) (vm *zstacktype.VmInstance, err error) {
	rsp, err := d.client.Query("/v1/vm-instances", map[string]string{"uuid": uuid})
	if err != nil {
		return nil, err
	}
	inventories, err := toSlice(rsp[inv_str])
	if err != nil {
		return nil, err
	}
	if len(inventories) == 0 {
		return nil, nil
	}
	v := inventories[0].(map[string]interface{})
	vm = &zstacktype.VmInstance{
		Name:  fmt.Sprintf("%v", v["name"]),
		Uuid:  fmt.Sprintf("%v", v["uuid"]),
		State: fmt.Sprintf("%v", v["state"]),
	}
	return
}

func (d *driverZStack) WaitForInstance(state, uuid, instancetype string) <-chan error {
	d.client.ui.Say(fmt.Sprintf("wait %s: %s to %s status", instancetype, uuid, state))
	errCh := make(chan error, 1)
	switch instancetype {
	case "vm":
		go waitForState(errCh, d.checkVm, state, uuid)
	case "image":
		go waitForState(errCh, d.checkImage, state, uuid)
	case "volume":
		go waitForState(errCh, d.checkVolume, state, uuid)
	default:
		errCh <- fmt.Errorf("not valid instancetype: %s", instancetype)
	}

	return errCh
}

func (d *driverZStack) checkVolume(target, uuid string) error {
	volume, err := d.QueryVolume(uuid)
	if err != nil {
		return err
	}
	if volume.Status == target {
		return nil
	}
	return fmt.Errorf("retrying for datavolume state %s, got %s", target, volume.Status)
}

func (d *driverZStack) checkImage(target, uuid string) error {
	image, err := d.QueryImage(uuid)
	if err != nil {
		return err
	}
	if image.Status == target {
		return nil
	}
	return fmt.Errorf("retrying for image state %s, got %s", target, image.Status)
}

func (d *driverZStack) checkVm(target, uuid string) error {
	vm, err := d.QueryVm(uuid)
	if err != nil {
		return err
	} else if vm.State == target {
		return nil
	}
	return fmt.Errorf("retrying for vm state %s, got %s", target, vm.State)
}

func waitForState(errCh chan<- error, f func(string, string) error, target, uuid string) {
	ctx := context.TODO()
	err := retry.Config{
		RetryDelay: (&retry.Backoff{InitialBackoff: 100 * time.Millisecond, MaxBackoff: 100 * time.Millisecond, Multiplier: 1}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		return f(target, uuid)
	})
	errCh <- err
	return
}

func (d *driverZStack) GetZStackVersion() (string, error) {
	rsp, err := d.client.SendRequest("/v1/management-nodes/actions", "PUT", "getVersion", nil)
	if err != nil {
		return "", err
	}
	return rsp["version"].(string), err
}
