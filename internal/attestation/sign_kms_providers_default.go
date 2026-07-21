// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

//go:build !kms_cherrypick

// Default builds include every KMS/Vault provider so that all signing_mode=kms
// URIs work out of the box. Builds using the "kms_cherrypick" tag exclude this
// file and instead opt into individual providers via the per-provider tags
// (kms_aws, kms_azure, kms_gcp, kms_hashivault).

package attestation

import (
	_ "github.com/sigstore/sigstore/pkg/signature/kms/aws"
	_ "github.com/sigstore/sigstore/pkg/signature/kms/azure"
	_ "github.com/sigstore/sigstore/pkg/signature/kms/gcp"
	_ "github.com/sigstore/sigstore/pkg/signature/kms/hashivault"
)
