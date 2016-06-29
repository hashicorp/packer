package profitbricks

import (
	"encoding/json"
)

type NicCreateRequest struct {
	NicProperties `json:"properties"`
}

type NicProperties struct {
	Name string   `json:"name,omitempty"`
	Ips  []string `json:"ips,omitempty"`
	Dhcp bool     `json:"dhcp"`
	Lan  string   `json:"lan"`
}

// ListNics returns a Nics struct collection
func ListNics(dcid, srvid string) Collection {
	path := nic_col_path(dcid, srvid)
	return is_list(path)
}

// CreateNic creates a nic on a server
// from a jason []byte and returns a Instance struct
func CreateNic(dcid string, srvid string, request NicCreateRequest) Instance {
	obj, _ := json.Marshal(request)
	path := nic_col_path(dcid, srvid)
	return is_post(path, obj)
}

// GetNic pulls data for the nic where id = srvid returns a Instance struct
func GetNic(dcid, srvid, nicid string) Instance {
	path := nic_path(dcid, srvid, nicid)
	return is_get(path)
}

// PatchNic partial update of nic properties passed in as jason []byte
// Returns Instance struct
func PatchNic(dcid string, srvid string, nicid string, obj map[string]string) Instance {
	jason := []byte(MkJson(obj))
	path := nic_path(dcid, srvid, nicid)
	return is_patch(path, jason)
}

// DeleteNic deletes the nic where id=nicid and returns a Resp struct
func DeleteNic(dcid, srvid, nicid string) Resp {
	path := nic_path(dcid, srvid, nicid)
	return is_delete(path)
}
