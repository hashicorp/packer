package profitbricks

import "encoding/json"

type IPBlockReserveRequest struct {
	IPBlockProperties `json:"properties"`
}

type IPBlockProperties struct {
	Size     int    `json:"size,omitempty"`
	Location string `json:"location,omitempty"`
}

// ListIpBlocks
func ListIpBlocks() Collection {
	path := ipblock_col_path()
	return is_list(path)
}

func ReserveIpBlock(request IPBlockReserveRequest) Instance {
	obj, _ := json.Marshal(request)
	path := ipblock_col_path()
	return is_post(path, obj)

}
func GetIpBlock(ipblockid string) Instance {
	path := ipblock_path(ipblockid)
	return is_get(path)
}

func ReleaseIpBlock(ipblockid string) Resp {
	path := ipblock_path(ipblockid)
	return is_delete(path)
}
