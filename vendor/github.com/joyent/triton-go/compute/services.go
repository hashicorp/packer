package compute

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/hashicorp/errwrap"
	"github.com/joyent/triton-go/client"
)

type ServicesClient struct {
	client *client.Client
}

type Service struct {
	Name     string
	Endpoint string
}

type ListServicesInput struct{}

func (c *ServicesClient) List(ctx context.Context, _ *ListServicesInput) ([]*Service, error) {
	path := fmt.Sprintf("/%s/services", c.client.AccountName)
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

	result := make([]*Service, len(intermediate))
	i = 0
	for _, key := range keys {
		result[i] = &Service{
			Name:     key,
			Endpoint: intermediate[key],
		}
		i++
	}

	return result, nil
}
