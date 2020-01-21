package compute

const (
	iPNetworkExchangeDescription   = "ip network exchange"
	iPNetworkExchangeContainerPath = "/network/v1/ipnetworkexchange/"
	iPNetworkExchangeResourcePath  = "/network/v1/ipnetworkexchange"
)

// IPNetworkExchangesClient details the ip network exchange client
type IPNetworkExchangesClient struct {
	ResourceClient
}

// IPNetworkExchanges returns an IPNetworkExchangesClient that can be used to access the
// necessary CRUD functions for IP Network Exchanges.
func (c *Client) IPNetworkExchanges() *IPNetworkExchangesClient {
	return &IPNetworkExchangesClient{
		ResourceClient: ResourceClient{
			Client:              c,
			ResourceDescription: iPNetworkExchangeDescription,
			ContainerPath:       iPNetworkExchangeContainerPath,
			ResourceRootPath:    iPNetworkExchangeResourcePath,
		},
	}
}

// IPNetworkExchangeInfo contains the exported fields necessary to hold all the information about an
// IP Network Exchange
type IPNetworkExchangeInfo struct {
	// Fully Qualified Domain Name
	FQDN string `json:"name"`
	// The name of the IP Network Exchange
	Name string
	// Description of the IP Network Exchange
	Description string `json:"description"`
	// Slice of tags associated with the IP Network Exchange
	Tags []string `json:"tags"`
	// Uniform Resource Identifier for the IP Network Exchange
	URI string `json:"uri"`
}

// CreateIPNetworkExchangeInput details the attributes needed to create an ip network exchange
type CreateIPNetworkExchangeInput struct {
	// The name of the IP Network Exchange to create. Object names can only contain alphanumeric,
	// underscore, dash, and period characters. Names are case-sensitive.
	// Required
	Name string `json:"name"`

	// Description of the IPNetworkExchange
	// Optional
	Description string `json:"description"`

	// String slice of tags to apply to the IP Network Exchange object
	// Optional
	Tags []string `json:"tags"`
}

// CreateIPNetworkExchange creates a new IP Network Exchange from an IPNetworkExchangesClient and an input struct.
// Returns a populated Info struct for the IP Network Exchange, and any errors
func (c *IPNetworkExchangesClient) CreateIPNetworkExchange(input *CreateIPNetworkExchangeInput) (*IPNetworkExchangeInfo, error) {
	input.Name = c.getQualifiedName(input.Name)

	var ipInfo IPNetworkExchangeInfo
	if err := c.createResource(&input, &ipInfo); err != nil {
		return nil, err
	}

	return c.success(&ipInfo)
}

// GetIPNetworkExchangeInput details the attributes needed to retrieve an ip network exchange
type GetIPNetworkExchangeInput struct {
	// The name of the IP Network Exchange to query for. Case-sensitive
	// Required
	Name string `json:"name"`
}

// GetIPNetworkExchange returns a populated IPNetworkExchangeInfo struct from an input struct
func (c *IPNetworkExchangesClient) GetIPNetworkExchange(input *GetIPNetworkExchangeInput) (*IPNetworkExchangeInfo, error) {
	input.Name = c.getQualifiedName(input.Name)

	var ipInfo IPNetworkExchangeInfo
	if err := c.getResource(input.Name, &ipInfo); err != nil {
		return nil, err
	}

	return c.success(&ipInfo)
}

// DeleteIPNetworkExchangeInput details the attributes neccessary to delete an ip network exchange
type DeleteIPNetworkExchangeInput struct {
	// The name of the IP Network Exchange to query for. Case-sensitive
	// Required
	Name string `json:"name"`
}

// DeleteIPNetworkExchange deletes the specified ip network exchange
func (c *IPNetworkExchangesClient) DeleteIPNetworkExchange(input *DeleteIPNetworkExchangeInput) error {
	return c.deleteResource(input.Name)
}

// Unqualifies any qualified fields in the IPNetworkExchangeInfo struct
func (c *IPNetworkExchangesClient) success(info *IPNetworkExchangeInfo) (*IPNetworkExchangeInfo, error) {
	info.Name = c.getUnqualifiedName(info.FQDN)
	return info, nil
}
