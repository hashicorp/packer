// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
)

const (
	_ = iota
	InvalidClientConfig
)

// ClientError represents a generic error for the Cloud Packer Service client.
type ClientError struct {
	StatusCode uint
	Err        error
}

// Error returns the string message for some ClientError.
func (c *ClientError) Error() string {
	return fmt.Sprintf("status %d: err %v", c.StatusCode, c.Err)
}

var errCodeRegex = regexp.MustCompilePOSIX(`[Cc]ode"?:([0-9]+)`)

// CheckErrorCode checks the error string for err for some code and returns true
// if the code is found. Ideally this function should use status.FromError
// https://pkg.go.dev/google.golang.org/grpc/status#pkg-functions but that
// doesn't appear to work for all of the Cloud Packer Service response errors.
func CheckErrorCode(err error, code codes.Code) bool {
	if err == nil {
		return false
	}

	// If the error string doesn't match the code we're looking for, we
	// can ignore it and return false immediately.
	matches := errCodeRegex.FindStringSubmatch(err.Error())
	if len(matches) == 0 {
		return false
	}

	// Safe to ignore the error here since the regex's submatch is always a
	// valid integer given the format ([0-9]+)
	errCode, _ := strconv.Atoi(matches[1])
	return errCode == int(code)
}

// Canonical enforced-provisioner wire error_code reasons (RFC 14.3). These are
// carried in the resolver error body via google.rpc.ErrorInfo and rendered into
// the gateway JSON error, so they are detectable in the error string.
const (
	EnforcementReasonResolverUnavailable   = "enforcement_resolver_unavailable"
	EnforcementReasonResolutionIncomplete  = "enforcement_resolution_incomplete"
	EnforcementReasonRevokedLinkBlocking   = "enforcement_revoked_link_blocking"
	EnforcementReasonDataIntegrityError    = "enforcement_data_integrity_error"
	EnforcementReasonClientUpgradeRequired = "enforcement_client_upgrade_required"
)

// EnforcementErrorReason extracts the canonical enforced-provisioner error_code
// reason from an error returned by the resolver, if present. Returns "" when no
// known reason is found.
func EnforcementErrorReason(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	for _, reason := range []string{
		EnforcementReasonResolverUnavailable,
		EnforcementReasonResolutionIncomplete,
		EnforcementReasonRevokedLinkBlocking,
		EnforcementReasonDataIntegrityError,
		EnforcementReasonClientUpgradeRequired,
	} {
		if strings.Contains(msg, reason) {
			return reason
		}
	}
	return ""
}

// IsClientUpgradeRequired reports whether the resolver rejected this CLI as too
// old to enforce a mandatory bucket (RFC 6.4 / 12.4 / 14.3, HTTP 426).
func IsClientUpgradeRequired(err error) bool {
	return EnforcementErrorReason(err) == EnforcementReasonClientUpgradeRequired ||
		CheckErrorCode(err, codes.Code(26)) // gateway maps 426 → no native gRPC code; reason match is primary
}
