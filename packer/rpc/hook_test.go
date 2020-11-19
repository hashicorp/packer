package rpc

import (
	"context"
	"testing"

	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func TestHook_Implements(t *testing.T) {
	var _ packersdk.Hook = new(hook)
}

func TestHook_cancelWhileRun(t *testing.T) {
	topCtx, cancelTopCtx := context.WithCancel(context.Background())

	h := &packer.MockHook{
		RunFunc: func(ctx context.Context) error {
			cancelTopCtx()
			<-ctx.Done()
			return ctx.Err()
		},
	}

	// Serve
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterHook(h)
	hClient := client.Hook()

	// Start the run
	err := hClient.Run(topCtx, "foo", nil, nil, nil)

	if err == nil {
		t.Fatal("should have errored")
	}
}
