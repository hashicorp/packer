package model



type ImageProperties struct {
    Name  string  `json:"name,omitempty"`
    Description  string  `json:"description,omitempty"`
    Location  string  `json:"location,omitempty"`
    Size  int  `json:"size,omitempty"`
    CpuHotPlug  bool  `json:"cpuHotPlug,omitempty"`
    CpuHotUnplug  bool  `json:"cpuHotUnplug,omitempty"`
    RamHotPlug  bool  `json:"ramHotPlug,omitempty"`
    RamHotUnplug  bool  `json:"ramHotUnplug,omitempty"`
    NicHotPlug  bool  `json:"nicHotPlug,omitempty"`
    NicHotUnplug  bool  `json:"nicHotUnplug,omitempty"`
    DiscVirtioHotPlug  bool  `json:"discVirtioHotPlug,omitempty"`
    DiscVirtioHotUnplug  bool  `json:"discVirtioHotUnplug,omitempty"`
    DiscScsiHotPlug  bool  `json:"discScsiHotPlug,omitempty"`
    DiscScsiHotUnplug  bool  `json:"discScsiHotUnplug,omitempty"`
    LicenceType  string  `json:"licenceType,omitempty"`
    ImageType  string  `json:"imageType,omitempty"`
    Public  bool  `json:"public,omitempty"`
    
}
