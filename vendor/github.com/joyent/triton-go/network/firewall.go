package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/errwrap"
	"github.com/joyent/triton-go/client"
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
	path := fmt.Sprintf("/%s/fwrules", c.client.AccountName)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing ListRules request: {{err}}", err)
	}

	var result []*FirewallRule
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding ListRules response: {{err}}", err)
	}

	return result, nil
}

type GetRuleInput struct {
	ID string
}

func (c *FirewallClient) GetRule(ctx context.Context, input *GetRuleInput) (*FirewallRule, error) {
	path := fmt.Sprintf("/%s/fwrules/%s", c.client.AccountName, input.ID)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing GetRule request: {{err}}", err)
	}

	var result *FirewallRule
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding GetRule response: {{err}}", err)
	}

	return result, nil
}

type CreateRuleInput struct {
	Enabled     bool   `json:"enabled"`
	Rule        string `json:"rule"`
	Description string `json:"description,omitempty"`
}

func (c *FirewallClient) CreateRule(ctx context.Context, input *CreateRuleInput) (*FirewallRule, error) {
	path := fmt.Sprintf("/%s/fwrules", c.client.AccountName)
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
		return nil, errwrap.Wrapf("Error executing CreateRule request: {{err}}", err)
	}

	var result *FirewallRule
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding CreateRule response: {{err}}", err)
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
	path := fmt.Sprintf("/%s/fwrules/%s", c.client.AccountName, input.ID)
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
		return nil, errwrap.Wrapf("Error executing UpdateRule request: {{err}}", err)
	}

	var result *FirewallRule
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding UpdateRule response: {{err}}", err)
	}

	return result, nil
}

type EnableRuleInput struct {
	ID string `json:"-"`
}

func (c *FirewallClient) EnableRule(ctx context.Context, input *EnableRuleInput) (*FirewallRule, error) {
	path := fmt.Sprintf("/%s/fwrules/%s/enable", c.client.AccountName, input.ID)
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
		return nil, errwrap.Wrapf("Error executing EnableRule request: {{err}}", err)
	}

	var result *FirewallRule
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding EnableRule response: {{err}}", err)
	}

	return result, nil
}

type DisableRuleInput struct {
	ID string `json:"-"`
}

func (c *FirewallClient) DisableRule(ctx context.Context, input *DisableRuleInput) (*FirewallRule, error) {
	path := fmt.Sprintf("/%s/fwrules/%s/disable", c.client.AccountName, input.ID)
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
		return nil, errwrap.Wrapf("Error executing DisableRule request: {{err}}", err)
	}

	var result *FirewallRule
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding DisableRule response: {{err}}", err)
	}

	return result, nil
}

type DeleteRuleInput struct {
	ID string
}

func (c *FirewallClient) DeleteRule(ctx context.Context, input *DeleteRuleInput) error {
	path := fmt.Sprintf("/%s/fwrules/%s", c.client.AccountName, input.ID)
	reqInputs := client.RequestInput{
		Method: http.MethodDelete,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return errwrap.Wrapf("Error executing DeleteRule request: {{err}}", err)
	}

	return nil
}

type ListMachineRulesInput struct {
	MachineID string
}

func (c *FirewallClient) ListMachineRules(ctx context.Context, input *ListMachineRulesInput) ([]*FirewallRule, error) {
	path := fmt.Sprintf("/%s/machines/%s/firewallrules", c.client.AccountName, input.MachineID)
	reqInputs := client.RequestInput{
		Method: http.MethodGet,
		Path:   path,
	}
	respReader, err := c.client.ExecuteRequest(ctx, reqInputs)
	if respReader != nil {
		defer respReader.Close()
	}
	if err != nil {
		return nil, errwrap.Wrapf("Error executing ListMachineRules request: {{err}}", err)
	}

	var result []*FirewallRule
	decoder := json.NewDecoder(respReader)
	if err = decoder.Decode(&result); err != nil {
		return nil, errwrap.Wrapf("Error decoding ListRules response: {{err}}", err)
	}

	return result, nil
}
