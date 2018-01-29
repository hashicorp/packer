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
	"time"

	"github.com/joyent/triton-go/client"
	"github.com/pkg/errors"
)

type FirewallClient struct {
	client *client.Client
}

// FirewallRule represents a firewall rule
type FirewallRule struct {
	// ID is a unique identifier for this rule
	ID string `json:"id"`

	// Enabled indicates if the rule is enabled
	Enabled bool `json:"enabled"`

	// Rule is the firewall rule text
	Rule string `json:"rule"`

	// Global indicates if the rule is global. Optional.
	Global bool `json:"global"`

	// Description is a human-readable description for the rule. Optional
	Description string `json:"description"`
}

type ListRulesInput struct{}

func (c *FirewallClient) ListRules(ctx context.Context, _ *ListRulesInput) ([]*FirewallRule, error) {
	fullPath := path.Join("/", c.client.AccountName, "fwrules")
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to list firewall rules")
	}

	var result []*FirewallRule
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode list firewall rules response")
	}

	return result, nil
}

type GetRuleInput struct {
	ID string
}

func (c *FirewallClient) GetRule(ctx context.Context, input *GetRuleInput) (*FirewallRule, error) {
	fullPath := path.Join("/", c.client.AccountName, "fwrules", input.ID)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to get firewall rule")
	}

	var result *FirewallRule
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode get firewall rule response")
	}

	return result, nil
}

type CreateRuleInput struct {
	Enabled     bool   `json:"enabled"`
	Rule        string `json:"rule"`
	Description string `json:"description,omitempty"`
}

func (c *FirewallClient) CreateRule(ctx context.Context, input *CreateRuleInput) (*FirewallRule, error) {
	fullPath := path.Join("/", c.client.AccountName, "fwrules")
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
		return nil, errors.Wrap(err, "unable to create firewall rule")
	}

	var result *FirewallRule
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode create firewall rule response")
	}

	return result, nil
}

type UpdateRuleInput struct {
	ID          string `json:"-"`
	Enabled     bool   `json:"enabled"`
	Rule        string `json:"rule"`
	Description string `json:"description,omitempty"`
}

func (c *FirewallClient) UpdateRule(ctx context.Context, input *UpdateRuleInput) (*FirewallRule, error) {
	fullPath := path.Join("/", c.client.AccountName, "fwrules", input.ID)
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
		return nil, errors.Wrap(err, "unable to update firewall rule")
	}

	var result *FirewallRule
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode update firewall rule response")
	}

	return result, nil
}

type EnableRuleInput struct {
	ID string `json:"-"`
}

func (c *FirewallClient) EnableRule(ctx context.Context, input *EnableRuleInput) (*FirewallRule, error) {
	fullPath := path.Join("/", c.client.AccountName, "fwrules", input.ID, "enable")
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
		return nil, errors.Wrap(err, "unable to enable firewall rule")
	}

	var result *FirewallRule
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode enable firewall rule response")
	}

	return result, nil
}

type DisableRuleInput struct {
	ID string `json:"-"`
}

func (c *FirewallClient) DisableRule(ctx context.Context, input *DisableRuleInput) (*FirewallRule, error) {
	fullPath := path.Join("/", c.client.AccountName, "fwrules", input.ID, "disable")
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
		return nil, errors.Wrap(err, "unable to disable firewall rule")
	}

	var result *FirewallRule
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode disable firewall rule response")
	}

	return result, nil
}

type DeleteRuleInput struct {
	ID string
}

func (c *FirewallClient) DeleteRule(ctx context.Context, input *DeleteRuleInput) error {
	fullPath := path.Join("/", c.client.AccountName, "fwrules", input.ID)
	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   fullPath,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errors.Wrap(err, "unable to delete firewall rule")
	}

	return nil
}

type ListMachineRulesInput struct {
	MachineID string
}

func (c *FirewallClient) ListMachineRules(ctx context.Context, input *ListMachineRulesInput) ([]*FirewallRule, error) {
	fullPath := path.Join("/", c.client.AccountName, "machines", input.MachineID, "fwrules")
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to list machine firewall rules")
	}

	var result []*FirewallRule
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode list machine firewall rules response")
	}

	return result, nil
}

type ListRuleMachinesInput struct {
	ID string
}

type Machine struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Type            string                 `json:"type"`
	Brand           string                 `json:"brand"`
	State           string                 `json:"state"`
	Image           string                 `json:"image"`
	Memory          int                    `json:"memory"`
	Disk            int                    `json:"disk"`
	Metadata        map[string]string      `json:"metadata"`
	Tags            map[string]interface{} `json:"tags"`
	Created         time.Time              `json:"created"`
	Updated         time.Time              `json:"updated"`
	Docker          bool                   `json:"docker"`
	IPs             []string               `json:"ips"`
	Networks        []string               `json:"networks"`
	PrimaryIP       string                 `json:"primaryIp"`
	FirewallEnabled bool                   `json:"firewall_enabled"`
	ComputeNode     string                 `json:"compute_node"`
	Package         string                 `json:"package"`
}

func (c *FirewallClient) ListRuleMachines(ctx context.Context, input *ListRuleMachinesInput) ([]*Machine, error) {
	fullPath := path.Join("/", c.client.AccountName, "fwrules", input.ID, "machines")
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   fullPath,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errors.Wrap(err, "unable to list firewall rule machines")
	}

	var result []*Machine
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errors.Wrap(err, "unable to decode list firewall rule machines response")
	}

	return result, nil
}
