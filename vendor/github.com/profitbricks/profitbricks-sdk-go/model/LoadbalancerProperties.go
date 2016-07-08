package model



type LoadbalancerProperties struct {
    Name  string  `json:"name,omitempty"`
    Ip  string  `json:"ip,omitempty"`
    Dhcp  bool  `json:"dhcp,omitempty"`
    
}
