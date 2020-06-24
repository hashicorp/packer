package metadata

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"strings"

	"github.com/ucloud/ucloud-sdk-go/private/protocol/http"
)

const globalEndpoint = "http://100.80.80.80"

type Client interface {
	GetMetadata(path string) (string, error)
	GetUserData() (string, error)
	GetVendorData() (string, error)
	GetInstanceIdentityDocument() (meta Metadata, err error)
}

type DefaultClient struct {
	httpClient http.Client
}

func NewClient() Client {
	client := http.NewHttpClient()
	return DefaultClient{httpClient: &client}
}

func (client DefaultClient) GetInstanceIdentityDocument() (meta Metadata, err error) {
	resp, err := client.GetMetadata(".json")
	if err != nil {
		return meta, errors.Errorf("failed to get instance identity document, %s", err)
	}

	if err := json.NewDecoder(strings.NewReader(resp)).Decode(&meta); err != nil {
		return meta, errors.Errorf("failed to decode instance identity document, %s", err)
	}

	return meta, nil
}

func (client DefaultClient) GetMetadata(path string) (string, error) {
	return client.SendRequest(fmt.Sprintf("/meta-data/v1%s", path))
}

func (client DefaultClient) GetUserData() (string, error) {
	return client.SendRequest(fmt.Sprintf("/user-data"))
}

func (client DefaultClient) GetVendorData() (string, error) {
	return client.SendRequest(fmt.Sprintf("/vendor-data"))
}

func (client DefaultClient) SendRequest(path string) (string, error) {
	req := http.NewHttpRequest()
	_ = req.SetMethod("GET")
	_ = req.SetURL(fmt.Sprintf("%s%s", globalEndpoint, path))

	resp, err := client.httpClient.Send(req)
	if err != nil {
		return "", err
	}

	return string(resp.GetBody()), nil
}

// SetHttpClient will setup a http client
func (client *DefaultClient) SetHttpClient(httpClient http.Client) error {
	client.httpClient = httpClient
	return nil
}
