package compute

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"context"

	"github.com/hashicorp/errwrap"
	"github.com/joyent/triton-go/client"
)

type DataCentersClient struct {
	client *client.Client
}

type DataCenter struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type ListDataCentersInput struct{}

func (c *DataCentersClient) List(ctx context.Context, _ *ListDataCentersInput) ([]*DataCenter, error) {
	path := fmt.Sprintf("/%s/datacenters", c.client.AccountName)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing List request: {{err}}", err)
	}

	var intermediate map[string]string
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&intermediate); err != nil {
		return nil, errwrap.Wrapf("Error decoding List response: {{err}}", err)
	}

	keys := make([]string, len(intermediate))
	i := 0
	for k := range intermediate {
		keys[i] = k
		i++
	}
	sort.Strings(keys)

	result := make([]*DataCenter, len(intermediate))
	i = 0
	for _, key := range keys {
		result[i] = &DataCenter{
			Name: key,
			URL:  intermediate[key],
		}
		i++
	}

	return result, nil
}

type GetDataCenterInput struct {
	Name string
}

func (c *DataCentersClient) Get(ctx context.Context, input *GetDataCenterInput) (*DataCenter, error) {
	path := fmt.Sprintf("/%s/datacenters/%s", c.client.AccountName, input.Name)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	resp, err := c.client.ExecuteRequestRaw(ctx, reqInputs)
	if err != nil {
		return nil, errwrap.Wrapf("Error executing Get request: {{err}}", err)
	}

	if resp.StatusCode != http.StatusFound {
		return nil, fmt.Errorf("Error executing Get request: expected status code 302, got %d",
			resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return nil, errors.New("Error decoding Get response: no Location header")
	}

	return &DataCenter{
		Name: input.Name,
		URL:  location,
	}, nil
}
