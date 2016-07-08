package model

type ServerProperties struct {
	Name             string  `json:"name,omitempty"`
	Cores            int  `json:"cores,omitempty"`
	Ram              int  `json:"ram,omitempty"`
	AvailabilityZone string  `json:"availabilityZone,omitempty"`
	VmState          string  `json:"vmState,omitempty"`
}
