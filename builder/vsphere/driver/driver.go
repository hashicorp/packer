package driver

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
)

type Driver struct {
	// context that controls the authenticated sessions used to run the VM commands
	ctx        context.Context
	client     *govmomi.Client
	vimClient  *vim25.Client
	restClient *RestClient
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

	vcenterUrl, err := url.Parse(fmt.Sprintf("https://%v/sdk", config.VCenterServer))
	if err != nil {
		return nil, err
	}
	credentials := url.UserPassword(config.Username, config.Password)
	vcenterUrl.User = credentials

	soapClient := soap.NewClient(vcenterUrl, config.InsecureConnection)
	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, err
	}

	vimClient.RoundTripper = session.KeepAlive(vimClient.RoundTripper, 10*time.Minute)
	client := &govmomi.Client{
		Client:         vimClient,
		SessionManager: session.NewManager(vimClient),
	}

	err = client.SessionManager.Login(ctx, credentials)
	if err != nil {
		return nil, err
	}

	finder := find.NewFinder(client.Client, false)
	datacenter, err := finder.DatacenterOrDefault(ctx, config.Datacenter)
	if err != nil {
		return nil, err
	}
	finder.SetDatacenter(datacenter)

	d := Driver{
		ctx:       ctx,
		client:    client,
		vimClient: vimClient,
		restClient: &RestClient{
			client:      rest.NewClient(vimClient),
			credentials: credentials,
		},
		datacenter: datacenter,
		finder:     finder,
	}
	return &d, nil
}

// The rest.Client requires vCenter.
// RestClient is to modularize the rest.Client session and use it only when is necessary.
// This will allow users without vCenter to use the other features that doesn't use the rest.Client.
// To use the client login/logout must be done to create an authenticated session.
type RestClient struct {
	client      *rest.Client
	credentials *url.Userinfo
}

func (r *RestClient) Login(ctx context.Context) error {
	return r.client.Login(ctx, r.credentials)
}

func (r *RestClient) Logout(ctx context.Context) error {
	return r.client.Logout(ctx)
}
