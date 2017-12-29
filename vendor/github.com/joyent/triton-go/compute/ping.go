package compute

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/errwrap"
	"github.com/joyent/triton-go/client"
)

const pingEndpoint = "/--ping"

type CloudAPI struct {
	Versions []string `json:"versions"`
}

type PingOutput struct {
	Ping     string   `json:"ping"`
	CloudAPI CloudAPI `json:"cloudapi"`
}

// Ping sends a request to the '/--ping' endpoint and returns a `pong` as well
// as a list of API version numbers your instance of CloudAPI is presenting.
func (c *ComputeClient) Ping(ctx context.Context) (*PingOutput, error) {
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   pingEndpoint,
	}
	response, err := c.Client.ExecuteRequestRaw(ctx, reqInputs)
	if response == nil {
		return nil, fmt.Errorf("Ping request has empty response")
	}
	if response.Body != nil {
		defer response.Body.Close()
	}
	if response.StatusCode == http.StatusNotFound || response.StatusCode == http.StatusGone {
		return nil, &client.TritonError{
			StatusCode: response.StatusCode,
			Code:       "ResourceNotFound",
		}
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing Get request: {{err}}",
			c.Client.DecodeError(response.StatusCode, response.Body))
	}

	var result *PingOutput
	decoder := json.NewDecoder(response.Body)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding Get response: {{err}}", err)
	}

	return result, nil
}
