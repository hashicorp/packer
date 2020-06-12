package linodego

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/linode/linodego/internal/parseabletime"
)

// FirewallStatus enum type
type FirewallStatus string

// FirewallStatus enums start with Firewall
const (
	FirewallEnabled  FirewallStatus = "enabled"
	FirewallDisabled FirewallStatus = "disabled"
	FirewallDeleted  FirewallStatus = "deleted"
)

// A Firewall is a set of networking rules (iptables) applied to Devices with which it is associated
type Firewall struct {
	ID      int             `json:"id"`
	Label   string          `json:"label"`
	Status  FirewallStatus  `json:"status"`
	Tags    []string        `json:"tags,omitempty"`
	Rules   FirewallRuleSet `json:"rules"`
	Created *time.Time      `json:"-"`
	Updated *time.Time      `json:"-"`
}

// DevicesCreationOptions fields are used when adding devices during the Firewall creation process.
type DevicesCreationOptions struct {
	Linodes       []string `json:"linodes,omitempty"`
	NodeBalancers []string `json:"nodebalancers,omitempty"`
}

// FirewallCreateOptions fields are those accepted by CreateFirewall
type FirewallCreateOptions struct {
	Label   string                 `json:"label,omitempty"`
	Rules   FirewallRuleSet        `json:"rules"`
	Tags    []string               `json:"tags,omitempty"`
	Devices DevicesCreationOptions `json:"devices,omitempty"`
}

// UnmarshalJSON for Firewall responses
func (i *Firewall) UnmarshalJSON(b []byte) error {
	type Mask Firewall

	p := struct {
		*Mask
		Created *parseabletime.ParseableTime `json:"created"`
		Updated *parseabletime.ParseableTime `json:"updated"`
	}{
		Mask: (*Mask)(i),
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return err
	}

	i.Created = (*time.Time)(p.Created)
	i.Updated = (*time.Time)(p.Updated)

	return nil
}

// FirewallsPagedResponse represents a Linode API response for listing of Cloud Firewalls
type FirewallsPagedResponse struct {
	*PageOptions
	Data []Firewall `json:"data"`
}

func (FirewallsPagedResponse) endpoint(c *Client) string {
	endpoint, err := c.Firewalls.Endpoint()
	if err != nil {
		panic(err)
	}
	return endpoint
}

func (resp *FirewallsPagedResponse) appendData(r *FirewallsPagedResponse) {
	resp.Data = append(resp.Data, r.Data...)
}

// ListFirewalls returns a paginated list of Cloud Firewalls
func (c *Client) ListFirewalls(ctx context.Context, opts *ListOptions) ([]Firewall, error) {
	response := FirewallsPagedResponse{}

	err := c.listHelper(ctx, &response, opts)

	if err != nil {
		return nil, err
	}

	return response.Data, nil
}

// CreateFirewall creates a single Firewall with at least one set of inbound or outbound rules
func (c *Client) CreateFirewall(ctx context.Context, createOpts FirewallCreateOptions) (*Firewall, error) {
	var body string
	e, err := c.Firewalls.Endpoint()
	if err != nil {
		return nil, err
	}

	req := c.R(ctx).SetResult(&Firewall{})

	if bodyData, err := json.Marshal(createOpts); err == nil {
		body = string(bodyData)
	} else {
		return nil, NewError(err)
	}

	r, err := coupleAPIErrors(req.
		SetBody(body).
		Post(e))

	if err != nil {
		return nil, err
	}
	return r.Result().(*Firewall), nil
}

// DeleteFirewall deletes a single Firewall with the provided ID
func (c *Client) DeleteFirewall(ctx context.Context, id int) error {
	e, err := c.Firewalls.Endpoint()
	if err != nil {
		return err
	}

	req := c.R(ctx)

	e = fmt.Sprintf("%s/%d", e, id)
	_, err = coupleAPIErrors(req.Delete(e))
	return err
}
