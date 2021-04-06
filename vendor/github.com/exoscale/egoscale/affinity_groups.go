package egoscale

import (
	"context"
	"fmt"
	"net/url"
)

// AffinityGroup represents an Affinity Group.
//
// Affinity and Anti-Affinity Groups provide a way to influence where VMs should run.
// See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/stable/virtual_machines.html#affinity-groups
type AffinityGroup struct {
	Account           string `json:"account,omitempty" doc:"the account owning the Affinity Group"`
	Description       string `json:"description,omitempty" doc:"the description of the Affinity Group"`
	ID                *UUID  `json:"id,omitempty" doc:"the ID of the Affinity Group"`
	Name              string `json:"name,omitempty" doc:"the name of the Affinity Group"`
	Type              string `json:"type,omitempty" doc:"the type of the Affinity Group"`
	VirtualMachineIDs []UUID `json:"virtualmachineIds,omitempty" doc:"virtual machine IDs associated with this Affinity Group"`
}

// ListRequest builds the ListAffinityGroups request.
func (ag AffinityGroup) ListRequest() (ListCommand, error) {
	return &ListAffinityGroups{
		ID:   ag.ID,
		Name: ag.Name,
	}, nil
}

// Delete deletes the given Affinity Group.
func (ag AffinityGroup) Delete(ctx context.Context, client *Client) error {
	if ag.ID == nil && ag.Name == "" {
		return fmt.Errorf("an Affinity Group may only be deleted using ID or Name")
	}

	req := &DeleteAffinityGroup{}

	if ag.ID != nil {
		req.ID = ag.ID
	} else {
		req.Name = ag.Name
	}

	return client.BooleanRequestWithContext(ctx, req)
}

// AffinityGroupType represent an Affinity Group type.
type AffinityGroupType struct {
	Type string `json:"type,omitempty" doc:"the type of the Affinity Group"`
}

// CreateAffinityGroup (Async) represents a new Affinity Group.
type CreateAffinityGroup struct {
	Description string `json:"description,omitempty" doc:"Optional description of the Affinity Group"`
	Name        string `json:"name" doc:"Name of the Affinity Group"`
	Type        string `json:"type" doc:"Type of the Affinity Group from the available Affinity Group Group types"`
	_           bool   `name:"createAffinityGroup" description:"Creates an Affinity Group Group"`
}

func (req CreateAffinityGroup) onBeforeSend(params url.Values) error {
	// Name must be set, but can be empty.
	if req.Name == "" {
		params.Set("name", "")
	}
	return nil
}

// Response returns the struct to unmarshal.
func (CreateAffinityGroup) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job.
func (CreateAffinityGroup) AsyncResponse() interface{} {
	return new(AffinityGroup)
}

// UpdateVMAffinityGroup (Async) represents a modification of an Affinity Group.
type UpdateVMAffinityGroup struct {
	ID                 *UUID    `json:"id" doc:"The ID of the virtual machine"`
	AffinityGroupIDs   []UUID   `json:"affinitygroupids,omitempty" doc:"comma separated list of Affinity Groups id that are going to be applied to the virtual machine. Should be passed only when vm is created from a zone with Basic Network support. Mutually exclusive with securitygroupnames parameter"`
	AffinityGroupNames []string `json:"affinitygroupnames,omitempty" doc:"comma separated list of Affinity Groups names that are going to be applied to the virtual machine. Should be passed only when vm is created from a zone with Basic Network support. Mutually exclusive with securitygroupids parameter"`
	_                  bool     `name:"updateVMAffinityGroup" description:"Updates the Affinity Group Group associations of a virtual machine. The VM has to be stopped and restarted for the new properties to take effect."`
}

func (req UpdateVMAffinityGroup) onBeforeSend(params url.Values) error {
	// Either AffinityGroupIDs or AffinityGroupNames must be set.
	if len(req.AffinityGroupIDs) == 0 && len(req.AffinityGroupNames) == 0 {
		params.Set("affinitygroupids", "")
	}
	return nil
}

// Response returns the struct to unmarshal.
func (UpdateVMAffinityGroup) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job.
func (UpdateVMAffinityGroup) AsyncResponse() interface{} {
	return new(VirtualMachine)
}

// DeleteAffinityGroup (Async) represents an Affinity Group to be deleted.
type DeleteAffinityGroup struct {
	ID   *UUID  `json:"id,omitempty" doc:"The ID of the Affinity Group. Mutually exclusive with name parameter"`
	Name string `json:"name,omitempty" doc:"The name of the Affinity Group. Mutually exclusive with ID parameter"`
	_    bool   `name:"deleteAffinityGroup" description:"Deletes Affinity Group"`
}

// Response returns the struct to unmarshal.
func (DeleteAffinityGroup) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job.
func (DeleteAffinityGroup) AsyncResponse() interface{} {
	return new(BooleanResponse)
}

//go:generate go run generate/main.go -interface=Listable ListAffinityGroups

// ListAffinityGroups represents an Affinity Groups search.
type ListAffinityGroups struct {
	ID               *UUID  `json:"id,omitempty" doc:"List the Affinity Group by the ID provided"`
	Keyword          string `json:"keyword,omitempty" doc:"List by keyword"`
	Name             string `json:"name,omitempty" doc:"Lists Affinity Groups by name"`
	Page             int    `json:"page,omitempty"`
	PageSize         int    `json:"pagesize,omitempty"`
	Type             string `json:"type,omitempty" doc:"Lists Affinity Groups by type"`
	VirtualMachineID *UUID  `json:"virtualmachineid,omitempty" doc:"Lists Affinity Groups by virtual machine ID"`
	_                bool   `name:"listAffinityGroups" description:"Lists Affinity Groups"`
}

// ListAffinityGroupsResponse represents a list of Affinity Groups.
type ListAffinityGroupsResponse struct {
	Count         int             `json:"count"`
	AffinityGroup []AffinityGroup `json:"affinitygroup"`
}

// ListAffinityGroupTypes represents an Affinity Groups types search.
type ListAffinityGroupTypes struct {
	Keyword  string `json:"keyword,omitempty" doc:"List by keyword"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"pagesize,omitempty"`
	_        bool   `name:"listAffinityGroupTypes" description:"Lists Affinity Group types available"`
}

// Response returns the struct to unmarshal.
func (ListAffinityGroupTypes) Response() interface{} {
	return new(ListAffinityGroupTypesResponse)
}

// ListAffinityGroupTypesResponse represents a list of Affinity Group types.
type ListAffinityGroupTypesResponse struct {
	Count             int                 `json:"count"`
	AffinityGroupType []AffinityGroupType `json:"affinitygrouptype"`
}
