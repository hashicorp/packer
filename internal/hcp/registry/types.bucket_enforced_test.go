// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

package registry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	hcpPackerModels "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/stable/2023-01-01/models"
	hcpPackerAPI "github.com/hashicorp/packer/internal/hcp/api"
	"google.golang.org/grpc/codes"
)

func newTestBucket(mock *hcpPackerAPI.MockPackerClientService) *Bucket {
	return &Bucket{
		Name:        "test-bucket",
		BuildLabels: map[string]string{},
		client:      &hcpPackerAPI.Client{Packer: mock, OrganizationID: "org-1", ProjectID: "prj-1"},
	}
}

func mandatory() *hcpPackerModels.HashicorpCloudPacker20230101EnforcementPolicyMode {
	return hcpPackerModels.NewHashicorpCloudPacker20230101EnforcementPolicyMode(
		hcpPackerModels.HashicorpCloudPacker20230101EnforcementPolicyModeENFORCEMENTPOLICYMODEMANDATORY)
}

func advisory() *hcpPackerModels.HashicorpCloudPacker20230101EnforcementPolicyMode {
	return hcpPackerModels.NewHashicorpCloudPacker20230101EnforcementPolicyMode(
		hcpPackerModels.HashicorpCloudPacker20230101EnforcementPolicyModeENFORCEMENTPOLICYMODEADVISORY)
}

func TestBucket_FetchEnforcedBlocks_ResolverReturnsOrderedSet(t *testing.T) {
	t.Setenv("PACKER_ENFORCEMENT_CACHE_DIR", t.TempDir())
	hcl2Type := hcpPackerModels.HashicorpCloudPacker20230101TemplateTypeHCL2

	mockService := hcpPackerAPI.NewMockPackerClientService()
	// Provide entries out of ordinal order to verify the CLI applies the
	// canonical (ascending ordinal) execution order.
	mockService.ResolveEnforcedProvisionersResp = &hcpPackerModels.HashicorpCloudPacker20230101ResolveEnforcedProvisionersResponse{
		PolicyMode:           mandatory(),
		ResolutionTTLSeconds: 300,
		EffectiveProvisioners: []*hcpPackerModels.HashicorpCloudPacker20230101EffectiveProvisioner{
			{
				EnforcedBlockID:        "eb-20",
				EnforcedBlockVersionID: "ebv-20",
				BlockContent:           "provisioner \"shell\" {}",
				ContentHash:            "sha256:bbb",
				Ordinal:                20,
				TemplateType:           &hcl2Type,
			},
			{
				EnforcedBlockID:        "eb-10",
				EnforcedBlockVersionID: "ebv-10",
				BlockContent:           "provisioner \"shell\" {}",
				ContentHash:            "sha256:aaa",
				Ordinal:                10,
				TemplateType:           &hcl2Type,
			},
		},
		AuditContext: &hcpPackerModels.HashicorpCloudPacker20230101ResolveAuditContext{
			ResolutionID: "res-1",
			Etag:         "W/\"etag-1\"",
		},
	}

	bucket := newTestBucket(mockService)

	if err := bucket.FetchEnforcedBlocks(context.Background(), EnforcementOptions{CLIVersion: "1.2.3"}); err != nil {
		t.Fatalf("FetchEnforcedBlocks() unexpected error: %v", err)
	}

	if got := len(bucket.EnforcedBlocks); got != 2 {
		t.Fatalf("got %d blocks, want 2", got)
	}
	if bucket.EnforcedBlocks[0].Ordinal != 10 || bucket.EnforcedBlocks[1].Ordinal != 20 {
		t.Fatalf("blocks not in ascending ordinal order: got %d then %d",
			bucket.EnforcedBlocks[0].Ordinal, bucket.EnforcedBlocks[1].Ordinal)
	}
	if bucket.EnforcedBlocks[0].ContentHash != "sha256:aaa" {
		t.Fatalf("first block content hash = %q, want sha256:aaa", bucket.EnforcedBlocks[0].ContentHash)
	}
	if bucket.EnforcementPolicyMode != policyModeMandatory {
		t.Fatalf("policy mode = %q, want mandatory", bucket.EnforcementPolicyMode)
	}
	if bucket.EnforcementResolutionID != "res-1" {
		t.Fatalf("resolution id = %q, want res-1", bucket.EnforcementResolutionID)
	}
	// Build metadata must capture the resolved context (RFC 6.3).
	if bucket.BuildLabels[enforcementLabelResolutionID] != "res-1" {
		t.Fatalf("metadata resolution id label = %q, want res-1", bucket.BuildLabels[enforcementLabelResolutionID])
	}
	if bucket.BuildLabels[enforcementLabelContentHashes] != "sha256:aaa,sha256:bbb" {
		t.Fatalf("metadata content hashes label = %q", bucket.BuildLabels[enforcementLabelContentHashes])
	}
}

// makeEffectiveProvisioners builds n resolver entries with ascending ordinals.
// Each block_content is contentBytes long (use it to exercise the payload cap).
func makeEffectiveProvisioners(n, contentBytes int) []*hcpPackerModels.HashicorpCloudPacker20230101EffectiveProvisioner {
	hcl2Type := hcpPackerModels.HashicorpCloudPacker20230101TemplateTypeHCL2
	out := make([]*hcpPackerModels.HashicorpCloudPacker20230101EffectiveProvisioner, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, &hcpPackerModels.HashicorpCloudPacker20230101EffectiveProvisioner{
			EnforcedBlockID:        fmt.Sprintf("eb-%d", i),
			EnforcedBlockVersionID: fmt.Sprintf("ebv-%d", i),
			BlockContent:           strings.Repeat("a", contentBytes),
			ContentHash:            fmt.Sprintf("sha256:%d", i),
			Ordinal:                int32((i + 1) * 10),
			TemplateType:           &hcl2Type,
		})
	}
	return out
}

func TestBucket_FetchEnforcedBlocks_HardLimits(t *testing.T) {
	tests := []struct {
		name           string
		mode           *hcpPackerModels.HashicorpCloudPacker20230101EnforcementPolicyMode
		count          int
		contentBytes   int
		wantErr        bool
		wantConfigured bool
		wantWarnings   bool
	}{
		{
			name:           "at count limit (25) is accepted",
			mode:           mandatory(),
			count:          maxLinkedProvisionersPerBucket,
			contentBytes:   16,
			wantErr:        false,
			wantConfigured: true,
		},
		{
			name:           "at payload limit (128 KiB) is accepted",
			mode:           mandatory(),
			count:          1,
			contentBytes:   maxBlockContentBytes,
			wantErr:        false,
			wantConfigured: true,
		},
		{
			name:         "over count limit fails closed (mandatory)",
			mode:         mandatory(),
			count:        maxLinkedProvisionersPerBucket + 1,
			contentBytes: 16,
			wantErr:      true,
		},
		{
			name:         "over payload limit fails closed (mandatory)",
			mode:         mandatory(),
			count:        1,
			contentBytes: maxBlockContentBytes + 1,
			wantErr:      true,
		},
		{
			name:           "over count limit warns and drops (advisory)",
			mode:           advisory(),
			count:          maxLinkedProvisionersPerBucket + 1,
			contentBytes:   16,
			wantErr:        false,
			wantConfigured: false,
			wantWarnings:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PACKER_ENFORCEMENT_CACHE_DIR", t.TempDir())
			mockService := hcpPackerAPI.NewMockPackerClientService()
			mockService.ResolveEnforcedProvisionersResp = &hcpPackerModels.HashicorpCloudPacker20230101ResolveEnforcedProvisionersResponse{
				PolicyMode:            tt.mode,
				ResolutionTTLSeconds:  300,
				EffectiveProvisioners: makeEffectiveProvisioners(tt.count, tt.contentBytes),
				AuditContext: &hcpPackerModels.HashicorpCloudPacker20230101ResolveAuditContext{
					ResolutionID: "res-1",
					Etag:         "W/\"etag-1\"",
				},
			}

			bucket := newTestBucket(mockService)
			err := bucket.FetchEnforcedBlocks(context.Background(), EnforcementOptions{CLIVersion: "1.2.3"})

			if (err != nil) != tt.wantErr {
				t.Fatalf("FetchEnforcedBlocks() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if bucket.EnforcementConfigured != tt.wantConfigured {
				t.Fatalf("EnforcementConfigured = %v, want %v", bucket.EnforcementConfigured, tt.wantConfigured)
			}
			if tt.wantWarnings && len(bucket.EnforcementWarnings) == 0 {
				t.Fatal("expected an advisory degradation warning, got none")
			}
		})
	}
}

func TestBucket_FetchEnforcedBlocks_MandatoryFailClosedOnError(t *testing.T) {
	t.Setenv("PACKER_ENFORCEMENT_CACHE_DIR", t.TempDir())
	mockService := hcpPackerAPI.NewMockPackerClientService()
	mockService.ResolveEnforcedProvisionersErr = errors.New("service unavailable")

	bucket := newTestBucket(mockService)

	err := bucket.FetchEnforcedBlocks(context.Background(), EnforcementOptions{})
	if err == nil {
		t.Fatal("FetchEnforcedBlocks() expected fail-closed error in mandatory mode, got nil")
	}
}

func TestBucket_FetchEnforcedBlocks_NotFoundIsNonFatal(t *testing.T) {
	t.Setenv("PACKER_ENFORCEMENT_CACHE_DIR", t.TempDir())
	mockService := hcpPackerAPI.NewMockPackerClientService()
	mockService.ResolveEnforcedProvisionersErr = fmt.Errorf("Code:%d %s", codes.NotFound, codes.NotFound.String())

	bucket := newTestBucket(mockService)

	if err := bucket.FetchEnforcedBlocks(context.Background(), EnforcementOptions{}); err != nil {
		t.Fatalf("FetchEnforcedBlocks() expected nil error for NotFound, got: %v", err)
	}
	if bucket.EnforcementConfigured {
		t.Fatal("expected EnforcementConfigured=false when resolver unsupported")
	}
}

func TestBucket_FetchEnforcedBlocks_ClientUpgradeRequired(t *testing.T) {
	t.Setenv("PACKER_ENFORCEMENT_CACHE_DIR", t.TempDir())
	mockService := hcpPackerAPI.NewMockPackerClientService()
	mockService.ResolveEnforcedProvisionersErr = errors.New("rpc error: enforcement_client_upgrade_required")

	bucket := newTestBucket(mockService)

	err := bucket.FetchEnforcedBlocks(context.Background(), EnforcementOptions{})
	if err == nil {
		t.Fatal("expected upgrade-required error, got nil")
	}
}

func TestBucket_FetchEnforcedBlocks_ReusesFreshCacheOnOutage(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PACKER_ENFORCEMENT_CACHE_DIR", dir)

	// Seed a fresh advisory resolution into the cache.
	entry := &enforcementCacheEntry{
		Etag:         "W/\"cached\"",
		ResolvedAt:   time.Now().UTC(),
		PolicyMode:   policyModeAdvisory,
		ResolutionID: "res-cached",
		Blocks: []*EnforcedBlock{
			{VersionID: "ebv-cached", BlockContent: "provisioner \"shell\" {}", Ordinal: 10},
		},
	}
	if err := saveEnforcementCache("org-1", "prj-1", "test-bucket", entry); err != nil {
		t.Fatalf("seed cache: %v", err)
	}

	mockService := hcpPackerAPI.NewMockPackerClientService()
	mockService.ResolveEnforcedProvisionersErr = errors.New("service unavailable")

	bucket := newTestBucket(mockService)

	if err := bucket.FetchEnforcedBlocks(context.Background(), EnforcementOptions{}); err != nil {
		t.Fatalf("expected cached reuse (no error), got: %v", err)
	}
	if !bucket.EnforcementFromCache {
		t.Fatal("expected EnforcementFromCache=true")
	}
	if len(bucket.EnforcedBlocks) != 1 || bucket.EnforcedBlocks[0].VersionID != "ebv-cached" {
		t.Fatalf("expected cached block reused, got %+v", bucket.EnforcedBlocks)
	}
	if len(bucket.EnforcementWarnings) == 0 {
		t.Fatal("expected an advisory degradation warning")
	}
}
