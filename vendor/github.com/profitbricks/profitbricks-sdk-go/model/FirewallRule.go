package model



type FirewallRule struct {
    Id  string  `json:"id,omitempty"`
    Type_  string  `json:"type,omitempty"`
    Href  string  `json:"href,omitempty"`
    Metadata  DatacenterElementMetadata  `json:"metadata,omitempty"`
    Properties  FirewallruleProperties  `json:"properties,omitempty"`
    
}
