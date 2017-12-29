package compute

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"time"

	"github.com/hashicorp/errwrap"
	"github.com/joyent/triton-go/client"
)

type SnapshotsClient struct {
	client *client.Client
}

type Snapshot struct {
	Name    string
	State   string
	Created time.Time
	Updated time.Time
}

type ListSnapshotsInput struct {
	MachineID string
}

func (c *SnapshotsClient) List(ctx context.Context, input *ListSnapshotsInput) ([]*Snapshot, error) {
	path := fmt.Sprintf("/%s/machines/%s/snapshots", c.client.AccountName, input.MachineID)
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

	var result []*Snapshot
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding List response: {{err}}", err)
	}

	return result, nil
}

type GetSnapshotInput struct {
	MachineID string
	Name      string
}

func (c *SnapshotsClient) Get(ctx context.Context, input *GetSnapshotInput) (*Snapshot, error) {
	path := fmt.Sprintf("/%s/machines/%s/snapshots/%s", c.client.AccountName, input.MachineID, input.Name)
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

	var result *Snapshot
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding Get response: {{err}}", err)
	}

	return result, nil
}

type DeleteSnapshotInput struct {
	MachineID string
	Name      string
}

func (c *SnapshotsClient) Delete(ctx context.Context, input *DeleteSnapshotInput) error {
	path := fmt.Sprintf("/%s/machines/%s/snapshots/%s", c.client.AccountName, input.MachineID, input.Name)
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

type StartMachineFromSnapshotInput struct {
	MachineID string
	Name      string
}

func (c *SnapshotsClient) StartMachine(ctx context.Context, input *StartMachineFromSnapshotInput) error {
	path := fmt.Sprintf("/%s/machines/%s/snapshots/%s", c.client.AccountName, input.MachineID, input.Name)
	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errwrap.Wrapf("Error executing StartMachine request: {{err}}", err)
	}

	return nil
}

type CreateSnapshotInput struct {
	MachineID string
	Name      string
}

func (c *SnapshotsClient) Create(ctx context.Context, input *CreateSnapshotInput) (*Snapshot, error) {
	path := fmt.Sprintf("/%s/machines/%s/snapshots", c.client.AccountName, input.MachineID)

	data := make(map[string]interface{})
	data["name"] = input.Name

	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   path,
		Body:   data,
	}

	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing Create request: {{err}}", err)
	}

	var result *Snapshot
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding Create response: {{err}}", err)
	}

	return result, nil
}
