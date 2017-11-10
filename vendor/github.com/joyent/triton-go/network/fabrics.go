package network

import (
	"encoding/json"
	"fmt"
	"net/http"

	"context"

	"github.com/hashicorp/errwrap"
	"github.com/joyent/triton-go/client"
)

type FabricsClient struct {
	client *client.Client
}

type FabricVLAN struct {
	Name        string `json:"name"`
	ID          int    `json:"vlan_id"`
	Description string `json:"description"`
}

type ListVLANsInput struct{}

func (c *FabricsClient) ListVLANs(ctx context.Context, _ *ListVLANsInput) ([]*FabricVLAN, error) {
	path := fmt.Sprintf("/%s/fabrics/default/vlans", c.client.AccountName)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing ListVLANs request: {{err}}", err)
	}

	var result []*FabricVLAN
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding ListVLANs response: {{err}}", err)
	}

	return result, nil
}

type CreateVLANInput struct {
	Name        string `json:"name"`
	ID          int    `json:"vlan_id"`
	Description string `json:"description,omitempty"`
}

func (c *FabricsClient) CreateVLAN(ctx context.Context, input *CreateVLANInput) (*FabricVLAN, error) {
	path := fmt.Sprintf("/%s/fabrics/default/vlans", c.client.AccountName)
	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   path,
		Body:   input,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing CreateVLAN request: {{err}}", err)
	}

	var result *FabricVLAN
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding CreateVLAN response: {{err}}", err)
	}

	return result, nil
}

type UpdateVLANInput struct {
	ID          int    `json:"-"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *FabricsClient) UpdateVLAN(ctx context.Context, input *UpdateVLANInput) (*FabricVLAN, error) {
	path := fmt.Sprintf("/%s/fabrics/default/vlans/%d", c.client.AccountName, input.ID)
	reqInputs := client.RequestInput{
		Method: http.MethodPut,
		Path:   path,
		Body:   input,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing UpdateVLAN request: {{err}}", err)
	}

	var result *FabricVLAN
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding UpdateVLAN response: {{err}}", err)
	}

	return result, nil
}

type GetVLANInput struct {
	ID int `json:"-"`
}

func (c *FabricsClient) GetVLAN(ctx context.Context, input *GetVLANInput) (*FabricVLAN, error) {
	path := fmt.Sprintf("/%s/fabrics/default/vlans/%d", c.client.AccountName, input.ID)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing GetVLAN request: {{err}}", err)
	}

	var result *FabricVLAN
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding GetVLAN response: {{err}}", err)
	}

	return result, nil
}

type DeleteVLANInput struct {
	ID int `json:"-"`
}

func (c *FabricsClient) DeleteVLAN(ctx context.Context, input *DeleteVLANInput) error {
	path := fmt.Sprintf("/%s/fabrics/default/vlans/%d", c.client.AccountName, input.ID)
	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errwrap.Wrapf("Error executing DeleteVLAN request: {{err}}", err)
	}

	return nil
}

type ListFabricsInput struct {
	FabricVLANID int `json:"-"`
}

func (c *FabricsClient) List(ctx context.Context, input *ListFabricsInput) ([]*Network, error) {
	path := fmt.Sprintf("/%s/fabrics/default/vlans/%d/networks", c.client.AccountName, input.FabricVLANID)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing ListFabrics request: {{err}}", err)
	}

	var result []*Network
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding ListFabrics response: {{err}}", err)
	}

	return result, nil
}

type CreateFabricInput struct {
	FabricVLANID     int               `json:"-"`
	Name             string            `json:"name"`
	Description      string            `json:"description,omitempty"`
	Subnet           string            `json:"subnet"`
	ProvisionStartIP string            `json:"provision_start_ip"`
	ProvisionEndIP   string            `json:"provision_end_ip"`
	Gateway          string            `json:"gateway,omitempty"`
	Resolvers        []string          `json:"resolvers,omitempty"`
	Routes           map[string]string `json:"routes,omitempty"`
	InternetNAT      bool              `json:"internet_nat"`
}

func (c *FabricsClient) Create(ctx context.Context, input *CreateFabricInput) (*Network, error) {
	path := fmt.Sprintf("/%s/fabrics/default/vlans/%d/networks", c.client.AccountName, input.FabricVLANID)
	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   path,
		Body:   input,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing CreateFabric request: {{err}}", err)
	}

	var result *Network
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding CreateFabric response: {{err}}", err)
	}

	return result, nil
}

type GetFabricInput struct {
	FabricVLANID int    `json:"-"`
	NetworkID    string `json:"-"`
}

func (c *FabricsClient) Get(ctx context.Context, input *GetFabricInput) (*Network, error) {
	path := fmt.Sprintf("/%s/fabrics/default/vlans/%d/networks/%s", c.client.AccountName, input.FabricVLANID, input.NetworkID)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing GetFabric request: {{err}}", err)
	}

	var result *Network
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding GetFabric response: {{err}}", err)
	}

	return result, nil
}

type DeleteFabricInput struct {
	FabricVLANID int    `json:"-"`
	NetworkID    string `json:"-"`
}

func (c *FabricsClient) Delete(ctx context.Context, input *DeleteFabricInput) error {
	path := fmt.Sprintf("/%s/fabrics/default/vlans/%d/networks/%s", c.client.AccountName, input.FabricVLANID, input.NetworkID)
	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errwrap.Wrapf("Error executing DeleteFabric request: {{err}}", err)
	}

	return nil
}
