package model



type Lan struct {
    Id  string  `json:"id,omitempty"`
    Type_  string  `json:"type,omitempty"`
    Href  string  `json:"href,omitempty"`
    Metadata  DatacenterElementMetadata  `json:"metadata,omitempty"`
    Properties  LanProperties  `json:"properties,omitempty"`
    Entities  LanEntities  `json:"entities,omitempty"`
    
}
