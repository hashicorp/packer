package model

type Nic struct {
	Id    string `json:"id,omitempty"`
	Type_ string `json:"type,omitempty"`
	Href  string `json:"href,omitempty"`
	// Metadata  DatacenterElementMetadata  `json:"metadata,omitempty"`
	Properties NicProperties `json:"properties,omitempty"`
	Entities   *NicEntities  `json:"entities,omitempty"`
}
