package model

type Volume struct {
	Id         string  `json:"id,omitempty"`
	Type_      string  `json:"type,omitempty"`
	Href       string  `json:"href,omitempty"`
	Properties VolumeProperties  `json:"properties,omitempty"`
}
