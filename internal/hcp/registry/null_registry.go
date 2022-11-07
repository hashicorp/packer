package registry

import (
	"context"

	sdkpacker "github.com/hashicorp/packer-plugin-sdk/packer"
)

// nullRegistry is a special handler that does nothing
type nullRegistry struct{}

func (r nullRegistry) PopulateIteration(context.Context) error {
	return nil
}

func (r nullRegistry) StartBuild(context.Context, string) error {
	return nil
}

func (r nullRegistry) CompleteBuild(
	ctx context.Context,
	buildName string,
	artifacts []sdkpacker.Artifact,
	buildErr error,
) ([]sdkpacker.Artifact, error) {
	return artifacts, nil
}
