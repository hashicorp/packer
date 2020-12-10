package plugin

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/version"
)

type MockBuilder struct {
	packersdk.Builder
}

var _ packersdk.Builder = new(MockBuilder)

type MockProvisioner struct {
	packersdk.Provisioner
}

var _ packersdk.Provisioner = new(MockProvisioner)

type MockPostProcessor struct {
	packersdk.PostProcessor
}

var _ packersdk.PostProcessor = new(MockPostProcessor)

func TestSet(t *testing.T) {
	set := NewSet()
	set.RegisterBuilder("example-2", new(MockBuilder))
	set.RegisterBuilder("example", new(MockBuilder))
	set.RegisterPostProcessor("example", new(MockPostProcessor))
	set.RegisterPostProcessor("example-2", new(MockPostProcessor))
	set.RegisterProvisioner("example", new(MockProvisioner))
	set.RegisterProvisioner("example-2", new(MockProvisioner))

	outputDesc := set.description()

	if diff := cmp.Diff(SetDescription{
		Version:        version.String(),
		SDKVersion:     version.String(),
		Builders:       []string{"example", "example-2"},
		PostProcessors: []string{"example", "example-2"},
		Provisioners:   []string{"example", "example-2"},
	}, outputDesc); diff != "" {
		t.Fatalf("Unexpected description: %s", diff)
	}

	err := set.run("start", "builder", "example")
	if diff := cmp.Diff(err.Error(), ErrManuallyStartedPlugin.Error()); diff != "" {
		t.Fatalf("Unexpected error: %s", diff)
	}
}
