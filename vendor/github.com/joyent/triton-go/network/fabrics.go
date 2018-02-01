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
	"strconv"

	"github.com/joyent/triton-go/client"
	"github.com/pkg/errors"
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
	fullPath := path.Join("/", c.client.AccountName, "fabrics", "default", "vlans")
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to list VLANs")
	}

	var result []*FabricVLAN
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode list VLANs response")
	}

	return result, nil
}

type CreateVLANInput struct {
	Name        string `json:"name"`
	ID          int    `json:"vlan_id"`
	Description string `json:"description,omitempty"`
}

func (c *FabricsClient) CreateVLAN(ctx context.Context, input *CreateVLANInput) (*FabricVLAN, error) {
	fullPath := path.Join("/", c.client.AccountName, "fabrics", "default", "vlans")
	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   fullPath,
		Body:   input,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to create VLAN")
	}

	var result *FabricVLAN
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode create VLAN response")
	}

	return result, nil
}

type UpdateVLANInput struct {
	ID          int    `json:"-"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *FabricsClient) UpdateVLAN(ctx context.Context, input *UpdateVLANInput) (*FabricVLAN, error) {
	fullPath := path.Join("/", c.client.AccountName, "fabrics", "default", "vlans", strconv.Itoa(input.ID))
	reqInputs := client.RequestInput{
		Method: http.MethodPut,
		Path:   fullPath,
		Body:   input,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to update VLAN")
	}

	var result *FabricVLAN
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode update VLAN response")
	}

	return result, nil
}

type GetVLANInput struct {
	ID int `json:"-"`
}

func (c *FabricsClient) GetVLAN(ctx context.Context, input *GetVLANInput) (*FabricVLAN, error) {
	fullPath := path.Join("/", c.client.AccountName, "fabrics", "default", "vlans", strconv.Itoa(input.ID))
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to get VLAN")
	}

	var result *FabricVLAN
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode get VLAN response")
	}

	return result, nil
}

type DeleteVLANInput struct {
	ID int `json:"-"`
}

func (c *FabricsClient) DeleteVLAN(ctx context.Context, input *DeleteVLANInput) error {
	fullPath := path.Join("/", c.client.AccountName, "fabrics", "default", "vlans", strconv.Itoa(input.ID))
	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   fullPath,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errors.Wrap(err, "unable to delete VLAN")
	}

	return nil
}

type ListFabricsInput struct {
	FabricVLANID int `json:"-"`
}

func (c *FabricsClient) List(ctx context.Context, input *ListFabricsInput) ([]*Network, error) {
	fullPath := path.Join("/", c.client.AccountName, "fabrics", "default", "vlans", strconv.Itoa(input.FabricVLANID), "networks")
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to list fabrics")
	}

	var result []*Network
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode list fabrics response")
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
	fullPath := path.Join("/", c.client.AccountName, "fabrics", "default", "vlans", strconv.Itoa(input.FabricVLANID), "networks")
	reqInputs := client.RequestInput{
		Method: http.MethodPost,
		Path:   fullPath,
		Body:   input,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to create fabric")
	}

	var result *Network
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode create fabric response")
	}

	return result, nil
}

type GetFabricInput struct {
	FabricVLANID int    `json:"-"`
	NetworkID    string `json:"-"`
}

func (c *FabricsClient) Get(ctx context.Context, input *GetFabricInput) (*Network, error) {
	fullPath := path.Join("/", c.client.AccountName, "fabrics", "default", "vlans", strconv.Itoa(input.FabricVLANID), "networks", input.NetworkID)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to get fabric")
	}

	var result *Network
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode get fabric response")
	}

	return result, nil
}

type DeleteFabricInput struct {
	FabricVLANID int    `json:"-"`
	NetworkID    string `json:"-"`
}

func (c *FabricsClient) Delete(ctx context.Context, input *DeleteFabricInput) error {
	fullPath := path.Join("/", c.client.AccountName, "fabrics", "default", "vlans", strconv.Itoa(input.FabricVLANID), "networks", input.NetworkID)
	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   fullPath,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errors.Wrap(err, "unable to delete fabric")
	}

	return nil
}
