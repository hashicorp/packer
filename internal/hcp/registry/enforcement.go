// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

package registry

import (
	"fmt"
	"sort"

	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
)

// Normalized bucket enforcement policy modes (RFC 6.1).
const (
	policyModeMandatory = "mandatory"
	policyModeAdvisory  = "advisory"
)

// RFC section 11 hard limits, mirrored client-side as a defensive guardrail
// against resolver contract drift. These MUST stay in sync with the
// cloud-packer-service handlers (maxBlockContentBytes / maxLinkedProvisionersPerBucket).
const (
	// maxBlockContentBytes caps a single enforced provisioner's block_content.
	maxBlockContentBytes = 128 * 1024 // 128 KiB
	// maxLinkedProvisionersPerBucket caps how many enforced provisioners a
	// bucket may resolve to.
	maxLinkedProvisionersPerBucket = 25
)

// buildEnforcedBlocks maps the resolver's effective-provisioner entries to the
// CLI's EnforcedBlock type and returns them in canonical (ascending ordinal)
// execution order. Nil entries are skipped; the sort is defensive against
// transport reordering (RFC 6.2).
func buildEnforcedBlocks(eps []*hcpPackerModels.HashicorpCloudPacker20230101EffectiveProvisioner) []*EnforcedBlock {
	blocks := make([]*EnforcedBlock, 0, len(eps))
	for _, ep := range eps {
		if ep == nil {
			continue
		}
		block := &EnforcedBlock{
			ID:           ep.EnforcedBlockID,
			BlockContent: ep.BlockContent,
			VersionID:    ep.EnforcedBlockVersionID,
			Ordinal:      ep.Ordinal,
			ContentHash:  ep.ContentHash,
			// Name is not part of the effective entry; fall back to the version
			// id for log/UI identification.
			Name: ep.EnforcedBlockID,
		}
		if ep.TemplateType != nil {
			block.TemplateType = string(*ep.TemplateType)
		}
		if ep.VersionState != nil {
			block.VersionState = string(*ep.VersionState)
		}
		blocks = append(blocks, block)
	}
	sort.SliceStable(blocks, func(i, j int) bool { return blocks[i].Ordinal < blocks[j].Ordinal })
	return blocks
}

// validateResolvedEnforcementLimits enforces the RFC section 11 hard limits on
// the resolved enforced-provisioner set as a client-side guardrail. It returns
// a non-nil error describing the first violation; callers decide whether to
// fail closed (mandatory) or warn and continue (advisory).
func validateResolvedEnforcementLimits(blocks []*EnforcedBlock) error {
	if len(blocks) > maxLinkedProvisionersPerBucket {
		return fmt.Errorf(
			"resolver returned %d enforced provisioners, exceeding the maximum of %d per bucket",
			len(blocks), maxLinkedProvisionersPerBucket,
		)
	}
	for _, b := range blocks {
		if b == nil {
			continue
		}
		if len(b.BlockContent) > maxBlockContentBytes {
			id := b.VersionID
			if id == "" {
				id = b.ID
			}
			return fmt.Errorf(
				"enforced provisioner %q block_content is %d bytes, exceeding the maximum of %d bytes",
				id, len(b.BlockContent), maxBlockContentBytes,
			)
		}
	}
	return nil
}

// Closed set of skip reason codes for GA (RFC 10). Additions require an RFC
// amendment.
const (
	SkipReasonBreakglassIncident = "breakglass_incident"
	SkipReasonResolverOutage     = "resolver_outage"
	SkipReasonVerifiedException  = "verified_exception"
	SkipReasonMigrationCompat    = "migration_compatibility"
)

// ValidSkipReasonCodes is the closed enum of accepted --skip-reason-code values.
var ValidSkipReasonCodes = []string{
	SkipReasonBreakglassIncident,
	SkipReasonResolverOutage,
	SkipReasonVerifiedException,
	SkipReasonMigrationCompat,
}

// IsValidSkipReasonCode reports whether code is a member of the closed reason
// enum (RFC 10).
func IsValidSkipReasonCode(code string) bool {
	for _, c := range ValidSkipReasonCodes {
		if c == code {
			return true
		}
	}
	return false
}

// EnforcementOptions carries CLI-supplied context into the resolver call.
type EnforcementOptions struct {
	// CLIVersion is the calling Packer version, used by the server for
	// minimum-version enforcement on mandatory buckets (RFC 12.4).
	CLIVersion string
	// BuildCorrelationID correlates the resolution with build audit events.
	BuildCorrelationID string
}

// normalizePolicyMode maps the SDK enum (or its short form) to the normalized
// "mandatory"/"advisory" vocabulary. Unset defaults to mandatory (RFC 6.1).
func normalizePolicyMode(mode *hcpPackerModels.HashicorpCloudPacker20230101EnforcementPolicyMode) string {
	if mode == nil {
		return policyModeMandatory
	}
	switch *mode {
	case hcpPackerModels.HashicorpCloudPacker20230101EnforcementPolicyModeENFORCEMENTPOLICYMODEADVISORY:
		return policyModeAdvisory
	case hcpPackerModels.HashicorpCloudPacker20230101EnforcementPolicyModeENFORCEMENTPOLICYMODEMANDATORY:
		return policyModeMandatory
	default:
		// UNSET or unknown: fail-safe to the GA default.
		return policyModeMandatory
	}
}

// versionStateReleased is the expected lifecycle state of a resolved version.
const versionStateReleased = "ENFORCED_BLOCK_VERSION_STATUS_RELEASED"
