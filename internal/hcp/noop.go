package hcp

import (
	"context"

	sdkpacker "github.com/hashicorp/packer-plugin-sdk/packer"
)

// noopOrchestrator is a special handler that does nothing
type noopOrchestrator struct{}

func newNoopHandler() Orchestrator {
	return noopOrchestrator{}
}

func (h noopOrchestrator) PopulateIteration(context.Context) error {
	return nil
}

func (h noopOrchestrator) BuildStart(context.Context, string) error {
	return nil
}

func (h noopOrchestrator) BuildDone(
	ctx context.Context,
	buildName string,
	artifacts []sdkpacker.Artifact,
	buildErr error,
) ([]sdkpacker.Artifact, error) {
	return artifacts, nil
}
