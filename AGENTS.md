# AGENTS.md

## Purpose

Packer is HashiCorp's open-source (BUSL-1.1) CLI for building automated machine
images. It is a Go program that orchestrates pluggable **builders**,
**provisioners**, **post-processors**, and **data sources** through the
`hashicorp/packer-plugin-sdk`, driven by HCL2 templates (legacy JSON templates are
still supported). There is no database, no gRPC/gateway service, and no protobuf
API surface in this repository.

Use this file as the repo-specific operating contract. Prefer the existing `make`
targets and established patterns over invented workflows.

## Default Working Mode

- Keep chat concise. Share decision-worthy context, short progress updates, and
  summarized command output; do not paste large logs or diffs unless asked.
- Lead with the answer or outcome. Skip motivational filler and obvious recaps.
- Start broad, unclear, or "review/investigate/audit" tasks with read-only
  discovery and report findings before editing.
- Ask for confirmation before non-trivial implementation when there are multiple
  viable approaches, the request is ambiguous, or the change touches the plugin
  SDK boundary, HCL2 template parsing, the command surface, CI/release workflows,
  or security-relevant behavior.
- Proceed without another confirmation when the user explicitly asks to implement,
  fix, add, remove, regenerate, or run an approved command, or for trivial typo,
  formatting, or docs cleanup.
- When blocked by missing credentials, cloud access, or unclear intent, stop and
  ask. Do not improvise around those blockers.

## Scope And Boundaries

- Work only inside this repository unless the user explicitly requests cross-repo
  changes (e.g., the plugin SDK or a specific plugin repo).
- Do not read from, write to, or execute files outside the workspace, including
  `/tmp`, `~`, or `/etc`. Create temporary artifacts inside the repository only.
- Treat files marked `// Code generated ... DO NOT EDIT` as derived output. Change
  the source struct/`//go:generate` directive and regenerate; never hand-edit the
  generated file. This includes `*.hcl2spec.go` and `*_enumer.go`.
- Do not hand-edit vendored content, `go.sum`, or other generated artifacts unless
  the task is explicitly about that output.
- Do not access production systems, cloud consoles, Vault, 1Password, or cloud
  accounts unless explicitly asked and safely configured. Acceptance tests boot
  real infrastructure and may cost money — never run them without explicit intent.
- Do not inspect, print, copy, or persist secrets from shell history, env vars,
  `.envrc`, CI configuration, or credential stores.
- Do not run destructive git or remote-system actions without explicit approval.

## Approved Commands

Run `make help` to list targets. Common ones:

### Build

- `make dev` builds and installs a development binary to `bin/packer` (requires a
  prerelease tag in `version/version.go`).
- `go build -o bin/packer .` is the minimal build if `make` is unavailable.

### Generate

- `make generate` runs `go generate ./...` to rebuild dynamically generated code
  (HCL2 specs via `packer-sdc mapstructure-to-hcl2`, enumer output, fixer
  deprecations). Run this after changing any config struct or adding a component.
- `make generate-check` verifies generated code is up to date (fails on drift).

### Format & Lint

- `make fmt` runs `go fmt ./...`; `make fmt-check` fails if code is not formatted.
- `make lint` runs `golangci-lint` over the repo (config in `.golangci.yml`). Use
  `PKG_NAME=<dir> make lint` to scope. `make ci-lint` lints only newly changed
  files against `origin/main`.

### Test

- `make test` runs unit tests (`go vet` + `go test`, 3m timeout). Prefer
  `TEST=./path/... make test` or `TESTARGS="-run TestName" make test` to scope.
- `make testrace` runs unit tests with the race detector.
- `make testacc` runs acceptance tests with `PACKER_ACC=1`. **These are slow, boot
  real machines/cloud resources, and can cost money.** Only run when the user
  explicitly asks and prerequisites are configured.

## Architecture Rules

- Components are plugins behind SDK interfaces. New behavior belongs in the right
  component type: **builders** create machines/artifacts, **provisioners** run
  against a machine via a **communicator** (SSH/WinRM/Docker), **post-processors**
  transform/act on artifacts, **data sources** fetch inputs. Prefer post-processors
  and provisioners for cross-builder ("plugin-independent") features, since they
  operate on the SDK's `Artifact`/`Communicator` abstractions rather than a
  specific builder.
- Component config is defined by a Go struct plus a generated `*.hcl2spec.go`.
  After editing config fields, run `make generate` and commit the regenerated file.
- Keep changes backward compatible for existing templates. New fields must be
  optional with sensible defaults; do not change the meaning of existing fields.
- HCL2 parsing lives in `hcl2template/`; the CLI commands live in `command/`;
  template fixers live in `fix/`. When adding a command, register it in
  `commands.go`. When deprecating/renaming config, add a fixer.
- Pass `context.Context` through build/orchestration and network-facing paths.
- Use the SDK's existing helpers and error patterns instead of ad hoc equivalents.

## Go Style

- Follow standard Go formatting and the repository linter (`.golangci.yml`).
- Every Go/HCL source file carries the copyright + `SPDX-License-Identifier:
  BUSL-1.1` header (managed by `copywrite`, config in `.copywrite.hcl`). New files
  must include it.
- Keep packages lowercase and concise; export names only when they must cross a
  package boundary.
- Use domain terms already present in the repo: `builder`, `provisioner`,
  `post-processor`, `datasource`, `communicator`, `artifact`, `template`,
  `plugin`, `fixer`, `hcl2template`.
- Return wrapped errors with useful context and preserve the original error.
- Match surrounding struct layout, constructors, and table-driven test style.

## Testing And Validation

- Run the narrowest test set that proves the change, then broaden when the blast
  radius justifies it (`TEST=./command/... make test`, then `make test`).
- For PR-bound changes, run the applicable repo maintenance targets before
  handoff: `make fmt` for Go edits, `make generate` plus `make generate-check`
  when source changes affect generated files, and `make ci-lint` when the
  change should satisfy the same lint expectations as CI. If one of these is
  skipped, state why.
- Add or update unit tests for behavior changes. If a bug fix is not covered,
  explain why in the handoff.
- After changing config structs or adding components, run `make generate` and
  `make generate-check` so generated code stays in sync.
- Do not leave formatting or generated-file drift for the user to discover at
  push or review time; run the relevant `make` target and include the resulting
  updates in the same change.
- Do not run acceptance tests (`PACKER_ACC=1`) casually — they are slow and may
  provision billable resources. State when validation was skipped and why.

## Review Guidelines

Prioritize findings over summary; do not edit unless asked.

- Check correctness first: behavior, edge cases, regressions, nil handling,
  context handling, and backward compatibility for existing templates.
- Check that config changes are additive and that defaults preserve prior behavior.
- Check that generated code (`*.hcl2spec.go`, enumer output) was regenerated when
  its source changed — flag generated-code drift.
- Check new commands are registered in `commands.go` and deprecations have fixers.
- Check tests cover the changed behavior, not just compilation.
- Flag risky changes to CI, release (`.release/`), CODEOWNERS, or security scanning
  unless they are the explicit task.

## Commit And PR Guidelines

- Do not commit, branch, push, or open PRs unless explicitly requested.
- Keep commits focused and logically atomic; use short imperative subjects.
- Separate regenerated output from behavioral changes where it can be done cleanly.
- Follow `.github/PULL_REQUEST_TEMPLATE.md`. PR text should say what changed and
  why, and note any backward-compatibility impact on existing templates.
- Add a changelog note (`CHANGELOG.md`) for user-facing changes.
- Never include secrets, tokens, or credentials in commits, PR text, or logs.

## Completion Checklist

Before finishing a task, verify the following when applicable.

- Source-of-truth files were changed instead of generated output.
- Applicable repo maintenance commands were run before handoff: `make fmt` for
  Go edits, `make generate` and `make generate-check` for generated-code
  changes, and `make ci-lint` for PR-relevant lint validation, or the reason
  they were skipped is documented.
- `make generate` was run when config structs or components changed, and the
  regenerated files are committed.
- Relevant unit tests and `make fmt-check` / `make lint` were run, or the reason
  they were not run is documented.
- New commands are registered in `commands.go`; deprecations have fixers.
- SPDX/copyright headers are present on new source files.
- Final handoff states what changed, what was validated, and any remaining risk or
  follow-up.