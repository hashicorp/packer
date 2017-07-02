package main

import (
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"context"
	"net/url"
	"fmt"
	"github.com/vmware/govmomi/object"
)

type Driver struct {
	ctx        context.Context
	client     *govmomi.Client
	datacenter *object.Datacenter
	finder     *find.Finder
}

func NewDriverVSphere(config *ConnectConfig) (Driver, error) {
	ctx := context.TODO()

	vcenter_url, err := url.Parse(fmt.Sprintf("https://%v/sdk", config.VCenterServer))
	if err != nil {
		return Driver{}, err
	}
	vcenter_url.User = url.UserPassword(config.Username, config.Password)
	client, err := govmomi.NewClient(ctx, vcenter_url, config.InsecureConnection)
	if err != nil {
		return Driver{}, err
	}

	finder := find.NewFinder(client.Client, false)
	datacenter, err := finder.DatacenterOrDefault(ctx, config.Datacenter)
	if err != nil {
		return Driver{}, err
	}
	finder.SetDatacenter(datacenter)

	d := Driver{
		ctx:        ctx,
		client:     client,
		datacenter: datacenter,
		finder:     finder,
	}
	return d, nil
}
