package model



type Loadbalancer struct {
    Id  string  `json:"id,omitempty"`
    Type_  string  `json:"type,omitempty"`
    Href  string  `json:"href,omitempty"`
    Metadata  DatacenterElementMetadata  `json:"metadata,omitempty"`
    Properties  LoadbalancerProperties  `json:"properties,omitempty"`
    Entities  LoadbalancerEntities  `json:"entities,omitempty"`
    
}
