package driver

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
)

// Defines whether acceptance tests should be run
const TestHostName = "esxi-1.vsphere65.test"

func newTestDriver(t *testing.T) Driver {
	username := os.Getenv("VSPHERE_USERNAME")
	if username == "" {
		username = "root"
	}
	password := os.Getenv("VSPHERE_PASSWORD")
	if password == "" {
		password = "jetbrains"
	}

	d, err := NewDriver(&ConnectConfig{
		VCenterServer:      "vcenter.vsphere65.test",
		Username:           username,
		Password:           password,
		InsecureConnection: true,
	})
	if err != nil {
		t.Fatalf("Cannot connect: %v", err)
	}
	return d
}

func newVMName() string {
	rand.Seed(time.Now().UTC().UnixNano())
	return fmt.Sprintf("test-%v", rand.Intn(1000))
}

func NewSimulatorServer(model *simulator.Model) (*simulator.Server, error) {
	err := model.Create()
	if err != nil {
		return nil, err
	}

	model.Service.RegisterEndpoints = true
	model.Service.TLS = new(tls.Config)
	model.Service.ServeMux = http.NewServeMux()
	return model.Service.NewServer(), nil
}

func NewSimulatorDriver(s *simulator.Server) (*VCenterDriver, error) {
	ctx := context.TODO()
	user := &url.Userinfo{}
	s.URL.User = user

	soapClient := soap.NewClient(s.URL, true)
	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, err
	}

	vimClient.RoundTripper = session.KeepAlive(vimClient.RoundTripper, 10*time.Minute)
	client := &govmomi.Client{
		Client:         vimClient,
		SessionManager: session.NewManager(vimClient),
	}

	err = client.SessionManager.Login(ctx, user)
	if err != nil {
		return nil, err
	}

	finder := find.NewFinder(client.Client, false)
	datacenter, err := finder.DatacenterOrDefault(ctx, "")
	if err != nil {
		return nil, err
	}
	finder.SetDatacenter(datacenter)

	d := &VCenterDriver{
		ctx:       ctx,
		client:    client,
		vimClient: vimClient,
		restClient: &RestClient{
			client:      rest.NewClient(vimClient),
			credentials: user,
		},
		datacenter: datacenter,
		finder:     finder,
	}
	return d, nil
}
