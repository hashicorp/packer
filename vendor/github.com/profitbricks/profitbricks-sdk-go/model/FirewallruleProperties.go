package model



type FirewallruleProperties struct {
    Name  string  `json:"name,omitempty"`
    Protocol  string  `json:"protocol,omitempty"`
    SourceMac  string  `json:"sourceMac,omitempty"`
    SourceIp  string  `json:"sourceIp,omitempty"`
    TargetIp  string  `json:"targetIp,omitempty"`
    IcmpCode  int32  `json:"icmpCode,omitempty"`
    IcmpType  int32  `json:"icmpType,omitempty"`
    PortRangeStart  int32  `json:"portRangeStart,omitempty"`
    PortRangeEnd  int32  `json:"portRangeEnd,omitempty"`
    
}
