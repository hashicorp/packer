package metadata

type Metadata struct {
	AvailabilityZone string          `json:"availability-zone"`
	CloudName        string          `json:"cloud-name"`
	InstanceId       string          `json:"instance-id"`
	LocalHostname    string          `json:"local-hostname"`
	NetworkConfig    MDNetworkConfig `json:"network-config"`
	Platform         string          `json:"platform"`
	PublicSSHKeys    []string        `json:"public-ssh-keys"`
	Region           string          `json:"region"`
	UHost            MDUHost         `json:"uhost"`
}

type MDMatch struct {
	MacAddress string `json:"macaddress"`
}

type MDNameServers struct {
	Addresses []string `json:"addresses"`
}

type MDEthernet struct {
	Addresses   []string      `json:"addresses"`
	Gateway4    string        `json:"gateway4"`
	Match       MDMatch       `json:"match"`
	MTU         int           `json:"mtu"`
	NameServers MDNameServers `json:"nameservers"`
}

type MDNetworkConfig struct {
	Ethernets map[string]MDEthernet `json:"ethernets"`
	Version   int                   `json:"version"`
}

type MDDisks struct {
	BackupType string `json:"backup-type"`
	DiskId     string `json:"disk-id"`
	DiskType   string `json:"disk-type"`
	Drive      string `json:"drive"`
	Encrypted  bool   `json:"encrypted"`
	IsBoot     bool   `json:"is-boot"`
	Name       string `json:"name"`
	Size       int    `json:"size"`
}

type MDIPs struct {
	IPAddress string `json:"ip-address"`
	Type      string `json:"type"`
	Bandwidth int    `json:"bandwidth,omitempty"`
	IPId      string `json:"ip-id,omitempty"`
}

type MDNetworkInterfaces struct {
	IPs      []MDIPs `json:"ips"`
	Mac      string  `json:"mac"`
	SubnetId string  `json:"subnet-id"`
	VpcId    string  `json:"vpc-id"`
}

type MDUHost struct {
	CPU               int                   `json:"cpu"`
	Disks             []MDDisks             `json:"disks"`
	GPU               int                   `json:"gpu"`
	Hotplug           bool                  `json:"hotplug"`
	ImageId           string                `json:"image-id"`
	IsolationGroup    string                `json:"isolation-group"`
	MachineType       string                `json:"machine-type"`
	Memory            int                   `json:"memory"`
	Name              string                `json:"name"`
	NetCapability     string                `json:"net-capability"`
	NetworkInterfaces []MDNetworkInterfaces `json:"network-interfaces"`
	OsName            string                `json:"os-name"`
	ProjectId         string                `json:"project-id"`
	Region            string                `json:"region"`
	Remark            string                `json:"remark"`
	Tag               string                `json:"tag"`
	UHostId           string                `json:"uhost-id"`
	Zone              string                `json:"zone"`
}
