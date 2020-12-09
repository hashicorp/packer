package rpc

import (
	"context"
	"reflect"
	"testing"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func TestProvisionerRPC(t *testing.T) {
	topCtx, topCtxCancel := context.WithCancel(context.Background())

	// Create the interface to test
	p := new(packersdk.MockProvisioner)
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
	comm := &packersdk.MockCommunicator{}
	if err := pClient.Provision(topCtx, ui, comm, make(map[string]interface{})); err == nil {
		t.Fatalf("Provison should have err")
	}
	if !p.ProvCalled {
		t.Fatal("should be called")
	}

}

func TestProvisioner_Implements(t *testing.T) {
	var _ packersdk.Provisioner = new(provisioner)
}
