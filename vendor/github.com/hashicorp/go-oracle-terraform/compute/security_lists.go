package compute

// SecurityListsClient is a client for the Security List functions of the Compute API.
type SecurityListsClient struct {
	ResourceClient
}

// SecurityLists obtains a SecurityListsClient which can be used to access to the
// Security List functions of the Compute API
func (c *Client) SecurityLists() *SecurityListsClient {
	return &SecurityListsClient{
		ResourceClient: ResourceClient{
			Client:              c,
			ResourceDescription: "security list",
			ContainerPath:       "/seclist/",
			ResourceRootPath:    "/seclist",
		}}
}

// SecurityListPolicy defines the constants a security list policy can be
type SecurityListPolicy string

const (
	// SecurityListPolicyDeny - deny
	SecurityListPolicyDeny SecurityListPolicy = "deny"
	// SecurityListPolicyReject - reject
	SecurityListPolicyReject SecurityListPolicy = "reject"
	// SecurityListPolicyPermit - permit
	SecurityListPolicyPermit SecurityListPolicy = "permit"
)

// SecurityListInfo describes an existing security list.
type SecurityListInfo struct {
	// Shows the default account for your identity domain.
	Account string `json:"account"`
	// A description of the security list.
	Description string `json:"description"`
	// Fully Qualified Domain Name
	FQDN string `json:"name"`
	// The name of the security list
	Name string
	// The policy for outbound traffic from the security list.
	OutboundCIDRPolicy SecurityListPolicy `json:"outbound_cidr_policy"`
	// The policy for inbound traffic to the security list
	Policy SecurityListPolicy `json:"policy"`
	// Uniform Resource Identifier
	URI string `json:"uri"`
}

// CreateSecurityListInput defines a security list to be created.
type CreateSecurityListInput struct {
	// A description of the security list.
	// Optional
	Description string `json:"description"`
	// The three-part name of the Security List (/Compute-identity_domain/user/object).
	// Object names can contain only alphanumeric characters, hyphens, underscores, and periods. Object names are case-sensitive.
	// Required
	Name string `json:"name"`
	// The policy for outbound traffic from the security list.
	// Optional (defaults to `permit`)
	OutboundCIDRPolicy SecurityListPolicy `json:"outbound_cidr_policy"`
	// The policy for inbound traffic to the security list.
	// Optional (defaults to `deny`)
	Policy SecurityListPolicy `json:"policy"`
}

// CreateSecurityList creates a new security list with the given name, policy and outbound CIDR policy.
func (c *SecurityListsClient) CreateSecurityList(createInput *CreateSecurityListInput) (*SecurityListInfo, error) {
	createInput.Name = c.getQualifiedName(createInput.Name)
	var listInfo SecurityListInfo
	if err := c.createResource(createInput, &listInfo); err != nil {
		return nil, err
	}

	return c.success(&listInfo)
}

// GetSecurityListInput describes the security list you want to get
type GetSecurityListInput struct {
	// The three-part name of the Security List (/Compute-identity_domain/user/object).
	// Required
	Name string `json:"name"`
}

// GetSecurityList retrieves the security list with the given name.
func (c *SecurityListsClient) GetSecurityList(getInput *GetSecurityListInput) (*SecurityListInfo, error) {
	var listInfo SecurityListInfo
	if err := c.getResource(getInput.Name, &listInfo); err != nil {
		return nil, err
	}

	return c.success(&listInfo)
}

// UpdateSecurityListInput defines what to update in a security list
type UpdateSecurityListInput struct {
	// A description of the security list.
	// Optional
	Description string `json:"description"`
	// The three-part name of the Security List (/Compute-identity_domain/user/object).
	// Required
	Name string `json:"name"`
	// The policy for outbound traffic from the security list.
	// Optional (defaults to `permit`)
	OutboundCIDRPolicy SecurityListPolicy `json:"outbound_cidr_policy"`
	// The policy for inbound traffic to the security list.
	// Optional (defaults to `deny`)
	Policy SecurityListPolicy `json:"policy"`
}

// UpdateSecurityList updates the policy and outbound CIDR pol
func (c *SecurityListsClient) UpdateSecurityList(updateInput *UpdateSecurityListInput) (*SecurityListInfo, error) {
	updateInput.Name = c.getQualifiedName(updateInput.Name)
	var listInfo SecurityListInfo
	if err := c.updateResource(updateInput.Name, updateInput, &listInfo); err != nil {
		return nil, err
	}

	return c.success(&listInfo)
}

// DeleteSecurityListInput describes the security list to destroy
type DeleteSecurityListInput struct {
	// The three-part name of the Security List (/Compute-identity_domain/user/object).
	// Required
	Name string `json:"name"`
}

// DeleteSecurityList deletes the security list with the given name.
func (c *SecurityListsClient) DeleteSecurityList(deleteInput *DeleteSecurityListInput) error {
	return c.deleteResource(deleteInput.Name)
}

func (c *SecurityListsClient) success(listInfo *SecurityListInfo) (*SecurityListInfo, error) {
	listInfo.Name = c.getUnqualifiedName(listInfo.FQDN)
	return listInfo, nil
}
