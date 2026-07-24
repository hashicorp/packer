// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package packer

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestProvisionHook_Impl(t *testing.T) {
	var raw any = &ProvisionHook{}
	if _, ok := raw.(packersdk.Hook); !ok {
		t.Fatalf("must be a Hook")
	}
}

func TestProvisionHook(t *testing.T) {
	pA := &packersdk.MockProvisioner{}
	pB := &packersdk.MockProvisioner{}

	ui := testUi()
	var comm packersdk.Communicator = new(packersdk.MockCommunicator)
	var data any = nil

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
	pA := &packersdk.MockProvisioner{}
	pB := &packersdk.MockProvisioner{}

	ui := testUi()
	var comm packersdk.Communicator = nil
	var data any = nil

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
	topCtx, topCtxCancel := context.WithCancel(context.Background())

	p := &packersdk.MockProvisioner{
		ProvFunc: func(ctx context.Context) error {
			topCtxCancel()
			<-ctx.Done()
			return ctx.Err()
		},
	}

	hook := &ProvisionHook{
		Provisioners: []*HookedProvisioner{
			{p, nil, ""},
		},
	}

	err := hook.Run(topCtx, "foo", nil, new(packersdk.MockCommunicator), nil)
	if err == nil {
		t.Fatal("should have err")
	}
}

// TODO(mitchellh): Test that they're run in the proper order

func TestPausedProvisioner_impl(t *testing.T) {
	var _ packersdk.Provisioner = new(PausedProvisioner)
}

func TestPausedProvisionerPrepare(t *testing.T) {
	mock := new(packersdk.MockProvisioner)
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
	mock := new(packersdk.MockProvisioner)
	prov := &PausedProvisioner{
		Provisioner: mock,
	}

	ui := testUi()
	comm := new(packersdk.MockCommunicator)
	if err := prov.Provision(context.Background(), ui, comm, make(map[string]any)); err != nil {
		t.Fatalf("provision failed: %v", err)
	}
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
	startTime := time.Now()
	waitTime := 50 * time.Millisecond

	prov := &PausedProvisioner{
		PauseBefore: waitTime,
		Provisioner: &packersdk.MockProvisioner{
			ProvFunc: func(context.Context) error {
				timeSinceStartTime := time.Since(startTime)
				if timeSinceStartTime < waitTime {
					return fmt.Errorf("Spent not enough time waiting: %s", timeSinceStartTime)
				}
				return nil
			},
		},
	}

	err := prov.Provision(context.Background(), testUi(), new(packersdk.MockCommunicator), make(map[string]any))

	if err != nil {
		t.Fatalf("prov failed: %v", err)
	}
}

func TestPausedProvisionerCancel(t *testing.T) {
	topCtx, cancelTopCtx := context.WithCancel(context.Background())

	mock := new(packersdk.MockProvisioner)
	prov := &PausedProvisioner{
		Provisioner: mock,
	}

	mock.ProvFunc = func(ctx context.Context) error {
		cancelTopCtx()
		<-ctx.Done()
		return ctx.Err()
	}

	err := prov.Provision(topCtx, testUi(), new(packersdk.MockCommunicator), make(map[string]any))
	if err == nil {
		t.Fatal("should have err")
	}
}

func TestDebuggedProvisioner_impl(t *testing.T) {
	var _ packersdk.Provisioner = new(DebuggedProvisioner)
}

func TestDebuggedProvisionerPrepare(t *testing.T) {
	mock := new(packersdk.MockProvisioner)
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
	mock := new(packersdk.MockProvisioner)
	prov := &DebuggedProvisioner{
		Provisioner: mock,
	}

	ui := testUi()
	comm := new(packersdk.MockCommunicator)
	writeReader(ui, "\n")
	if err := prov.Provision(context.Background(), ui, comm, make(map[string]any)); err != nil {
		t.Fatalf("provision failed: %v", err)
	}
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
	topCtx, topCtxCancel := context.WithCancel(context.Background())

	mock := new(packersdk.MockProvisioner)
	prov := &DebuggedProvisioner{
		Provisioner: mock,
	}

	mock.ProvFunc = func(ctx context.Context) error {
		topCtxCancel()
		<-ctx.Done()
		return ctx.Err()
	}

	err := prov.Provision(topCtx, testUi(), new(packersdk.MockCommunicator), make(map[string]any))
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestRetriedProvisioner_impl(t *testing.T) {
	var _ packersdk.Provisioner = new(RetriedProvisioner)
}

func TestRetriedProvisionerPrepare(t *testing.T) {
	mock := new(packersdk.MockProvisioner)
	prov := &RetriedProvisioner{
		Provisioner: mock,
	}

	err := prov.Prepare(42)
	if err != nil {
		t.Fatal("should not have errored")
	}
	if !mock.PrepCalled {
		t.Fatal("prepare should be called")
	}
	if mock.PrepConfigs[0] != 42 {
		t.Fatal("should have proper configs")
	}
}

func TestRetriedProvisionerProvision(t *testing.T) {
	mock := &packersdk.MockProvisioner{
		ProvFunc: func(ctx context.Context) error {
			return errors.New("failed")
		},
	}

	prov := &RetriedProvisioner{
		MaxRetries:  2,
		Provisioner: mock,
	}

	ui := testUi()
	comm := new(packersdk.MockCommunicator)
	err := prov.Provision(context.Background(), ui, comm, make(map[string]any))
	if err != nil {
		t.Fatal("should not have errored")
	}
	if !mock.ProvCalled {
		t.Fatal("prov should be called")
	}
	if !mock.ProvRetried {
		t.Fatal("prov should be retried")
	}
	if mock.ProvUi != ui {
		t.Fatal("should have proper ui")
	}
	if mock.ProvCommunicator != comm {
		t.Fatal("should have proper comm")
	}
}

func TestRetriedProvisionerCancelledProvision(t *testing.T) {
	// Don't retry if context is cancelled
	ctx, topCtxCancel := context.WithCancel(context.Background())

	mock := &packersdk.MockProvisioner{
		ProvFunc: func(ctx context.Context) error {
			topCtxCancel()
			<-ctx.Done()
			return ctx.Err()
		},
	}

	prov := &RetriedProvisioner{
		MaxRetries:  2,
		Provisioner: mock,
	}

	ui := testUi()
	comm := new(packersdk.MockCommunicator)
	err := prov.Provision(ctx, ui, comm, make(map[string]any))
	if err == nil {
		t.Fatal("should have errored")
	}
	if !mock.ProvCalled {
		t.Fatal("prov should be called")
	}
	if mock.ProvRetried {
		t.Fatal("prov should NOT be retried")
	}
	if mock.ProvUi != ui {
		t.Fatal("should have proper ui")
	}
	if mock.ProvCommunicator != comm {
		t.Fatal("should have proper comm")
	}
}

func TestRetriedProvisionerCancel(t *testing.T) {
	topCtx, cancelTopCtx := context.WithCancel(context.Background())

	mock := new(packersdk.MockProvisioner)
	prov := &RetriedProvisioner{
		Provisioner: mock,
	}

	mock.ProvFunc = func(ctx context.Context) error {
		cancelTopCtx()
		<-ctx.Done()
		return ctx.Err()
	}

	err := prov.Provision(topCtx, testUi(), new(packersdk.MockCommunicator), make(map[string]any))
	if err == nil {
		t.Fatal("should have err")
	}
}

func TestContinueOnErrorProvisioner_impl(t *testing.T) {
	var _ packersdk.Provisioner = new(ContinueOnErrorProvisioner)
}

func TestContinueOnErrorProvisionerConfigSpecAndFlatConfig_doNotRecurse(t *testing.T) {
	mock := new(packersdk.MockProvisioner)
	prov := &ContinueOnErrorProvisioner{Provisioner: mock}
	_ = prov.ConfigSpec()
	_ = prov.FlatConfig()
}

func TestContinueOnErrorProvisionerPrepare(t *testing.T) {
	mock := new(packersdk.MockProvisioner)
	prov := &ContinueOnErrorProvisioner{
		Provisioner: mock,
	}

	err := prov.Prepare(42)
	if err != nil {
		t.Fatal("should not have errored")
	}
	if !mock.PrepCalled {
		t.Fatal("prepare should be called")
	}
	if mock.PrepConfigs[0] != 42 {
		t.Fatal("should have proper configs")
	}
}

func TestContinueOnErrorProvisionerProvision(t *testing.T) {
	// A failing provisioner must not propagate its error.
	mock := &packersdk.MockProvisioner{
		ProvFunc: func(ctx context.Context) error {
			return errors.New("failed")
		},
	}

	prov := &ContinueOnErrorProvisioner{
		Provisioner: mock,
	}

	ui := testUi()
	comm := new(packersdk.MockCommunicator)
	err := prov.Provision(context.Background(), ui, comm, make(map[string]any))
	if err != nil {
		t.Fatalf("should have swallowed the error, got: %s", err)
	}
	if !mock.ProvCalled {
		t.Fatal("prov should be called")
	}
}

func TestContinueOnErrorProvisionerProvision_success(t *testing.T) {
	// A successful provisioner returns nil as usual.
	mock := new(packersdk.MockProvisioner)

	prov := &ContinueOnErrorProvisioner{
		Provisioner: mock,
	}

	err := prov.Provision(context.Background(), testUi(), new(packersdk.MockCommunicator), make(map[string]any))
	if err != nil {
		t.Fatalf("should not have errored, got: %s", err)
	}
	if !mock.ProvCalled {
		t.Fatal("prov should be called")
	}
}

func TestContinueOnErrorProvisionerCancelledProvision(t *testing.T) {
	// A cancelled context must still propagate, even with continue_on_error.
	ctx, cancel := context.WithCancel(context.Background())

	mock := &packersdk.MockProvisioner{
		ProvFunc: func(ctx context.Context) error {
			cancel()
			<-ctx.Done()
			return ctx.Err()
		},
	}

	prov := &ContinueOnErrorProvisioner{
		Provisioner: mock,
	}

	err := prov.Provision(ctx, testUi(), new(packersdk.MockCommunicator), make(map[string]any))
	if err == nil {
		t.Fatal("should have propagated the cancellation error")
	}
}

// TestProvisionHook_failsWithoutContinueOnError is the negative case: when a
// provisioner fails and is NOT wrapped with continue_on_error (i.e. the option
// is false or unset), the hook must return the error and must NOT run any
// subsequent provisioners.
func TestProvisionHook_failsWithoutContinueOnError(t *testing.T) {
	pA := &packersdk.MockProvisioner{
		ProvFunc: func(ctx context.Context) error {
			return errors.New("failed")
		},
	}
	pB := &packersdk.MockProvisioner{}

	hook := &ProvisionHook{
		Provisioners: []*HookedProvisioner{
			{pA, nil, ""},
			{pB, nil, ""},
		},
	}

	err := hook.Run(context.Background(), "foo", testUi(), new(packersdk.MockCommunicator), nil)
	if err == nil {
		t.Fatal("hook should have errored when a provisioner fails")
	}
	if !pA.ProvCalled {
		t.Fatal("failing provisioner should have been called")
	}
	if pB.ProvCalled {
		t.Fatal("subsequent provisioner should NOT run after a failure")
	}
}

// TestProvisionHook_continueOnErrorRunsRemaining is the positive case: when the
// failing provisioner is wrapped with continue_on_error, the hook swallows the
// error and continues running the remaining provisioners.
func TestProvisionHook_continueOnErrorRunsRemaining(t *testing.T) {
	failing := &packersdk.MockProvisioner{
		ProvFunc: func(ctx context.Context) error {
			return errors.New("failed")
		},
	}
	pB := &packersdk.MockProvisioner{}

	hook := &ProvisionHook{
		Provisioners: []*HookedProvisioner{
			{&ContinueOnErrorProvisioner{Provisioner: failing}, nil, ""},
			{pB, nil, ""},
		},
	}

	err := hook.Run(context.Background(), "foo", testUi(), new(packersdk.MockCommunicator), nil)
	if err != nil {
		t.Fatalf("hook should not have errored, got: %s", err)
	}
	if !failing.ProvCalled {
		t.Fatal("failing provisioner should have been called")
	}
	if !pB.ProvCalled {
		t.Fatal("subsequent provisioner should run after a swallowed failure")
	}
}
