package model



type Loadbalancers struct {
    Id  string  `json:"id,omitempty"`
    Type_  string  `json:"type,omitempty"`
    Href  string  `json:"href,omitempty"`
    Items  []Loadbalancer  `json:"items,omitempty"`
    
}
