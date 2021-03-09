package api

import (
	"context"
	"fmt"
)

const (
	EndpointURL = "https://api.exoscale.com/"
	Prefix      = "v2.alpha"
)

const defaultReqEndpointEnv = "api"

// ReqEndpoint represents an Exoscale API request endpoint.
type ReqEndpoint struct {
	env  string
	zone string
}

// NewReqEndpoint returns a new Exoscale API request endpoint from an environment and zone.
func NewReqEndpoint(env, zone string) ReqEndpoint {
	re := ReqEndpoint{
		env:  env,
		zone: zone,
	}

	if re.env == "" {
		re.env = defaultReqEndpointEnv
	}

	return re
}

// Env returns the Exoscale API endpoint environment.
func (r *ReqEndpoint) Env() string {
	return r.env
}

// Zone returns the Exoscale API endpoint zone.
func (r *ReqEndpoint) Zone() string {
	return r.zone
}

// Host returns the Exoscale API endpoint host FQDN.
func (r *ReqEndpoint) Host() string {
	return fmt.Sprintf("%s-%s.exoscale.com", r.env, r.zone)
}

// WithEndpoint returns an augmented context instance containing the Exoscale endpoint to send
// the request to.
func WithEndpoint(ctx context.Context, endpoint ReqEndpoint) context.Context {
	return context.WithValue(ctx, ReqEndpoint{}, endpoint)
}

// WithZone is a shorthand to WithEndpoint where only the endpoint zone has to be specified.
// If a request endpoint is already set in the specified context instance, the value currently
// set for the environment will be reused.
func WithZone(ctx context.Context, zone string) context.Context {
	var env string

	if v, ok := ctx.Value(ReqEndpoint{}).(ReqEndpoint); ok {
		env = v.Env()
	}

	return WithEndpoint(ctx, NewReqEndpoint(env, zone))
}
