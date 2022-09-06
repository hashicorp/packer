package hcp

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	sdkpacker "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/hashicorp/packer/packer"
)

// Orchestrator is an entity capable to orchestrate a Packer build and upload metadata to HCP
type Orchestrator interface {
	PopulateIteration(context.Context) error
	BuildStart(context.Context, string) error
	BuildDone(ctx context.Context, buildName string, artifacts []sdkpacker.Artifact, buildErr error) ([]sdkpacker.Artifact, error)
}

// GetOrchestrator instanciates the appropriate handler for the configuration type given as parameter.
//
// If no HCP-related data is present, it will be a NoopHandler.
func GetOrchestrator(cfg packer.Handler) (Orchestrator, hcl.Diagnostics) {
	var handler Orchestrator
	var err hcl.Diagnostics

	switch cfg := cfg.(type) {
	case *hcl2template.PackerConfig:
		handler, err = newHCLOrchestrator(cfg)
	case *packer.Core:
		handler, err = newJSONOrchestrator(cfg)
	}

	return handler, err
}
