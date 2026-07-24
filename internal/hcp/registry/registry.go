// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

// Package registry provides access to the HCP registry.
package registry

import (
	"context"

	"github.com/hashicorp/hcl/v2"
	sdkpacker "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/hashicorp/packer/packer"
)

// Registry is an entity capable to orchestrate a Packer build and upload metadata to HCP
type Registry interface {
	PopulateVersion(context.Context) error
	StartBuild(context.Context, *packer.CoreBuild) error
	CompleteBuild(ctx context.Context, build *packer.CoreBuild, artifacts []sdkpacker.Artifact, buildErr error) ([]sdkpacker.Artifact, error)
	VersionStatusSummary()
	Metadata() Metadata
	// FetchEnforcedBlocks resolves the effective enforced-provisioner set from
	// HCP Packer (RFC 6.2) and applies the mandatory/advisory failure matrix.
	FetchEnforcedBlocks(ctx context.Context, opts EnforcementOptions) error
	// InjectEnforcedProvisioners injects the resolved enforced provisioners into
	// the builds in canonical execution order.
	InjectEnforcedProvisioners(builds []*packer.CoreBuild) hcl.Diagnostics
	// RecordEnforcementSkip records an authorized --skip-enforcement decision
	// into build metadata (RFC 10).
	RecordEnforcementSkip(reasonCode, reasonNote string)
}

// New instantiates the appropriate registry for the Packer configuration template type.
// A nullRegistry is returned for non-HCP Packer registry enabled templates.
func New(cfg packer.Handler, ui sdkpacker.Ui) (Registry, hcl.Diagnostics) {
	if !IsHCPEnabled(cfg) {
		return &nullRegistry{}, nil
	}

	switch config := cfg.(type) {
	case *hcl2template.PackerConfig:
		// Maybe rename to what it represents....
		return NewHCLRegistry(config, ui)
	case *packer.Core:
		return NewJSONRegistry(config, ui)
	}

	return nil, hcl.Diagnostics{
		&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unknown Config type",
			Detail: "The config type %s does not match a Packer-known template type. " +
				"This is a Packer error and should be brought up to the Packer " +
				"team via a GitHub Issue.",
		},
	}
}
