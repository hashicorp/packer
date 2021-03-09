package v2

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	apiv2 "github.com/exoscale/egoscale/v2/api"
	papi "github.com/exoscale/egoscale/v2/internal/public-api"
)

// NetworkLoadBalancerServerStatus represents a Network Load Balancer service target server status.
type NetworkLoadBalancerServerStatus struct {
	InstanceIP net.IP
	Status     string
}

func nlbServerStatusFromAPI(st *papi.LoadBalancerServerStatus) *NetworkLoadBalancerServerStatus {
	return &NetworkLoadBalancerServerStatus{
		InstanceIP: net.ParseIP(papi.OptionalString(st.PublicIp)),
		Status:     papi.OptionalString(st.Status),
	}
}

// NetworkLoadBalancerServiceHealthcheck represents a Network Load Balancer service healthcheck.
type NetworkLoadBalancerServiceHealthcheck struct {
	Mode     string
	Port     uint16
	Interval time.Duration
	Timeout  time.Duration
	Retries  int64
	URI      string
	TLSSNI   string
}

// NetworkLoadBalancerService represents a Network Load Balancer service.
type NetworkLoadBalancerService struct {
	ID                string
	Name              string
	Description       string
	InstancePoolID    string
	Protocol          string
	Port              uint16
	TargetPort        uint16
	Strategy          string
	Healthcheck       NetworkLoadBalancerServiceHealthcheck
	State             string
	HealthcheckStatus []*NetworkLoadBalancerServerStatus
}

func nlbServiceFromAPI(svc *papi.LoadBalancerService) *NetworkLoadBalancerService {
	return &NetworkLoadBalancerService{
		ID:             papi.OptionalString(svc.Id),
		Name:           papi.OptionalString(svc.Name),
		Description:    papi.OptionalString(svc.Description),
		InstancePoolID: papi.OptionalString(svc.InstancePool.Id),
		Protocol:       papi.OptionalString(svc.Protocol),
		Port:           uint16(papi.OptionalInt64(svc.Port)),
		TargetPort:     uint16(papi.OptionalInt64(svc.TargetPort)),
		Strategy:       papi.OptionalString(svc.Strategy),
		Healthcheck: NetworkLoadBalancerServiceHealthcheck{
			Mode:     svc.Healthcheck.Mode,
			Port:     uint16(svc.Healthcheck.Port),
			Interval: time.Duration(papi.OptionalInt64(svc.Healthcheck.Interval)) * time.Second,
			Timeout:  time.Duration(papi.OptionalInt64(svc.Healthcheck.Timeout)) * time.Second,
			Retries:  papi.OptionalInt64(svc.Healthcheck.Retries),
			URI:      papi.OptionalString(svc.Healthcheck.Uri),
			TLSSNI:   papi.OptionalString(svc.Healthcheck.TlsSni),
		},
		HealthcheckStatus: func() []*NetworkLoadBalancerServerStatus {
			statuses := make([]*NetworkLoadBalancerServerStatus, 0)

			if svc.HealthcheckStatus != nil {
				for _, st := range *svc.HealthcheckStatus {
					st := st
					statuses = append(statuses, nlbServerStatusFromAPI(&st))
				}
			}

			return statuses
		}(),
		State: papi.OptionalString(svc.State),
	}
}

// NetworkLoadBalancer represents a Network Load Balancer instance.
type NetworkLoadBalancer struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	IPAddress   net.IP
	Services    []*NetworkLoadBalancerService
	State       string

	c    *Client
	zone string
}

func nlbFromAPI(nlb *papi.LoadBalancer) *NetworkLoadBalancer {
	return &NetworkLoadBalancer{
		ID:          papi.OptionalString(nlb.Id),
		Name:        papi.OptionalString(nlb.Name),
		Description: papi.OptionalString(nlb.Description),
		CreatedAt:   *nlb.CreatedAt,
		IPAddress:   net.ParseIP(papi.OptionalString(nlb.Ip)),
		State:       papi.OptionalString(nlb.State),
		Services: func() []*NetworkLoadBalancerService {
			services := make([]*NetworkLoadBalancerService, 0)

			if nlb.Services != nil {
				for _, svc := range *nlb.Services {
					svc := svc
					services = append(services, nlbServiceFromAPI(&svc))
				}
			}

			return services
		}(),
	}
}

// AddService adds a service to the Network Load Balancer instance.
func (nlb *NetworkLoadBalancer) AddService(ctx context.Context,
	svc *NetworkLoadBalancerService) (*NetworkLoadBalancerService, error) {
	var (
		port                = int64(svc.Port)
		targetPort          = int64(svc.TargetPort)
		healthcheckPort     = int64(svc.Healthcheck.Port)
		healthcheckInterval = int64(svc.Healthcheck.Interval.Seconds())
		healthcheckTimeout  = int64(svc.Healthcheck.Timeout.Seconds())
	)

	// The API doesn't return the NLB service created directly, so in order to return a
	// *NetworkLoadBalancerService corresponding to the new service we have to manually
	// compare the list of services on the NLB instance before and after the service
	// creation, and identify the service that wasn't there before.
	// Note: in case of multiple services creation in parallel this technique is subject
	// to race condition as we could return an unrelated service. To prevent this, we
	// also compare the name of the new service to the name specified in the svc
	// parameter.
	services := make(map[string]struct{})
	for _, svc := range nlb.Services {
		services[svc.ID] = struct{}{}
	}

	resp, err := nlb.c.AddServiceToLoadBalancerWithResponse(
		apiv2.WithZone(ctx, nlb.zone),
		nlb.ID,
		papi.AddServiceToLoadBalancerJSONRequestBody{
			Name:        svc.Name,
			Description: &svc.Description,
			Healthcheck: papi.Healthcheck{
				Mode:     svc.Healthcheck.Mode,
				Port:     healthcheckPort,
				Interval: &healthcheckInterval,
				Timeout:  &healthcheckTimeout,
				Retries:  &svc.Healthcheck.Retries,
				Uri: func() *string {
					if strings.HasPrefix(svc.Healthcheck.Mode, "http") {
						return &svc.Healthcheck.URI
					}
					return nil
				}(),
				TlsSni: func() *string {
					if svc.Healthcheck.Mode == "https" && svc.Healthcheck.TLSSNI != "" {
						return &svc.Healthcheck.TLSSNI
					}
					return nil
				}(),
			},
			InstancePool: papi.InstancePool{Id: &svc.InstancePoolID},
			Port:         port,
			TargetPort:   targetPort,
			Protocol:     svc.Protocol,
			Strategy:     svc.Strategy,
		})
	if err != nil {
		return nil, err
	}

	res, err := papi.NewPoller().
		WithTimeout(nlb.c.timeout).
		Poll(ctx, nlb.c.OperationPoller(nlb.zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	nlbUpdated, err := nlb.c.GetNetworkLoadBalancer(ctx, nlb.zone, *res.(*papi.Reference).Id)
	if err != nil {
		return nil, err
	}

	// Look for an unknown service: if we find one we hope it's the one we've just created.
	for _, s := range nlbUpdated.Services {
		if _, ok := services[svc.ID]; !ok && s.Name == svc.Name {
			return s, nil
		}
	}

	return nil, errors.New("unable to identify the service created")
}

// UpdateService updates the specified Network Load Balancer service.
func (nlb *NetworkLoadBalancer) UpdateService(ctx context.Context, svc *NetworkLoadBalancerService) error {
	var (
		port                = int64(svc.Port)
		targetPort          = int64(svc.TargetPort)
		healthcheckPort     = int64(svc.Healthcheck.Port)
		healthcheckInterval = int64(svc.Healthcheck.Interval.Seconds())
		healthcheckTimeout  = int64(svc.Healthcheck.Timeout.Seconds())
	)

	resp, err := nlb.c.UpdateLoadBalancerServiceWithResponse(
		apiv2.WithZone(ctx, nlb.zone),
		nlb.ID,
		svc.ID,
		papi.UpdateLoadBalancerServiceJSONRequestBody{
			Name:        &svc.Name,
			Description: &svc.Description,
			Port:        &port,
			TargetPort:  &targetPort,
			Protocol:    &svc.Protocol,
			Strategy:    &svc.Strategy,
			Healthcheck: &papi.Healthcheck{
				Mode:     svc.Healthcheck.Mode,
				Port:     healthcheckPort,
				Interval: &healthcheckInterval,
				Timeout:  &healthcheckTimeout,
				Retries:  &svc.Healthcheck.Retries,
				Uri: func() *string {
					if strings.HasPrefix(svc.Healthcheck.Mode, "http") {
						return &svc.Healthcheck.URI
					}
					return nil
				}(),
				TlsSni: func() *string {
					if svc.Healthcheck.Mode == "https" && svc.Healthcheck.TLSSNI != "" {
						return &svc.Healthcheck.TLSSNI
					}
					return nil
				}(),
			},
		})
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(nlb.c.timeout).
		Poll(ctx, nlb.c.OperationPoller(nlb.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// DeleteService deletes the specified service from the Network Load Balancer instance.
func (nlb *NetworkLoadBalancer) DeleteService(ctx context.Context, svc *NetworkLoadBalancerService) error {
	resp, err := nlb.c.DeleteLoadBalancerServiceWithResponse(
		apiv2.WithZone(ctx, nlb.zone),
		nlb.ID,
		svc.ID,
	)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(nlb.c.timeout).
		Poll(ctx, nlb.c.OperationPoller(nlb.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// CreateNetworkLoadBalancer creates a Network Load Balancer instance in the specified zone.
func (c *Client) CreateNetworkLoadBalancer(ctx context.Context, zone string,
	nlb *NetworkLoadBalancer) (*NetworkLoadBalancer, error) {
	resp, err := c.CreateLoadBalancerWithResponse(
		apiv2.WithZone(ctx, zone),
		papi.CreateLoadBalancerJSONRequestBody{
			Name:        nlb.Name,
			Description: &nlb.Description,
		})
	if err != nil {
		return nil, err
	}

	res, err := papi.NewPoller().
		WithTimeout(c.timeout).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	return c.GetNetworkLoadBalancer(ctx, zone, *res.(*papi.Reference).Id)
}

// ListNetworkLoadBalancers returns the list of existing Network Load Balancers in the
// specified zone.
func (c *Client) ListNetworkLoadBalancers(ctx context.Context, zone string) ([]*NetworkLoadBalancer, error) {
	list := make([]*NetworkLoadBalancer, 0)

	resp, err := c.ListLoadBalancersWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return nil, err
	}

	if resp.JSON200.LoadBalancers != nil {
		for i := range *resp.JSON200.LoadBalancers {
			nlb := nlbFromAPI(&(*resp.JSON200.LoadBalancers)[i])
			nlb.c = c
			nlb.zone = zone

			list = append(list, nlb)
		}
	}

	return list, nil
}

// GetNetworkLoadBalancer returns the Network Load Balancer instance corresponding to the
// specified ID in the specified zone.
func (c *Client) GetNetworkLoadBalancer(ctx context.Context, zone, id string) (*NetworkLoadBalancer, error) {
	resp, err := c.GetLoadBalancerWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return nil, err
	}

	nlb := nlbFromAPI(resp.JSON200)
	nlb.c = c
	nlb.zone = zone

	return nlb, nil
}

// UpdateNetworkLoadBalancer updates the specified Network Load Balancer instance in the specified zone.
func (c *Client) UpdateNetworkLoadBalancer(ctx context.Context, zone string, // nolint:dupl
	nlb *NetworkLoadBalancer) (*NetworkLoadBalancer, error) {
	resp, err := c.UpdateLoadBalancerWithResponse(
		apiv2.WithZone(ctx, zone),
		nlb.ID,
		papi.UpdateLoadBalancerJSONRequestBody{
			Name:        &nlb.Name,
			Description: &nlb.Description,
		})
	if err != nil {
		return nil, err
	}

	res, err := papi.NewPoller().
		WithTimeout(c.timeout).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	return c.GetNetworkLoadBalancer(ctx, zone, *res.(*papi.Reference).Id)
}

// DeleteNetworkLoadBalancer deletes the specified Network Load Balancer instance in the specified zone.
func (c *Client) DeleteNetworkLoadBalancer(ctx context.Context, zone, id string) error {
	resp, err := c.DeleteLoadBalancerWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(c.timeout).
		Poll(ctx, c.OperationPoller(zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}
