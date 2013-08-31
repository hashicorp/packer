package packer

import (
	"sync"
	"testing"
	"time"
)

func TestProvisionHook_Impl(t *testing.T) {
	var raw interface{}
	raw = &ProvisionHook{}
	if _, ok := raw.(Hook); !ok {
		t.Fatalf("must be a Hook")
	}
}

func TestProvisionHook(t *testing.T) {
	pA := &MockProvisioner{}
	pB := &MockProvisioner{}

	ui := testUi()
	var comm Communicator = nil
	var data interface{} = nil

	hook := &ProvisionHook{
		Provisioners: []Provisioner{pA, pB},
	}

	hook.Run("foo", ui, comm, data)

	if !pA.ProvCalled {
		t.Error("provision should be called on pA")
	}

	if !pB.ProvCalled {
		t.Error("provision should be called on pB")
	}
}

func TestProvisionHook_cancel(t *testing.T) {
	var lock sync.Mutex
	order := make([]string, 0, 2)

	p := &MockProvisioner{
		ProvFunc: func() error {
			time.Sleep(50 * time.Millisecond)

			lock.Lock()
			defer lock.Unlock()
			order = append(order, "prov")

			return nil
		},
	}

	hook := &ProvisionHook{
		Provisioners: []Provisioner{p},
	}

	finished := make(chan struct{})
	go func() {
		hook.Run("foo", nil, nil, nil)
		close(finished)
	}()

	// Cancel it while it is running
	time.Sleep(10 * time.Millisecond)
	hook.Cancel()
	lock.Lock()
	order = append(order, "cancel")
	lock.Unlock()

	// Wait
	<-finished

	// Verify order
	if order[0] != "cancel" || order[1] != "prov" {
		t.Fatalf("bad: %#v", order)
	}
}

// TODO(mitchellh): Test that they're run in the proper order
