package compute

// SecurityApplicationsClient is a client for the Security Application functions of the Compute API.
type SecurityApplicationsClient struct {
	ResourceClient
}

// SecurityApplications obtains a SecurityApplicationsClient which can be used to access to the
// Security Application functions of the Compute API
func (c *Client) SecurityApplications() *SecurityApplicationsClient {
	return &SecurityApplicationsClient{
		ResourceClient: ResourceClient{
			Client:              c,
			ResourceDescription: "security application",
			ContainerPath:       "/secapplication/",
			ResourceRootPath:    "/secapplication",
		}}
}

// SecurityApplicationInfo describes an existing security application.
type SecurityApplicationInfo struct {
	// A description of the security application.
	Description string `json:"description"`
	// The TCP or UDP destination port number. This can be a port range, such as 5900-5999 for TCP.
	DPort string `json:"dport"`
	// Fully Qualified Domain Name
	FQDN string `json:"name"`
	// The ICMP code.
	ICMPCode SecurityApplicationICMPCode `json:"icmpcode"`
	// The ICMP type.
	ICMPType SecurityApplicationICMPType `json:"icmptype"`
	// The three-part name of the Security Application (/Compute-identity_domain/user/object).
	Name string
	// The protocol to use.
	Protocol SecurityApplicationProtocol `json:"protocol"`
	// The Uniform Resource Identifier
	URI string `json:"uri"`
}

// SecurityApplicationProtocol defines the constants for a security application protocol
type SecurityApplicationProtocol string

const (
	// All - all
	All SecurityApplicationProtocol = "all"
	// AH - ah
	AH SecurityApplicationProtocol = "ah"
	// ESP - esp
	ESP SecurityApplicationProtocol = "esp"
	// ICMP - icmp
	ICMP SecurityApplicationProtocol = "icmp"
	// ICMPV6 - icmpv6
	ICMPV6 SecurityApplicationProtocol = "icmpv6"
	// IGMP - igmp
	IGMP SecurityApplicationProtocol = "igmp"
	// IPIP - ipip
	IPIP SecurityApplicationProtocol = "ipip"
	// GRE - gre
	GRE SecurityApplicationProtocol = "gre"
	// MPLSIP - mplsip
	MPLSIP SecurityApplicationProtocol = "mplsip"
	// OSPF - ospf
	OSPF SecurityApplicationProtocol = "ospf"
	// PIM - pim
	PIM SecurityApplicationProtocol = "pim"
	// RDP - rdp
	RDP SecurityApplicationProtocol = "rdp"
	// SCTP - sctp
	SCTP SecurityApplicationProtocol = "sctp"
	// TCP - tcp
	TCP SecurityApplicationProtocol = "tcp"
	// UDP - udp
	UDP SecurityApplicationProtocol = "udp"
)

// SecurityApplicationICMPCode  defines the constants an icmp code can be
type SecurityApplicationICMPCode string

const (
	// Admin - admin
	Admin SecurityApplicationICMPCode = "admin"
	// Df - df
	Df SecurityApplicationICMPCode = "df"
	// Host - host
	Host SecurityApplicationICMPCode = "host"
	// Network - network
	Network SecurityApplicationICMPCode = "network"
	// Port - port
	Port SecurityApplicationICMPCode = "port"
	// Protocol - protocol
	Protocol SecurityApplicationICMPCode = "protocol"
)

// SecurityApplicationICMPType defines the constants an icmp type can be
type SecurityApplicationICMPType string

const (
	// Echo - echo
	Echo SecurityApplicationICMPType = "echo"
	// Reply - reply
	Reply SecurityApplicationICMPType = "reply"
	// TTL - ttl
	TTL SecurityApplicationICMPType = "ttl"
	// TraceRoute - traceroute
	TraceRoute SecurityApplicationICMPType = "traceroute"
	// Unreachable - unreachable
	Unreachable SecurityApplicationICMPType = "unreachable"
)

func (c *SecurityApplicationsClient) success(result *SecurityApplicationInfo) (*SecurityApplicationInfo, error) {
	result.Name = c.getUnqualifiedName(result.FQDN)
	return result, nil
}

// CreateSecurityApplicationInput describes the Security Application to create
type CreateSecurityApplicationInput struct {
	// A description of the security application.
	// Optional
	Description string `json:"description"`
	// The TCP or UDP destination port number.
	// You can also specify a port range, such as 5900-5999 for TCP.
	// This parameter isn't relevant to the icmp protocol.
	// Required if the Protocol is TCP or UDP
	DPort string `json:"dport"`
	// The ICMP code. This parameter is relevant only if you specify ICMP as the protocol.
	// If you specify icmp as the protocol and don't specify icmptype or icmpcode, then all ICMP packets are matched.
	// Optional
	ICMPCode SecurityApplicationICMPCode `json:"icmpcode,omitempty"`
	// This parameter is relevant only if you specify ICMP as the protocol.
	// If you specify icmp as the protocol and don't specify icmptype or icmpcode, then all ICMP packets are matched.
	// Optional
	ICMPType SecurityApplicationICMPType `json:"icmptype,omitempty"`
	// The three-part name of the Security Application (/Compute-identity_domain/user/object).
	// Object names can contain only alphanumeric characters, hyphens, underscores, and periods. Object names are case-sensitive.
	// Required
	Name string `json:"name"`
	// The protocol to use.
	// Required
	Protocol SecurityApplicationProtocol `json:"protocol"`
}

// CreateSecurityApplication creates a new security application.
func (c *SecurityApplicationsClient) CreateSecurityApplication(input *CreateSecurityApplicationInput) (*SecurityApplicationInfo, error) {
	input.Name = c.getQualifiedName(input.Name)

	var appInfo SecurityApplicationInfo
	if err := c.createResource(&input, &appInfo); err != nil {
		return nil, err
	}

	return c.success(&appInfo)
}

// GetSecurityApplicationInput describes the Security Application to obtain
type GetSecurityApplicationInput struct {
	// The three-part name of the Security Application (/Compute-identity_domain/user/object).
	// Required
	Name string `json:"name"`
}

// GetSecurityApplication retrieves the security application with the given name.
func (c *SecurityApplicationsClient) GetSecurityApplication(input *GetSecurityApplicationInput) (*SecurityApplicationInfo, error) {
	var appInfo SecurityApplicationInfo
	if err := c.getResource(input.Name, &appInfo); err != nil {
		return nil, err
	}

	return c.success(&appInfo)
}

// DeleteSecurityApplicationInput  describes the Security Application to delete
type DeleteSecurityApplicationInput struct {
	// The three-part name of the Security Application (/Compute-identity_domain/user/object).
	// Required
	Name string `json:"name"`
}

// DeleteSecurityApplication deletes the security application with the given name.
func (c *SecurityApplicationsClient) DeleteSecurityApplication(input *DeleteSecurityApplicationInput) error {
	return c.deleteResource(input.Name)
}
