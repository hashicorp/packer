package egoscale

import (
	"context"
	"fmt"
	"net/url"
)

// AntiAffinityGroup represents an Anti-Affinity Group.
type AntiAffinityGroup struct {
	Account           string `json:"account,omitempty" doc:"the account owning the Anti-Affinity Group"`
	Description       string `json:"description,omitempty" doc:"the description of the Anti-Affinity Group"`
	ID                *UUID  `json:"id,omitempty" doc:"the ID of the Anti-Affinity Group"`
	Name              string `json:"name,omitempty" doc:"the name of the Anti-Affinity Group"`
	Type              string `json:"type,omitempty" doc:"the type of the Anti-Affinity Group"`
	VirtualMachineIDs []UUID `json:"virtualmachineIds,omitempty" doc:"virtual machine IDs associated with this Anti-Affinity Group"`
}

// ListRequest builds the ListAntiAffinityGroups request.
func (ag AntiAffinityGroup) ListRequest() (ListCommand, error) {
	return &ListAffinityGroups{
		ID:   ag.ID,
		Name: ag.Name,
	}, nil
}

// Delete deletes the given Anti-Affinity Group.
func (ag AntiAffinityGroup) Delete(ctx context.Context, client *Client) error {
	if ag.ID == nil && ag.Name == "" {
		return fmt.Errorf("an Anti-Affinity Group may only be deleted using ID or Name")
	}

	req := &DeleteAffinityGroup{}

	if ag.ID != nil {
		req.ID = ag.ID
	} else {
		req.Name = ag.Name
	}

	return client.BooleanRequestWithContext(ctx, req)
}

// CreateAntiAffinityGroup represents an Anti-Affinity Group creation.
type CreateAntiAffinityGroup struct {
	Name        string `json:"name" doc:"Name of the Anti-Affinity Group"`
	Description string `json:"description,omitempty" doc:"Optional description of the Anti-Affinity Group"`
	_           bool   `name:"createAntiAffinityGroup" description:"Creates an Anti-Affinity Group"`
}

func (req CreateAntiAffinityGroup) onBeforeSend(params url.Values) error {
	// Name must be set, but can be empty.
	if req.Name == "" {
		params.Set("name", "")
	}
	return nil
}

// Response returns the struct to unmarshal.
func (CreateAntiAffinityGroup) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job.
func (CreateAntiAffinityGroup) AsyncResponse() interface{} {
	return new(AffinityGroup)
}

//go:generate go run generate/main.go -interface=Listable ListAntiAffinityGroups

// ListAntiAffinityGroups represents an Anti-Affinity Groups search.
type ListAntiAffinityGroups struct {
	ID               *UUID  `json:"id,omitempty" doc:"List the Anti-Affinity Group by the ID provided"`
	Keyword          string `json:"keyword,omitempty" doc:"List by keyword"`
	Name             string `json:"name,omitempty" doc:"Lists Anti-Affinity Groups by name"`
	Page             int    `json:"page,omitempty"`
	PageSize         int    `json:"pagesize,omitempty"`
	VirtualMachineID *UUID  `json:"virtualmachineid,omitempty" doc:"Lists Anti-Affinity Groups by virtual machine ID"`
	_                bool   `name:"listAntiAffinityGroups" description:"Lists Anti-Affinity Groups"`
}

// ListAntiAffinityGroupsResponse represents a list of Anti-Affinity Groups.
type ListAntiAffinityGroupsResponse struct {
	Count             int             `json:"count"`
	AntiAffinityGroup []AffinityGroup `json:"antiaffinitygroup"`
}

// DeleteAntiAffinityGroup (Async) represents an Anti-Affinity Group to be deleted.
type DeleteAntiAffinityGroup struct {
	ID   *UUID  `json:"id,omitempty" doc:"The ID of the Anti-Affinity Group. Mutually exclusive with name parameter"`
	Name string `json:"name,omitempty" doc:"The name of the Anti-Affinity Group. Mutually exclusive with ID parameter"`
	_    bool   `name:"deleteAntiAffinityGroup" description:"Deletes Anti-Affinity Group"`
}

// Response returns the struct to unmarshal.
func (DeleteAntiAffinityGroup) Response() interface{} {
	return new(AsyncJobResult)
}

// AsyncResponse returns the struct to unmarshal the async job.
func (DeleteAntiAffinityGroup) AsyncResponse() interface{} {
	return new(BooleanResponse)
}
