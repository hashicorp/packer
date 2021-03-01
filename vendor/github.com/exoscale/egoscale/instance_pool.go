package egoscale

// InstancePoolState represents the state of an Instance Pool.
type InstancePoolState string

const (
	// InstancePoolCreating creating state.
	InstancePoolCreating InstancePoolState = "creating"
	// InstancePoolRunning running state.
	InstancePoolRunning InstancePoolState = "running"
	// InstancePoolDestroying destroying state.
	InstancePoolDestroying InstancePoolState = "destroying"
	// InstancePoolScalingUp scaling up state.
	InstancePoolScalingUp InstancePoolState = "scaling-up"
	// InstancePoolScalingDown scaling down state.
	InstancePoolScalingDown InstancePoolState = "scaling-down"
)

// InstancePool represents an Instance Pool.
type InstancePool struct {
	ID                   *UUID             `json:"id"`
	Name                 string            `json:"name"`
	Description          string            `json:"description"`
	ServiceOfferingID    *UUID             `json:"serviceofferingid"`
	TemplateID           *UUID             `json:"templateid"`
	ZoneID               *UUID             `json:"zoneid"`
	AntiAffinityGroupIDs []UUID            `json:"affinitygroupids"`
	SecurityGroupIDs     []UUID            `json:"securitygroupids"`
	NetworkIDs           []UUID            `json:"networkids"`
	IPv6                 bool              `json:"ipv6"`
	KeyPair              string            `json:"keypair"`
	UserData             string            `json:"userdata"`
	Size                 int               `json:"size"`
	RootDiskSize         int               `json:"rootdisksize"`
	State                InstancePoolState `json:"state"`
	VirtualMachines      []VirtualMachine  `json:"virtualmachines"`
}

// CreateInstancePool represents an Instance Pool creation API request.
type CreateInstancePool struct {
	Name                 string `json:"name"`
	Description          string `json:"description,omitempty"`
	ServiceOfferingID    *UUID  `json:"serviceofferingid"`
	TemplateID           *UUID  `json:"templateid"`
	ZoneID               *UUID  `json:"zoneid"`
	AntiAffinityGroupIDs []UUID `json:"affinitygroupids,omitempty"`
	SecurityGroupIDs     []UUID `json:"securitygroupids,omitempty"`
	NetworkIDs           []UUID `json:"networkids,omitempty"`
	IPv6                 bool   `json:"ipv6,omitempty"`
	KeyPair              string `json:"keypair,omitempty"`
	UserData             string `json:"userdata,omitempty"`
	Size                 int    `json:"size"`
	RootDiskSize         int    `json:"rootdisksize,omitempty"`
	_                    bool   `name:"createInstancePool" description:"Create an Instance Pool"`
}

// CreateInstancePoolResponse represents an Instance Pool creation API response.
type CreateInstancePoolResponse struct {
	ID                   *UUID             `json:"id"`
	Name                 string            `json:"name"`
	Description          string            `json:"description"`
	ServiceOfferingID    *UUID             `json:"serviceofferingid"`
	TemplateID           *UUID             `json:"templateid"`
	ZoneID               *UUID             `json:"zoneid"`
	AntiAffinityGroupIDs []UUID            `json:"affinitygroupids"`
	SecurityGroupIDs     []UUID            `json:"securitygroupids"`
	NetworkIDs           []UUID            `json:"networkids"`
	IPv6                 bool              `json:"ipv6"`
	KeyPair              string            `json:"keypair"`
	UserData             string            `json:"userdata"`
	Size                 int64             `json:"size"`
	RootDiskSize         int               `json:"rootdisksize"`
	State                InstancePoolState `json:"state"`
}

// Response returns an empty structure to unmarshal an Instance Pool creation API response into.
func (CreateInstancePool) Response() interface{} {
	return new(CreateInstancePoolResponse)
}

// UpdateInstancePool represents an Instance Pool update API request.
type UpdateInstancePool struct {
	ID           *UUID  `json:"id"`
	ZoneID       *UUID  `json:"zoneid"`
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	TemplateID   *UUID  `json:"templateid,omitempty"`
	RootDiskSize int    `json:"rootdisksize,omitempty"`
	UserData     string `json:"userdata,omitempty"`
	IPv6         bool   `json:"ipv6,omitempty"`
	_            bool   `name:"updateInstancePool" description:"Update an Instance Pool"`
}

// Response returns an empty structure to unmarshal an Instance Pool update API response into.
func (UpdateInstancePool) Response() interface{} {
	return new(BooleanResponse)
}

// ScaleInstancePool represents an Instance Pool scaling API request.
type ScaleInstancePool struct {
	ID     *UUID `json:"id"`
	ZoneID *UUID `json:"zoneid"`
	Size   int   `json:"size"`
	_      bool  `name:"scaleInstancePool" description:"Scale an Instance Pool"`
}

// Response returns an empty structure to unmarshal an Instance Pool scaling API response into.
func (ScaleInstancePool) Response() interface{} {
	return new(BooleanResponse)
}

// DestroyInstancePool represents an Instance Pool destruction API request.
type DestroyInstancePool struct {
	ID     *UUID `json:"id"`
	ZoneID *UUID `json:"zoneid"`
	_      bool  `name:"destroyInstancePool" description:"Destroy an Instance Pool"`
}

// Response returns an empty structure to unmarshal an Instance Pool destruction API response into.
func (DestroyInstancePool) Response() interface{} {
	return new(BooleanResponse)
}

// GetInstancePool retrieves an Instance Pool's details.
type GetInstancePool struct {
	ID     *UUID `json:"id"`
	ZoneID *UUID `json:"zoneid"`
	_      bool  `name:"getInstancePool" description:"Get an Instance Pool"`
}

// GetInstancePoolResponse get Instance Pool API response.
type GetInstancePoolResponse struct {
	Count         int
	InstancePools []InstancePool `json:"instancepool"`
}

// Response returns an empty structure to unmarshal an Instance Pool get API response into.
func (GetInstancePool) Response() interface{} {
	return new(GetInstancePoolResponse)
}

// ListInstancePools represents a list Instance Pool API request.
type ListInstancePools struct {
	ZoneID *UUID `json:"zoneid"`
	_      bool  `name:"listInstancePools" description:"List Instance Pools"`
}

// ListInstancePoolsResponse represents a list Instance Pool API response.
type ListInstancePoolsResponse struct {
	Count         int
	InstancePools []InstancePool `json:"instancepool"`
}

// Response returns an empty structure to unmarshal an Instance Pool list API response into.
func (ListInstancePools) Response() interface{} {
	return new(ListInstancePoolsResponse)
}

// EvictInstancePoolMembers represents an Instance Pool members eviction API request.
type EvictInstancePoolMembers struct {
	ID        *UUID  `json:"id"`
	ZoneID    *UUID  `json:"zoneid"`
	MemberIDs []UUID `json:"memberids"`
	_         bool   `name:"evictInstancePoolMembers" description:"Evict some Instance Pool members"`
}

// Response returns an empty structure to unmarshal an Instance Pool members eviction API response into.
func (EvictInstancePoolMembers) Response() interface{} {
	return new(BooleanResponse)
}
