package driver

import (
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"context"
	"net/url"
	"fmt"
	"github.com/vmware/govmomi/object"
	"time"
	"github.com/vmware/govmomi/session"
)

type Driver struct {
	ctx        context.Context
	client     *govmomi.Client
	finder     *find.Finder
	datacenter *object.Datacenter
}

type ConnectConfig struct {
	VCenterServer      string
	Username           string
	Password           string
	InsecureConnection bool
	Datacenter         string
}

func NewDriver(config *ConnectConfig) (*Driver, error) {
	ctx := context.TODO()

	vcenter_url, err := url.Parse(fmt.Sprintf("https://%v/sdk", config.VCenterServer))
	if err != nil {
		return nil, err
	}
	vcenter_url.User = url.UserPassword(config.Username, config.Password)

	client, err := govmomi.NewClient(ctx, vcenter_url, config.InsecureConnection)
	if err != nil {
		return nil, err
	}
	client.RoundTripper = session.KeepAlive(client.RoundTripper, 10*time.Minute)

	finder := find.NewFinder(client.Client, false)
	datacenter, err := finder.DatacenterOrDefault(ctx, config.Datacenter)
	if err != nil {
		return nil, err
	}
	finder.SetDatacenter(datacenter)

	d := Driver{
		ctx:        ctx,
		client:     client,
		datacenter: datacenter,
		finder:     finder,
	}
	return &d, nil
}
