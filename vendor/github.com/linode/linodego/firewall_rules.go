package linodego

// NetworkProtocol enum type
type NetworkProtocol string

// NetworkProtocol enum values
const (
	TCP  NetworkProtocol = "TCP"
	UDP  NetworkProtocol = "UDP"
	ICMP NetworkProtocol = "ALL"
)

// NetworkAddresses are arrays of ipv4 and v6 addresses
type NetworkAddresses struct {
	IPv4 []string `json:"ipv4"`
	IPv6 []string `json:"ipv6"`
}

// A FirewallRule is a whitelist of ports, protocols, and addresses for which traffic should be allowed.
type FirewallRule struct {
	Ports     string           `json:"ports"`
	Protocol  NetworkProtocol  `json:"protocol"`
	Addresses NetworkAddresses `json:"addresses"`
}

// FirewallRuleSet is a pair of inbound and outbound rules that specify what network traffic should be allowed.
type FirewallRuleSet struct {
	Inbound  []FirewallRule `json:"inbound,omitempty"`
	Outbound []FirewallRule `json:"outbound,omitempty"`
}
