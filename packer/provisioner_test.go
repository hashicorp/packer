package packer

import (
	"context"
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
	var comm Communicator = new(MockCommunicator)
	var data interface{} = nil

	hook := &ProvisionHook{
		Provisioners: []*HookedProvisioner{
			{pA, nil, ""},
			{pB, nil, ""},
		},
	}

	hook.Run(context.Background(), "foo", ui, comm, data)

	if !pA.ProvCalled {
		t.Error("provision should be called on pA")
	}

	if !pB.ProvCalled {
		t.Error("provision should be called on pB")
	}
}

func TestProvisionHook_nilComm(t *testing.T) {
	pA := &MockProvisioner{}
	pB := &MockProvisioner{}

	ui := testUi()
	var comm Communicator = nil
	var data interface{} = nil

	hook := &ProvisionHook{
		Provisioners: []*HookedProvisioner{
			{pA, nil, ""},
			{pB, nil, ""},
		},
	}

	err := hook.Run(context.Background(), "foo", ui, comm, data)
	if err == nil {
		t.Fatal("should error")
	}
}

func TestProvisionHook_cancel(t *testing.T) {
	var lock sync.Mutex
	order := make([]string, 0, 2)

	p := &MockProvisioner{
		ProvFunc: func() error {
			time.Sleep(100 * time.Millisecond)

			lock.Lock()
			defer lock.Unlock()
			order = append(order, "prov")

			return nil
		},
	}

	hook := &ProvisionHook{
		Provisioners: []*HookedProvisioner{
			{p, nil, ""},
		},
	}
	ctx, cancel := context.WithCancel(context.Background())

	finished := make(chan struct{})
	go func() {
		hook.Run(ctx, "foo", nil, new(MockCommunicator), nil)
		close(finished)
	}()

	// Cancel it while it is running
	time.Sleep(10 * time.Millisecond)
	cancel()
	lock.Lock()
	order = append(order, "cancel")
	lock.Unlock()

	// Wait
	<-finished

	// Verify order
	if len(order) != 2 || order[0] != "cancel" || order[1] != "prov" {
		t.Fatalf("bad: %#v", order)
	}
}

// TODO(mitchellh): Test that they're run in the proper order

func TestPausedProvisioner_impl(t *testing.T) {
	var _ Provisioner = new(PausedProvisioner)
}

func TestPausedProvisionerPrepare(t *testing.T) {
	mock := new(MockProvisioner)
	prov := &PausedProvisioner{
		Provisioner: mock,
	}

	prov.Prepare(42)
	if !mock.PrepCalled {
		t.Fatal("prepare should be called")
	}
	if mock.PrepConfigs[0] != 42 {
		t.Fatal("should have proper configs")
	}
}

func TestPausedProvisionerProvision(t *testing.T) {
	mock := new(MockProvisioner)
	prov := &PausedProvisioner{
		Provisioner: mock,
	}

	ui := testUi()
	comm := new(MockCommunicator)
	prov.Provision(context.Background(), ui, comm)
	if !mock.ProvCalled {
		t.Fatal("prov should be called")
	}
	if mock.ProvUi != ui {
		t.Fatal("should have proper ui")
	}
	if mock.ProvCommunicator != comm {
		t.Fatal("should have proper comm")
	}
}

func TestPausedProvisionerProvision_waits(t *testing.T) {
	mock := new(MockProvisioner)
	prov := &PausedProvisioner{
		PauseBefore: 50 * time.Millisecond,
		Provisioner: mock,
	}

	dataCh := make(chan struct{})
	mock.ProvFunc = func() error {
		close(dataCh)
		return nil
	}

	go prov.Provision(context.Background(), testUi(), new(MockCommunicator))

	select {
	case <-time.After(10 * time.Millisecond):
	case <-dataCh:
		t.Fatal("should not be called")
	}

	select {
	case <-time.After(100 * time.Millisecond):
		t.Fatal("never called")
	case <-dataCh:
	}
}

func TestPausedProvisionerCancel(t *testing.T) {
	mock := new(MockProvisioner)
	prov := &PausedProvisioner{
		Provisioner: mock,
	}

	provCh := make(chan struct{})
	mock.ProvFunc = func() error {
		close(provCh)
		time.Sleep(10 * time.Millisecond)
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())

	// Start provisioning and wait for it to start
	go func() {
		<-provCh
		cancel()
	}()

	prov.Provision(ctx, testUi(), new(MockCommunicator))
	if !mock.CancelCalled {
		t.Fatal("cancel should be called")
	}
}

func TestDebuggedProvisioner_impl(t *testing.T) {
	var _ Provisioner = new(DebuggedProvisioner)
}

func TestDebuggedProvisionerPrepare(t *testing.T) {
	mock := new(MockProvisioner)
	prov := &DebuggedProvisioner{
		Provisioner: mock,
	}

	prov.Prepare(42)
	if !mock.PrepCalled {
		t.Fatal("prepare should be called")
	}
	if mock.PrepConfigs[0] != 42 {
		t.Fatal("should have proper configs")
	}
}

func TestDebuggedProvisionerProvision(t *testing.T) {
	mock := new(MockProvisioner)
	prov := &DebuggedProvisioner{
		Provisioner: mock,
	}

	ui := testUi()
	comm := new(MockCommunicator)
	writeReader(ui, "\n")
	prov.Provision(context.Background(), ui, comm)
	if !mock.ProvCalled {
		t.Fatal("prov should be called")
	}
	if mock.ProvUi != ui {
		t.Fatal("should have proper ui")
	}
	if mock.ProvCommunicator != comm {
		t.Fatal("should have proper comm")
	}
}

func TestDebuggedProvisionerCancel(t *testing.T) {
	mock := new(MockProvisioner)
	prov := &DebuggedProvisioner{
		Provisioner: mock,
	}

	provCh := make(chan struct{})
	mock.ProvFunc = func() error {
		close(provCh)
		time.Sleep(10 * time.Millisecond)
		return nil
	}
	ctx, cancel := context.WithCancel(context.Background())

	// Start provisioning and wait for it to start
	go func() {
		<-provCh
		cancel()
	}()

	prov.Provision(ctx, testUi(), new(MockCommunicator))
	if !mock.CancelCalled {
		t.Fatal("cancel should be called")
	}
}
