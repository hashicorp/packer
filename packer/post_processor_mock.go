package packer

import (
	"context"

	"github.com/hashicorp/hcl/v2/hcldec"
)

// MockPostProcessor is an implementation of PostProcessor that can be
// used for tests.
type MockPostProcessor struct {
	ArtifactId    string
	Keep          bool
	ForceOverride bool
	Error         error

	ConfigureCalled  bool
	ConfigureConfigs []interface{}
	ConfigureError   error

	PostProcessCalled   bool
	PostProcessArtifact Artifact
	PostProcessUi       Ui
}

func (t *MockPostProcessor) ConfigSpec() hcldec.ObjectSpec { return t.FlatMapstructure().HCL2Spec() }

func (t *MockPostProcessor) Configure(configs ...interface{}) error {
	t.ConfigureCalled = true
	t.ConfigureConfigs = configs
	return t.ConfigureError
}

func (t *MockPostProcessor) PostProcess(ctx context.Context, ui Ui, a Artifact) (Artifact, bool, bool, error) {
	t.PostProcessCalled = true
	t.PostProcessArtifact = a
	t.PostProcessUi = ui

	return &MockArtifact{
		IdValue: t.ArtifactId,
	}, t.Keep, t.ForceOverride, t.Error
}
