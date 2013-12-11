package rpc

import (
	"github.com/mitchellh/packer/packer"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestHookRPC(t *testing.T) {
	// Create the UI to test
	h := new(packer.MockHook)

	// Serve
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterHook(h)
	hClient := client.Hook()

	// Test Run
	ui := &testUi{}
	hClient.Run("foo", ui, nil, 42)
	if !h.RunCalled {
		t.Fatal("should be called")
	}

	// Test Cancel
	hClient.Cancel()
	if !h.CancelCalled {
		t.Fatal("should be called")
	}
}

func TestHook_Implements(t *testing.T) {
	var _ packer.Hook = new(hook)
}

func TestHook_cancelWhileRun(t *testing.T) {
	var finishLock sync.Mutex
	finishOrder := make([]string, 0, 2)

	h := &packer.MockHook{
		RunFunc: func() error {
			time.Sleep(100 * time.Millisecond)

			finishLock.Lock()
			finishOrder = append(finishOrder, "run")
			finishLock.Unlock()
			return nil
		},
	}

	// Serve
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterHook(h)
	hClient := client.Hook()

	// Start the run
	finished := make(chan struct{})
	go func() {
		hClient.Run("foo", nil, nil, nil)
		close(finished)
	}()

	// Cancel it pretty quickly.
	time.Sleep(10 * time.Millisecond)
	hClient.Cancel()

	finishLock.Lock()
	finishOrder = append(finishOrder, "cancel")
	finishLock.Unlock()

	// Verify things are good
	<-finished

	// Check the results
	expected := []string{"cancel", "run"}
	if !reflect.DeepEqual(finishOrder, expected) {
		t.Fatalf("bad: %#v", finishOrder)
	}
}
