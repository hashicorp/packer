package driver

import "github.com/vmware/govmomi/object"

type Cluster struct {
	driver  *Driver
	cluster *object.ClusterComputeResource
}

func (d *Driver) FindCluster(name string) (*Cluster, error) {
	c, err := d.finder.ClusterComputeResource(d.ctx, name)
	if err != nil {
		return nil, err
	}
	return &Cluster{
		cluster: c,
		driver:  d,
	}, nil
}
