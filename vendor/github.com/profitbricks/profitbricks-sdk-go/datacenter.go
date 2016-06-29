package profitbricks

import "encoding/json"

type CreateDatacenterRequest struct {
	DCProperties `json:"properties"`
}

type DCProperties struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Location    string `json:"location,omitempty"`
}

// ListDatacenters returns a Collection struct
func ListDatacenters() Collection {
	path := dc_col_path()
	return is_list(path)
}

// CreateDatacenter creates a datacenter and returns a Instance struct
func CreateDatacenter(dc CreateDatacenterRequest) Instance {
	obj, _ := json.Marshal(dc)
	path := dc_col_path()
	return is_post(path, obj)
}

// GetDatacenter returns a Instance struct where id == dcid
func GetDatacenter(dcid string) Instance {
	path := dc_path(dcid)
	return is_get(path)
}

// PatchDatacenter replaces any Datacenter properties with the values in jason
//returns an Instance struct where id ==dcid
func PatchDatacenter(dcid string, obj map[string]string) Instance {
	jason_patch := []byte(MkJson(obj))
	path := dc_path(dcid)
	return is_patch(path, jason_patch)
}

// Deletes a Datacenter where id==dcid
func DeleteDatacenter(dcid string) Resp {
	path := dc_path(dcid)
	return is_delete(path)
}
