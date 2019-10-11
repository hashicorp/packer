//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package network

import (
	"context"
	"encoding/json"
	"net/http"
	"path"

	"github.com/joyent/triton-go/client"
	"github.com/pkg/errors"
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
	fullPath := path.Join("/", c.Client.AccountName, "networks")
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
	}
	respReader, err := c.Client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to list networks")
	}

	var result []*Network
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode list networks response")
	}

	return result, nil
}

type GetInput struct {
	ID string
}

func (c *NetworkClient) Get(ctx context.Context, input *GetInput) (*Network, error) {
	fullPath := path.Join("/", c.Client.AccountName, "networks", input.ID)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
	}
	respReader, err := c.Client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to get network")
	}

	var result *Network
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode get network response")
	}

	return result, nil
}
