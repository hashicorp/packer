package model

type Server struct {
	Id         string  `json:"id,omitempty"`
	Type_      string  `json:"type,omitempty"`
	Href       string  `json:"href,omitempty"`
	//Metadata  DatacenterElementMetadata  `json:"metadata,omitempty"`
	Properties ServerProperties  `json:"properties,omitempty"`
	Entities   ServerEntities  `json:"entities,omitempty"`
}
