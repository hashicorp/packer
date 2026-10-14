// Copyright IBM Corp. 2024, 2026
// SPDX-License-Identifier: BUSL-1.1

package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Enforcement cache implements the RFC 6.4 stale-cache semantics for the
// resolver. The freshness token of record is the resolver body's
// audit_context.etag; the CLI persists the last successful resolution so that,
// on a subsequent resolver outage, it can revalidate with If-None-Match and
// (subject to per-mode max-age) reuse the cached effective set.
//
// Max cache age (RFC 6.4 / 11): mandatory 300s, advisory 3600s.
const (
	mandatoryCacheMaxAge = 300 * time.Second
	advisoryCacheMaxAge  = 3600 * time.Second
)

// enforcementCacheEntry is the persisted snapshot of a successful resolution.
type enforcementCacheEntry struct {
	Etag         string           `json:"etag"`
	ResolvedAt   time.Time        `json:"resolved_at"`
	TTLSeconds   int32            `json:"ttl_seconds"`
	PolicyMode   string           `json:"policy_mode"`
	ResolutionID string           `json:"resolution_id"`
	Blocks       []*EnforcedBlock `json:"blocks"`
}

// fresh reports whether the cached resolution is still usable for the given
// policy mode, i.e. its age does not exceed the per-mode max age (RFC 6.4).
func (e *enforcementCacheEntry) fresh(mode string, now time.Time) bool {
	if e == nil || e.ResolvedAt.IsZero() {
		return false
	}
	maxAge := mandatoryCacheMaxAge
	if mode == policyModeAdvisory {
		maxAge = advisoryCacheMaxAge
	}
	return now.Sub(e.ResolvedAt) <= maxAge
}

// enforcementCacheDir returns the directory used to persist resolutions. It
// honors PACKER_ENFORCEMENT_CACHE_DIR for tests/overrides, otherwise uses the
// user cache dir. Returns "" (caching disabled) if no location is available.
func enforcementCacheDir() string {
	if dir := os.Getenv("PACKER_ENFORCEMENT_CACHE_DIR"); dir != "" {
		return dir
	}
	base, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(base, "packer", "enforcement")
}

// enforcementCacheKey produces a stable, filesystem-safe key for a bucket within
// a project so resolutions are scoped per org/project/bucket.
func enforcementCacheKey(orgID, projectID, bucketName string) string {
	sum := sha256.Sum256([]byte(orgID + "/" + projectID + "/" + bucketName))
	return hex.EncodeToString(sum[:]) + ".json"
}

// loadEnforcementCache reads the cached resolution for a bucket. A nil entry
// (and nil error) means no usable cache exists.
func loadEnforcementCache(orgID, projectID, bucketName string) (*enforcementCacheEntry, error) {
	dir := enforcementCacheDir()
	if dir == "" {
		return nil, nil
	}
	path := filepath.Join(dir, enforcementCacheKey(orgID, projectID, bucketName))
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var entry enforcementCacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		// A corrupt cache file is treated as absent rather than fatal.
		return nil, nil
	}
	return &entry, nil
}

// saveEnforcementCache persists a successful resolution. Failures to write are
// non-fatal to the build and are surfaced to the caller for logging only.
func saveEnforcementCache(orgID, projectID, bucketName string, entry *enforcementCacheEntry) error {
	dir := enforcementCacheDir()
	if dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, enforcementCacheKey(orgID, projectID, bucketName))
	return os.WriteFile(path, data, 0o600)
}
