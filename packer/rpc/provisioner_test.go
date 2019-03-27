package rpc

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestProvisionerRPC(t *testing.T) {
	topCtx, topCtxCancel := context.WithCancel(context.Background())

	// Create the interface to test
	p := new(packer.MockProvisioner)
	p.ProvFunc = func(ctx context.Context) error {
		topCtxCancel()
		<-ctx.Done()
		return ctx.Err()
	}

	// Start the server
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterProvisioner(p)
	pClient := client.Provisioner()
	// Test Prepare
	config := 42
	pClient.Prepare(config)
	if !p.PrepCalled {
		t.Fatal("should be called")
	}
	expected := []interface{}{int64(42)}
	if !reflect.DeepEqual(p.PrepConfigs, expected) {
		t.Fatalf("bad: %#v", p.PrepConfigs)
	}

	// Test Provision
	ui := &testUi{}
	comm := &packer.MockCommunicator{}
	if err := pClient.Provision(topCtx, ui, comm); err == nil {
		t.Fatalf("Provison should have err")
	}
	if !p.ProvCalled {
		t.Fatal("should be called")
	}

}

func TestProvisioner_Implements(t *testing.T) {
	var _ packer.Provisioner = new(provisioner)
}
