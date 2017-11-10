package compute

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/errwrap"
	"github.com/joyent/triton-go/client"
)

type PackagesClient struct {
	client *client.Client
}

type Package struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Memory      int64  `json:"memory"`
	Disk        int64  `json:"disk"`
	Swap        int64  `json:"swap"`
	LWPs        int64  `json:"lwps"`
	VCPUs       int64  `json:"vcpus"`
	Version     string `json:"version"`
	Group       string `json:"group"`
	Description string `json:"description"`
	Default     bool   `json:"default"`
}

type ListPackagesInput struct {
	Name    string `json:"name"`
	Memory  int64  `json:"memory"`
	Disk    int64  `json:"disk"`
	Swap    int64  `json:"swap"`
	LWPs    int64  `json:"lwps"`
	VCPUs   int64  `json:"vcpus"`
	Version string `json:"version"`
	Group   string `json:"group"`
}

func (c *PackagesClient) List(ctx context.Context, input *ListPackagesInput) ([]*Package, error) {
	path := fmt.Sprintf("/%s/packages", c.client.AccountName)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
		Body:   input,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing List request: {{err}}", err)
	}

	var result []*Package
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding List response: {{err}}", err)
	}

	return result, nil
}

type GetPackageInput struct {
	ID string
}

func (c *PackagesClient) Get(ctx context.Context, input *GetPackageInput) (*Package, error) {
	path := fmt.Sprintf("/%s/packages/%s", c.client.AccountName, input.ID)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing Get request: {{err}}", err)
	}

	var result *Package
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding Get response: {{err}}", err)
	}

	return result, nil
}
