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

type VCenterSimulator struct {
	model  *simulator.Model
	server *simulator.Server
	driver *VCenterDriver
}

func NewCustomVCenterSimulator(model *simulator.Model) (*VCenterSimulator, error) {
	sim := new(VCenterSimulator)
	sim.model = model

	server, err := sim.NewSimulatorServer()
	if err != nil {
		sim.Close()
		return nil, err
	}
	sim.server = server

	driver, err := sim.NewSimulatorDriver()
	if err != nil {
		sim.Close()
		return nil, err
	}
	sim.driver = driver
	return sim, nil
}

func NewVCenterSimulator() (*VCenterSimulator, error) {
	model := simulator.VPX()
	model.Machine = 1
	return NewCustomVCenterSimulator(model)
}

func (s *VCenterSimulator) Close() {
	if s.model != nil {
		s.model.Remove()
	}
	if s.server != nil {
		s.server.Close()
	}
}

//Simulator shortcut to choose any pre created VM.
func (s *VCenterSimulator) ChooseSimulatorPreCreatedVM() (VirtualMachine, *simulator.VirtualMachine) {
	machine := simulator.Map.Any("VirtualMachine").(*simulator.VirtualMachine)
	ref := machine.Reference()
	vm := s.driver.NewVM(&ref)
	return vm, machine
}

//Simulator shortcut to choose any pre created Datastore.
func (s *VCenterSimulator) ChooseSimulatorPreCreatedDatastore() (Datastore, *simulator.Datastore) {
	ds := simulator.Map.Any("Datastore").(*simulator.Datastore)
	ref := ds.Reference()
	datastore := s.driver.NewDatastore(&ref)
	return datastore, ds
}

//Simulator shortcut to choose any pre created Host.
func (s *VCenterSimulator) ChooseSimulatorPreCreatedHost() (*Host, *simulator.HostSystem) {
	h := simulator.Map.Any("HostSystem").(*simulator.HostSystem)
	ref := h.Reference()
	host := s.driver.NewHost(&ref)
	return host, h
}

func (s *VCenterSimulator) NewSimulatorServer() (*simulator.Server, error) {
	err := s.model.Create()
	if err != nil {
		return nil, err
	}

	s.model.Service.RegisterEndpoints = true
	s.model.Service.TLS = new(tls.Config)
	s.model.Service.ServeMux = http.NewServeMux()
	return s.model.Service.NewServer(), nil
}

func (s *VCenterSimulator) NewSimulatorDriver() (*VCenterDriver, error) {
	ctx := context.TODO()
	user := &url.Userinfo{}
	s.server.URL.User = user

	soapClient := soap.NewClient(s.server.URL, true)
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
