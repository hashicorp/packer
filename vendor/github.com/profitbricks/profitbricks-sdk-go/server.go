package profitbricks

import "encoding/json"

type CreateServerRequest struct {
	ServerProperties `json:"properties"`
}

type ServerProperties struct {
	Name             string    `json:"name,omitempty"`
	Ram              int       `json:"ram,omitempty"`
	Cores            int       `json:"cores,omitempty"`
	Availabilityzone string    `json:"availabilityzone,omitempty"`
	Licencetype      string    `json:"licencetype,omitempty"`
	BootVolume       *Instance `json:"bootVolume,omitempty"`
	BootCdrom        *Instance `json:"bootCdrom,omitempty"`
}

// ListServers returns a server struct collection
func ListServers(dcid string) Collection {
	path := server_col_path(dcid)
	return is_list(path)
}

// CreateServer creates a server from a jason []byte and returns a Instance struct
func CreateServer(dcid string, req CreateServerRequest) Instance {
	jason, _ := json.Marshal(req)
	path := server_col_path(dcid)
	return is_post(path, jason)
}

// GetServer pulls data for the server where id = srvid returns a Instance struct
func GetServer(dcid, srvid string) Instance {
	path := server_path(dcid, srvid)
	return is_get(path)
}

// PatchServer partial update of server properties passed in as jason []byte
// Returns Instance struct
func PatchServer(dcid string, srvid string, req ServerProperties) Instance {
	jason, _ := json.Marshal(req)
	path := server_path(dcid, srvid)
	return is_patch(path, jason)
}

// DeleteServer deletes the server where id=srvid and returns Resp struct
func DeleteServer(dcid, srvid string) Resp {
	path := server_path(dcid, srvid)
	return is_delete(path)
}

func ListAttachedCdroms(dcid, srvid string) Collection {
	path := server_cdrom_col_path(dcid, srvid)
	return is_list(path)
}

func AttachCdrom(dcid string, srvid string, cdid string) Instance {
	jason := []byte(`{"id":"` + cdid + `"}`)
	path := server_cdrom_col_path(dcid, srvid)
	return is_post(path, jason)
}

func GetAttachedCdrom(dcid, srvid, cdid string) Instance {
	path := server_cdrom_path(dcid, srvid, cdid)
	return is_get(path)
}

func DetachCdrom(dcid, srvid, cdid string) Resp {
	path := server_cdrom_path(dcid, srvid, cdid)
	return is_delete(path)
}

func ListAttachedVolumes(dcid, srvid string) Collection {
	path := server_volume_col_path(dcid, srvid)
	return is_list(path)
}

func AttachVolume(dcid string, srvid string, volid string) Instance {
	jason := []byte(`{"id":"` + volid + `"}`)
	path := server_volume_col_path(dcid, srvid)
	return is_post(path, jason)
}

func GetAttachedVolume(dcid, srvid, volid string) Instance {
	path := server_volume_path(dcid, srvid, volid)
	return is_get(path)
}

func DetachVolume(dcid, srvid, volid string) Resp {
	path := server_volume_path(dcid, srvid, volid)
	return is_delete(path)
}

// server_command is a generic function for running server commands
func server_command(dcid, srvid, cmd string) Resp {
	jason := `
		{}
		`
	path := server_command_path(dcid, srvid, cmd)
	return is_command(path, jason)
}

// StartServer starts a server
func StartServer(dcid, srvid string) Resp {
	return server_command(dcid, srvid, "start")
}

// StopServer stops a server
func StopServer(dcid, srvid string) Resp {
	return server_command(dcid, srvid, "stop")
}

// RebootServer reboots a server
func RebootServer(dcid, srvid string) Resp {
	return server_command(dcid, srvid, "reboot")
}
