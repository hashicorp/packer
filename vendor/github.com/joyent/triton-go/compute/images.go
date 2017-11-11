package compute

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/errwrap"
	"github.com/joyent/triton-go/client"
)

type ImagesClient struct {
	client *client.Client
}

type ImageFile struct {
	Compression string `json:"compression"`
	SHA1        string `json:"sha1"`
	Size        int64  `json:"size"`
}

type Image struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	OS           string                 `json:"os"`
	Description  string                 `json:"description"`
	Version      string                 `json:"version"`
	Type         string                 `json:"type"`
	Requirements map[string]interface{} `json:"requirements"`
	Homepage     string                 `json:"homepage"`
	Files        []*ImageFile           `json:"files"`
	PublishedAt  time.Time              `json:"published_at"`
	Owner        string                 `json:"owner"`
	Public       bool                   `json:"public"`
	State        string                 `json:"state"`
	Tags         map[string]string      `json:"tags"`
	EULA         string                 `json:"eula"`
	ACL          []string               `json:"acl"`
	Error        client.TritonError     `json:"error"`
}

type ListImagesInput struct {
	Name    string
	OS      string
	Version string
	Public  bool
	State   string
	Owner   string
	Type    string
}

func (c *ImagesClient) List(ctx context.Context, input *ListImagesInput) ([]*Image, error) {
	path := fmt.Sprintf("/%s/images", c.client.AccountName)

	query := &url.Values{}
	if input.Name != "" {
		query.Set("name", input.Name)
	}
	if input.OS != "" {
		query.Set("os", input.OS)
	}
	if input.Version != "" {
		query.Set("version", input.Version)
	}
	if input.Public {
		query.Set("public", "true")
	}
	if input.State != "" {
		query.Set("state", input.State)
	}
	if input.Owner != "" {
		query.Set("owner", input.Owner)
	}
	if input.Type != "" {
		query.Set("type", input.Type)
	}

	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
		Query:  query,
	}
	respReader, err := c.client.ExecuteRequestURIParams(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing List request: {{err}}", err)
	}

	var result []*Image
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding List response: {{err}}", err)
	}

	return result, nil
}

type GetImageInput struct {
	ImageID string
}

func (c *ImagesClient) Get(ctx context.Context, input *GetImageInput) (*Image, error) {
	path := fmt.Sprintf("/%s/images/%s", c.client.AccountName, input.ImageID)
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

	var result *Image
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding Get response: {{err}}", err)
	}

	return result, nil
}

type DeleteImageInput struct {
	ImageID string
}

func (c *ImagesClient) Delete(ctx context.Context, input *DeleteImageInput) error {
	path := fmt.Sprintf("/%s/images/%s", c.client.AccountName, input.ImageID)
	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errwrap.Wrapf("Error executing Delete request: {{err}}", err)
	}

	return nil
}

type ExportImageInput struct {
	ImageID   string
	MantaPath string
}

type MantaLocation struct {
	MantaURL     string `json:"manta_url"`
	ImagePath    string `json:"image_path"`
	ManifestPath string `json:"manifest_path"`
}

func (c *ImagesClient) Export(ctx context.Context, input *ExportImageInput) (*MantaLocation, error) {
	path := fmt.Sprintf("/%s/images/%s", c.client.AccountName, input.ImageID)
	query := &url.Values{}
	query.Set("action", "export")
	query.Set("manta_path", input.MantaPath)

	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
		Query:  query,
	}
	respReader, err := c.client.ExecuteRequestURIParams(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing Get request: {{err}}", err)
	}

	var result *MantaLocation
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding Get response: {{err}}", err)
	}

	return result, nil
}

type CreateImageFromMachineInput struct {
	MachineID   string            `json:"machine"`
	Name        string            `json:"name"`
	Version     string            `json:"version,omitempty"`
	Description string            `json:"description,omitempty"`
	HomePage    string            `json:"homepage,omitempty"`
	EULA        string            `json:"eula,omitempty"`
	ACL         []string          `json:"acl,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

func (c *ImagesClient) CreateFromMachine(ctx context.Context, input *CreateImageFromMachineInput) (*Image, error) {
	path := fmt.Sprintf("/%s/images", c.client.AccountName)
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
		return nil, errwrap.Wrapf("Error executing CreateFromMachine request: {{err}}", err)
	}

	var result *Image
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding CreateFromMachine response: {{err}}", err)
	}

	return result, nil
}

type UpdateImageInput struct {
	ImageID     string            `json:"-"`
	Name        string            `json:"name"`
	Version     string            `json:"version,omitempty"`
	Description string            `json:"description,omitempty"`
	HomePage    string            `json:"homepage,omitempty"`
	EULA        string            `json:"eula,omitempty"`
	ACL         []string          `json:"acl,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

func (c *ImagesClient) Update(ctx context.Context, input *UpdateImageInput) (*Image, error) {
	path := fmt.Sprintf("/%s/images/%s", c.client.AccountName, input.ImageID)
	query := &url.Values{}
	query.Set("action", "update")

	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   path,
		Query:  query,
		Body:   input,
	}
	respReader, err := c.client.ExecuteRequestURIParams(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing Update request: {{err}}", err)
	}

	var result *Image
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding Update response: {{err}}", err)
	}

	return result, nil
}
