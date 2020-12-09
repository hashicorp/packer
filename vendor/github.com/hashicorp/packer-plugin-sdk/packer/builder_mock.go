//go:generate mapstructure-to-hcl2 -type MockBuilder

package packer

import (
	"context"
	"errors"

	"github.com/hashicorp/hcl/v2/hcldec"
)

// MockBuilder is an implementation of Builder that can be used for tests.
// You can set some fake return values and you can keep track of what
// methods were called on the builder. It is fairly basic.
type MockBuilder struct {
	ArtifactId      string
	PrepareWarnings []string
	RunErrResult    bool
	RunNilResult    bool

	PrepareCalled bool
	PrepareConfig []interface{}
	RunCalled     bool
	RunHook       Hook
	RunUi         Ui
	CancelCalled  bool
	RunFn         func(ctx context.Context)

	GeneratedVars []string
}

func (tb *MockBuilder) ConfigSpec() hcldec.ObjectSpec { return tb.FlatMapstructure().HCL2Spec() }

func (tb *MockBuilder) FlatConfig() interface{} { return tb.FlatMapstructure() }

func (tb *MockBuilder) Prepare(config ...interface{}) ([]string, []string, error) {
	tb.PrepareCalled = true
	tb.PrepareConfig = config
	return tb.GeneratedVars, tb.PrepareWarnings, nil
}

func (tb *MockBuilder) Run(ctx context.Context, ui Ui, h Hook) (Artifact, error) {
	tb.RunCalled = true
	tb.RunHook = h
	tb.RunUi = ui

	if tb.RunErrResult {
		return nil, errors.New("foo")
	}

	if tb.RunNilResult {
		return nil, nil
	}
	if tb.RunFn != nil {
		tb.RunFn(ctx)
	}

	if h != nil {
		if err := h.Run(ctx, HookProvision, ui, new(MockCommunicator), nil); err != nil {
			return nil, err
		}
	}

	return &MockArtifact{
		IdValue: tb.ArtifactId,
	}, nil
}
