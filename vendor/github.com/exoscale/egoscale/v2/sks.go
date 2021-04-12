package v2

import (
	"context"
	"errors"
	"fmt"
	"time"

	apiv2 "github.com/exoscale/egoscale/v2/api"
	papi "github.com/exoscale/egoscale/v2/internal/public-api"
)

// SKSNodepool represents a SKS Nodepool.
type SKSNodepool struct {
	ID                   string
	Name                 string
	Description          string
	CreatedAt            time.Time
	InstancePoolID       string
	InstanceTypeID       string
	TemplateID           string
	DiskSize             int64
	AntiAffinityGroupIDs []string
	SecurityGroupIDs     []string
	Version              string
	Size                 int64
	State                string
}

func sksNodepoolFromAPI(n *papi.SksNodepool) *SKSNodepool {
	return &SKSNodepool{
		ID:             papi.OptionalString(n.Id),
		Name:           papi.OptionalString(n.Name),
		Description:    papi.OptionalString(n.Description),
		CreatedAt:      *n.CreatedAt,
		InstancePoolID: papi.OptionalString(n.InstancePool.Id),
		InstanceTypeID: papi.OptionalString(n.InstanceType.Id),
		TemplateID:     papi.OptionalString(n.Template.Id),
		DiskSize:       papi.OptionalInt64(n.DiskSize),
		AntiAffinityGroupIDs: func() []string {
			aags := make([]string, 0)

			if n.AntiAffinityGroups != nil {
				for _, aag := range *n.AntiAffinityGroups {
					aag := aag
					aags = append(aags, *aag.Id)
				}
			}

			return aags
		}(),
		SecurityGroupIDs: func() []string {
			sgs := make([]string, 0)

			if n.SecurityGroups != nil {
				for _, sg := range *n.SecurityGroups {
					sg := sg
					sgs = append(sgs, *sg.Id)
				}
			}

			return sgs
		}(),
		Version: papi.OptionalString(n.Version),
		Size:    papi.OptionalInt64(n.Size),
		State:   papi.OptionalString(n.State),
	}
}

// SKSCluster represents a SKS cluster.
type SKSCluster struct {
	ID           string
	Name         string
	Description  string
	CreatedAt    time.Time
	Endpoint     string
	Nodepools    []*SKSNodepool
	Version      string
	ServiceLevel string
	CNI          string
	AddOns       []string
	State        string

	c    *Client
	zone string
}

func sksClusterFromAPI(c *papi.SksCluster) *SKSCluster {
	return &SKSCluster{
		ID:          papi.OptionalString(c.Id),
		Name:        papi.OptionalString(c.Name),
		Description: papi.OptionalString(c.Description),
		CreatedAt:   *c.CreatedAt,
		Endpoint:    papi.OptionalString(c.Endpoint),
		Nodepools: func() []*SKSNodepool {
			nodepools := make([]*SKSNodepool, 0)

			if c.Nodepools != nil {
				for _, n := range *c.Nodepools {
					n := n
					nodepools = append(nodepools, sksNodepoolFromAPI(&n))
				}
			}

			return nodepools
		}(),
		Version:      papi.OptionalString(c.Version),
		ServiceLevel: papi.OptionalString(c.Level),
		CNI:          papi.OptionalString(c.Cni),
		AddOns: func() []string {
			addOns := make([]string, 0)
			if c.Addons != nil {
				addOns = append(addOns, *c.Addons...)
			}
			return addOns
		}(),
		State: papi.OptionalString(c.State),
	}
}

// RequestKubeconfig returns a base64-encoded kubeconfig content for the specified user name,
// optionally associated to specified groups for a duration d (default API-set TTL applies if not specified).
// Fore more information: https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/
func (c *SKSCluster) RequestKubeconfig(ctx context.Context, user string, groups []string,
	d time.Duration) (string, error) {
	if user == "" {
		return "", errors.New("user not specified")
	}

	resp, err := c.c.GenerateSksClusterKubeconfigWithResponse(
		apiv2.WithZone(ctx, c.zone),
		c.ID,
		papi.GenerateSksClusterKubeconfigJSONRequestBody{
			User:   &user,
			Groups: &groups,
			Ttl: func() *int64 {
				ttl := int64(d.Seconds())
				if ttl > 0 {
					return &ttl
				}
				return nil
			}(),
		})
	if err != nil {
		return "", err
	}

	return papi.OptionalString(resp.JSON200.Kubeconfig), nil
}

// AddNodepool adds a Nodepool to the SKS cluster.
func (c *SKSCluster) AddNodepool(ctx context.Context, np *SKSNodepool) (*SKSNodepool, error) {
	resp, err := c.c.CreateSksNodepoolWithResponse(
		apiv2.WithZone(ctx, c.zone),
		c.ID,
		papi.CreateSksNodepoolJSONRequestBody{
			Description:  &np.Description,
			DiskSize:     np.DiskSize,
			InstanceType: papi.InstanceType{Id: &np.InstanceTypeID},
			Name:         np.Name,
			AntiAffinityGroups: func() *[]papi.AntiAffinityGroup {
				aags := make([]papi.AntiAffinityGroup, len(np.AntiAffinityGroupIDs))
				for i, aagID := range np.AntiAffinityGroupIDs {
					aagID := aagID
					aags[i] = papi.AntiAffinityGroup{Id: &aagID}
				}
				return &aags
			}(),
			SecurityGroups: func() *[]papi.SecurityGroup {
				sgs := make([]papi.SecurityGroup, len(np.SecurityGroupIDs))
				for i, sgID := range np.SecurityGroupIDs {
					sgID := sgID
					sgs[i] = papi.SecurityGroup{Id: &sgID}
				}
				return &sgs
			}(),
			Size: np.Size,
		})
	if err != nil {
		return nil, err
	}

	res, err := papi.NewPoller().
		WithTimeout(c.c.timeout).
		Poll(ctx, c.c.OperationPoller(c.zone, *resp.JSON200.Id))
	if err != nil {
		return nil, err
	}

	nodepoolRes, err := c.c.GetSksNodepoolWithResponse(ctx, c.ID, *res.(*papi.Reference).Id)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Nodepool: %s", err)
	}

	return sksNodepoolFromAPI(nodepoolRes.JSON200), nil
}

// UpdateNodepool updates the specified SKS cluster Nodepool.
func (c *SKSCluster) UpdateNodepool(ctx context.Context, np *SKSNodepool) error {
	resp, err := c.c.UpdateSksNodepoolWithResponse(
		apiv2.WithZone(ctx, c.zone),
		c.ID,
		np.ID,
		papi.UpdateSksNodepoolJSONRequestBody{
			Name:         &np.Name,
			Description:  &np.Description,
			InstanceType: &papi.InstanceType{Id: &np.InstanceTypeID},
			DiskSize:     &np.DiskSize,
		})
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(c.c.timeout).
		Poll(ctx, c.c.OperationPoller(c.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// ScaleNodepool scales the SKS cluster Nodepool to the specified number of Kubernetes Nodes.
func (c *SKSCluster) ScaleNodepool(ctx context.Context, np *SKSNodepool, nodes int64) error {
	resp, err := c.c.ScaleSksNodepoolWithResponse(
		apiv2.WithZone(ctx, c.zone),
		c.ID,
		np.ID,
		papi.ScaleSksNodepoolJSONRequestBody{Size: nodes},
	)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(c.c.timeout).
		Poll(ctx, c.c.OperationPoller(c.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// EvictNodepoolMembers evicts the specified members (identified by their Compute instance ID) from the
// SKS cluster Nodepool.
func (c *SKSCluster) EvictNodepoolMembers(ctx context.Context, np *SKSNodepool, members []string) error {
	resp, err := c.c.EvictSksNodepoolMembersWithResponse(
		apiv2.WithZone(ctx, c.zone),
		c.ID,
		np.ID,
		papi.EvictSksNodepoolMembersJSONRequestBody{Instances: &members},
	)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(c.c.timeout).
		Poll(ctx, c.c.OperationPoller(c.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// DeleteNodepool deletes the specified Nodepool from the SKS cluster.
func (c *SKSCluster) DeleteNodepool(ctx context.Context, np *SKSNodepool) error {
	resp, err := c.c.DeleteSksNodepoolWithResponse(
		apiv2.WithZone(ctx, c.zone),
		c.ID,
		np.ID,
	)
	if err != nil {
		return err
	}

	_, err = papi.NewPoller().
		WithTimeout(c.c.timeout).
		Poll(ctx, c.c.OperationPoller(c.zone, *resp.JSON200.Id))
	if err != nil {
		return err
	}

	return nil
}

// CreateSKSCluster creates a SKS cluster in the specified zone.
func (c *Client) CreateSKSCluster(ctx context.Context, zone string, cluster *SKSCluster) (*SKSCluster, error) {
	resp, err := c.CreateSksClusterWithResponse(
		apiv2.WithZone(ctx, zone),
		papi.CreateSksClusterJSONRequestBody{
			Name:        cluster.Name,
			Description: &cluster.Description,
			Version:     cluster.Version,
			Level:       cluster.ServiceLevel,
			Cni: func() *string {
				var cni *string
				if cluster.CNI != "" {
					cni = &cluster.CNI
				}
				return cni
			}(),
			Addons: func() *[]string {
				var addOns *[]string
				if len(cluster.AddOns) > 0 {
					addOns = &cluster.AddOns
				}
				return addOns
			}(),
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

	return c.GetSKSCluster(ctx, zone, *res.(*papi.Reference).Id)
}

// ListSKSClusters returns the list of existing SKS clusters in the specified zone.
func (c *Client) ListSKSClusters(ctx context.Context, zone string) ([]*SKSCluster, error) {
	list := make([]*SKSCluster, 0)

	resp, err := c.ListSksClustersWithResponse(apiv2.WithZone(ctx, zone))
	if err != nil {
		return nil, err
	}

	if resp.JSON200.SksClusters != nil {
		for i := range *resp.JSON200.SksClusters {
			cluster := sksClusterFromAPI(&(*resp.JSON200.SksClusters)[i])
			cluster.c = c
			cluster.zone = zone

			list = append(list, cluster)
		}
	}

	return list, nil
}

// ListSKSClusterVersions returns the list of Kubernetes versions supported during SKS cluster creation.
func (c *Client) ListSKSClusterVersions(ctx context.Context) ([]string, error) {
	list := make([]string, 0)

	resp, err := c.ListSksClusterVersionsWithResponse(ctx)
	if err != nil {
		return nil, err
	}

	if resp.JSON200.SksClusterVersions != nil {
		for i := range *resp.JSON200.SksClusterVersions {
			version := &(*resp.JSON200.SksClusterVersions)[i]
			list = append(list, *version)
		}
	}

	return list, nil
}

// GetSKSCluster returns the SKS cluster corresponding to the specified ID in the specified zone.
func (c *Client) GetSKSCluster(ctx context.Context, zone, id string) (*SKSCluster, error) {
	resp, err := c.GetSksClusterWithResponse(apiv2.WithZone(ctx, zone), id)
	if err != nil {
		return nil, err
	}

	cluster := sksClusterFromAPI(resp.JSON200)
	cluster.c = c
	cluster.zone = zone

	return cluster, nil
}

// UpdateSKSCluster updates the specified SKS cluster in the specified zone.
func (c *Client) UpdateSKSCluster(ctx context.Context, zone string, cluster *SKSCluster) error {
	resp, err := c.UpdateSksClusterWithResponse(
		apiv2.WithZone(ctx, zone),
		cluster.ID,
		papi.UpdateSksClusterJSONRequestBody{
			Name:        &cluster.Name,
			Description: &cluster.Description,
		})
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

// UpgradeSKSCluster upgrades the SKS cluster corresponding to the specified ID in the specified zone to the
// requested Kubernetes version.
func (c *Client) UpgradeSKSCluster(ctx context.Context, zone, id, version string) error {
	resp, err := c.UpgradeSksClusterWithResponse(
		apiv2.WithZone(ctx, zone),
		id,
		papi.UpgradeSksClusterJSONRequestBody{Version: version})
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

// DeleteSKSCluster deletes the specified SKS cluster in the specified zone.
func (c *Client) DeleteSKSCluster(ctx context.Context, zone, id string) error {
	resp, err := c.DeleteSksClusterWithResponse(apiv2.WithZone(ctx, zone), id)
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
