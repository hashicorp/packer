# Provenance CI reference workflows

These are copy-paste reference workflows for producing SLSA provenance with the
Packer `provenance` post-processor. They are **not** wired into this
repository's own CI — copy them into your project's `.github/workflows/`
directory and adapt the template and artifact paths.

## SLSA levels: what Packer can honestly provide

SLSA Build levels are mostly properties of the build **platform**, not the
build **tool**. Packer is a tool, so its honest reach is:

| SLSA Build level | Requirement | Packer's role | Honest target |
|---|---|---|---|
| **L1** | Provenance exists and is distributed | Fully in Packer | Ship it |
| **L2** | Provenance signed by a hosted platform | Packer signs via CI OIDC identity | Default in CI |
| **L3** | Hardened platform; build steps cannot reach the signing key | Packer is *compatible*, cannot confer alone | "L3-compatible" via an isolated signer |
| **L4** | — | Does not exist in SLSA v1.0 | Do not claim |

Rules for claims:

- You MAY say Packer "generates SLSA Provenance v1" and "achieves Build L1, and
  L2 when run on a hosted CI with keyless signing."
- You MUST NOT say "Packer is SLSA L3." L3 requires the signing key to be
  unreachable by the build steps, which is a platform property. Reach L3 by
  delegating signing to an isolated workflow.

## Workflows

- [`github-actions-l2-keyless.yml`](github-actions-l2-keyless.yml) — L2: Packer
  signs provenance keyless using the workflow's OIDC identity, uploads to Rekor,
  and verifies the signed attestation with `packer verify-attestation`.
- [`github-actions-l3-delegated.yml`](github-actions-l3-delegated.yml) —
  L3-compatible: the build job only builds and publishes a digest; provenance
  generation and signing are delegated to an isolated reusable workflow so the
  build steps cannot reach the signing material.
