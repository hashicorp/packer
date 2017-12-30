package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/errwrap"
	"github.com/joyent/triton-go/client"
)

type Network struct {
	Id                  string            `json:"id"`
	Name                string            `json:"name"`
	Public              bool              `json:"public"`
	Fabric              bool              `json:"fabric"`
	Description         string            `json:"description"`
	Subnet              string            `json:"subnet"`
	ProvisioningStartIP string            `json:"provision_start_ip"`
	ProvisioningEndIP   string            `json:"provision_end_ip"`
	Gateway             string            `json:"gateway"`
	Resolvers           []string          `json:"resolvers"`
	Routes              map[string]string `json:"routes"`
	InternetNAT         bool              `json:"internet_nat"`
}

type ListInput struct{}

func (c *NetworkClient) List(ctx context.Context, _ *ListInput) ([]*Network, error) {
	path := fmt.Sprintf("/%s/networks", c.Client.AccountName)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	respReader, err := c.Client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing ListNetworks request: {{err}}", err)
	}

	var result []*Network
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding ListNetworks response: {{err}}", err)
	}

	return result, nil
}

type GetInput struct {
	ID string
}

func (c *NetworkClient) Get(ctx context.Context, input *GetInput) (*Network, error) {
	path := fmt.Sprintf("/%s/networks/%s", c.Client.AccountName, input.ID)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	respReader, err := c.Client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing GetNetwork request: {{err}}", err)
	}

	var result *Network
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding GetNetwork response: {{err}}", err)
	}

	return result, nil
}
