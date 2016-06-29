package profitbricks

import "encoding/json"

type CreateLanRequest struct {
	LanProperties `json:"properties"`
}

type LanProperties struct {
	Name   string   `json:"name,omitempty"`
	Public bool     `json:"public,omitempty"`
	Nics   []string `json:"nics,omitempty"`
}

// ListLan returns a Collection for lans in the Datacenter
func ListLans(dcid string) Collection {
	path := lan_col_path(dcid)
	return is_list(path)
}

// CreateLan creates a lan in the datacenter
// from a jason []byte and returns a Instance struct
func CreateLan(dcid string, request CreateLanRequest) Instance {
	obj, _ := json.Marshal(request)
	path := lan_col_path(dcid)
	return is_post(path, obj)
}

// GetLan pulls data for the lan where id = lanid returns an Instance struct
func GetLan(dcid, lanid string) Instance {
	path := lan_path(dcid, lanid)
	return is_get(path)
}

// PatchLan does a partial update to a lan using json from []byte jason
// returns a Instance struct
func PatchLan(dcid string, lanid string, obj map[string]string) Instance {
	jason := []byte(MkJson(obj))
	path := lan_path(dcid, lanid)
	return is_patch(path, jason)
}

// DeleteLan deletes a lan where id == lanid
func DeleteLan(dcid, lanid string) Resp {
	path := lan_path(dcid, lanid)
	return is_delete(path)
}

// ListLanMembers returns a Nic struct collection for the Lan
func ListLanMembers(dcid, lanid string) Collection {
	path := lan_nic_col(dcid, lanid)
	return is_list(path)
}
