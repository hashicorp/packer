package model



type NicProperties struct {
    Name  string  `json:"name,omitempty"`
    Mac  string  `json:"mac,omitempty"`
    Ips  []string  `json:"ips,omitempty"`
    Dhcp  bool  `json:"dhcp,omitempty"`
    Lan  string  `json:"lan,omitempty"`
    FirewallActive  bool  `json:"firewallActive,omitempty"`
    
}
