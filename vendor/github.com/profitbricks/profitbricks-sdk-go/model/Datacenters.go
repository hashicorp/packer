package model



type Datacenters struct {
    Id  string  `json:"id,omitempty"`
    Type_  string  `json:"type,omitempty"`
    Href  string  `json:"href,omitempty"`
    Items  []Datacenter  `json:"items,omitempty"`
    
}
