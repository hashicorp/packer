package packer

import (
	"context"
	"errors"
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
}

func (tb *MockBuilder) Prepare(config ...interface{}) ([]string, error) {
	tb.PrepareCalled = true
	tb.PrepareConfig = config
	return tb.PrepareWarnings, nil
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

	if h != nil {
		if err := h.Run(ctx, HookProvision, ui, new(MockCommunicator), nil); err != nil {
			return nil, err
		}
	}

	return &MockArtifact{
		IdValue: tb.ArtifactId,
	}, nil
}
