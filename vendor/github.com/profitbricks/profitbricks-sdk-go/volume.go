package profitbricks

import "encoding/json"

type CreateVolumeRequest struct {
	VolumeProperties `json:"properties"`
}

type VolumeProperties struct {
	Name          string   `json:"name,omitempty"`
	Size          int      `json:"size,omitempty"`
	Bus           string   `json:",bus,omitempty"`
	Image         string   `json:"image,omitempty"`
	Type          string   `json:"type,omitempty"`
	LicenceType   string   `json:"licenceType,omitempty"`
	ImagePassword string   `json:"imagePassword,omitempty"`
	SshKey        []string `json:"sshKeys,omitempty"`
}

// ListVolumes returns a Collection struct for volumes in the Datacenter
func ListVolumes(dcid string) Collection {
	path := volume_col_path(dcid)
	return is_list(path)
}

func GetVolume(dcid string, volumeId string) Instance {
	path := volume_path(dcid, volumeId)
	return is_get(path)
}

func PatchVolume(dcid string, volid string, request VolumeProperties) Instance {
	obj, _ := json.Marshal(request)
	path := volume_path(dcid, volid)
	return is_patch(path, obj)
}

func CreateVolume(dcid string, request CreateVolumeRequest) Instance {
	obj, _ := json.Marshal(request)
	path := volume_col_path(dcid)
	return is_post(path, obj)
}

func DeleteVolume(dcid, volid string) Resp {
	path := volume_path(dcid, volid)
	return is_delete(path)
}

func CreateSnapshot(dcid string, volid string, name string) Resp {
	var path = volume_path(dcid, volid)
	path = path + "/create-snapshot"

	return is_command(path, "name=" + name)
}

func RestoreSnapshot(dcid string, volid string, snapshotId string) Resp {
	var path = volume_path(dcid, volid)
	path = path + "/restore-snapshot"

	return is_command(path, "snapshotId=" + snapshotId)
}
