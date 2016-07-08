package model

type VolumeProperties struct {
	Name                string  `json:"name,omitempty"`
	Type_               string  `json:"type,omitempty"`
	Size                int  `json:"size,omitempty"`
	Image               string  `json:"image,omitempty"`
	ImagePassword       string  `json:"imagePassword,omitempty"`
	SshKeys             []string `json:"sshKeys,omitempty"`
	Bus                 string  `json:"bus,omitempty"`
	LicenceType         string  `json:"licenceType,omitempty"`
	CpuHotPlug          bool  `json:"cpuHotPlug,omitempty"`
	CpuHotUnplug        bool  `json:"cpuHotUnplug,omitempty"`
	RamHotPlug          bool  `json:"ramHotPlug,omitempty"`
	RamHotUnplug        bool  `json:"ramHotUnplug,omitempty"`
	NicHotPlug          bool  `json:"nicHotPlug,omitempty"`
	NicHotUnplug        bool  `json:"nicHotUnplug,omitempty"`
	DiscVirtioHotPlug   bool  `json:"discVirtioHotPlug,omitempty"`
	DiscVirtioHotUnplug bool  `json:"discVirtioHotUnplug,omitempty"`
	DiscScsiHotPlug     bool  `json:"discScsiHotPlug,omitempty"`
	DiscScsiHotUnplug   bool  `json:"discScsiHotUnplug,omitempty"`
	DeviceNumber        int64  `json:"deviceNumber,omitempty"`
}
