//
// Copyright (c) 2018, Joyent, Inc. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package network

import (
	"net/http"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/client"
)

type NetworkClient struct {
	Client *client.Client
}

func newNetworkClient(client *client.Client) *NetworkClient {
	return &NetworkClient{
		Client: client,
	}
}

// NewClient returns a new client for working with Network endpoints and
// resources within CloudAPI
func NewClient(config *triton.ClientConfig) (*NetworkClient, error) {
	// TODO: Utilize config interface within the function itself
	client, err := client.New(
		config.TritonURL,
		config.MantaURL,
		config.AccountName,
		config.Signers...,
	)
	if err != nil {
		return nil, err
	}
	return newNetworkClient(client), nil
}

// SetHeaders allows a consumer of the current client to set custom headers for
// the next backend HTTP request sent to CloudAPI
func (c *NetworkClient) SetHeader(header *http.Header) {
	c.Client.RequestHeader = header
}

// Fabrics returns a FabricsClient used for accessing functions pertaining to
// Fabric functionality in the Triton API.
func (c *NetworkClient) Fabrics() *FabricsClient {
	return &FabricsClient{c.Client}
}

// Firewall returns a FirewallClient client used for accessing functions
// pertaining to firewall functionality in the Triton API.
func (c *NetworkClient) Firewall() *FirewallClient {
	return &FirewallClient{c.Client}
}
