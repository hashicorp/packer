# Provenance CI reference workflows

These are copy-paste reference workflows for producing SLSA provenance with the
Packer `provenance` post-processor. They are **not** wired into this
repository's own CI — copy them into your project's `.github/workflows/`
directory and adapt the template and artifact paths.

## SLSA levels and Packer

SLSA Build levels are mostly properties of the build **platform**, not the
build **tool**. Packer is a tool, so its reach is:

| SLSA Build level | Requirement | Packer's role | What Packer provides |
|---|---|---|---|
| **L1** | Provenance exists and is distributed | Fully in Packer | Provenance generation |
| **L2** | Provenance signed by a hosted platform | Packer signs via CI OIDC identity | Keyless signing in CI |
| **L3** | Hardened platform; build steps cannot reach the signing key | Platform property; Packer is compatible | Delegated-signing pattern |
| **L4** | — | Not defined in SLSA v1.0 | — |

Packer generates SLSA Provenance v1 and reaches Build L1, and L2 when run on a
hosted CI with keyless signing. L3 is a property of the build platform: it
requires the signing key to be unreachable by the build steps. Packer does not
confer L3 on its own, but the delegated-signing pattern below is compatible with
an L3 platform.

## Workflows

- [`github-actions-l2-keyless.yml`](github-actions-l2-keyless.yml) — L2: Packer
  signs provenance keyless using the workflow's OIDC identity, uploads to Rekor,
  and verifies the signed attestation with `packer verify-attestation`.
- [`github-actions-l3-delegated.yml`](github-actions-l3-delegated.yml) —
  L3-compatible: the build job only builds and publishes a digest; provenance
  generation and signing are delegated to an isolated reusable workflow so the
  build steps cannot reach the signing material.
