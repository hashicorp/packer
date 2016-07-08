package model



type FirewallRules struct {
    Id  string  `json:"id,omitempty"`
    Type_  string  `json:"type,omitempty"`
    Href  string  `json:"href,omitempty"`
    Items  []FirewallRule  `json:"items,omitempty"`
    
}
