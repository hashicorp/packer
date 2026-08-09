# Provenance Post-Processor

The `provenance` post-processor writes in-toto attestations for the incoming
artifact.

It supports:

- SLSA provenance statements enriched with source-control and CI metadata.
- SBOM sidecars plus SBOM attestations.
- Optional DSSE signing with a configurable signer and verifier.

Signing defaults to `none`, which writes the unsigned JSON statement.

## Source detection

The provenance statement records the build's source repository as a resolved
dependency. Packer detects this from the Git repository containing the current
working directory (the directory Packer runs in). Set `source_uri` to override
the detected value, for example when the build runs outside the source checkout
or the remote URL should be normalized.

Signing modes:

- `key` signs with a PEM private key and verifies with either the signer's
  derived public key or an explicit verifier PEM.
- `kms` signs with a KMS or Vault URI such as `awskms://...`, `gcpkms://...`,
  `azurekms://...`, or `hashivault://...`. Verification uses the fetched KMS
  public key or an explicit verifier PEM.
- `keyless` signs with an ephemeral keypair and a Fulcio-issued certificate.
  It requires an ambient OIDC token, such as `SIGSTORE_ID_TOKEN`, or a CI
  provider token that can be exchanged for a Sigstore identity. When using the
  built-in verifier path, also configure the expected signing identity and OIDC
  issuer. Set an optional trusted-root JSON path to pin verification to a
  specific Sigstore root; otherwise the public Sigstore trusted root is fetched.
  Keyless signing can also emit a Sigstore bundle sidecar and upload to Rekor.

Example:

```hcl
post-processor "provenance" {
  build_type = "https://packer.io/buildtypes/hcl2/v1"
  template   = "ubuntu.pkr.hcl"
  only_builds = ["qemu.ubuntu"]
  user_variables = {
    region = "us-east-1"
  }
  sbom = true
}
```

Signed example:

```hcl
post-processor "provenance" {
  signing_mode = "key"
  signer       = "keys/provenance-signing.pem"
  verifier     = "keys/provenance-signing.pub.pem"
  sbom         = true
}
```

KMS example:

```hcl
post-processor "provenance" {
  signing_mode = "kms"
  signer       = "awskms://alias/packer-provenance"
  sbom         = true
}
```

Keyless example:

```hcl
post-processor "provenance" {
  signing_mode        = "keyless"
  fulcio_url          = "https://fulcio.sigstore.dev"
  rekor_url           = "https://rekor.sigstore.dev"
  upload_tlog         = true
  keyless_identity    = "https://github.com/hashicorp/packer/.github/workflows/build.yml@refs/heads/main"
  keyless_oidc_issuer = "https://token.actions.githubusercontent.com"
  trusted_root_path   = "sigstore-trusted-root.json"
}
```

With keyless signing, Packer writes an additional `*.sigstore.json` sidecar next
to each signed attestation. When `upload_tlog = true`, that bundle includes
Rekor-backed transparency evidence for `packer verify-attestation -bundle ...`.

During a build, keyless verification enforces the Fulcio certificate chain and
the configured identity policy. For Rekor-backed verification, run
`packer verify-attestation` with the generated Sigstore bundle and the
`-require-rekor` and/or `-require-timestamp` flags.

## SLSA levels and CI

SLSA Build levels are mostly properties of the build platform, not the build
tool. Packer is a tool, so its reach is:

| SLSA Build level | Requirement | Packer's role | What Packer provides |
|---|---|---|---|
| **L1** | Provenance exists and is distributed | Fully in Packer | Provenance generation |
| **L2** | Provenance signed by a hosted platform | Packer signs via CI OIDC identity | Keyless signing in CI |
| **L3** | Hardened platform; build steps cannot reach the signing key | Platform property; Packer is compatible | Delegated-signing pattern |
| **L4** | — | Not defined in SLSA v1.0 | — |

Packer generates SLSA Provenance v1 and reaches Build L1, and L2 when run on a
hosted CI with keyless signing. L3 is a property of the build platform: it
requires the signing key to be unreachable by the build steps. Packer does not
confer L3 on its own, but the delegated-signing pattern is compatible with an L3
platform.

Reference GitHub Actions workflows for both the L2 keyless pattern and the
L3-compatible delegated-signing pattern live under
[`examples/ci/`](../../examples/ci/).