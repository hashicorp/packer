// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package provenance

import (
	"fmt"
	"strings"
	"testing"
)

func TestDetectBuilderID(t *testing.T) {
	env := map[string]string{"GITHUB_WORKFLOW_REF": "acme/images/.github/workflows/build.yml@refs/heads/main"}
	if got, want := DetectBuilderID(env), env["GITHUB_WORKFLOW_REF"]; got != want {
		t.Fatalf("unexpected builder id %q, want %q", got, want)
	}

	if got, want := DetectBuilderID(map[string]string{}), DefaultLocalBuilderID; got != want {
		t.Fatalf("unexpected default builder id %q, want %q", got, want)
	}
}

func TestDetectInvocationID(t *testing.T) {
	env := map[string]string{"CI_PIPELINE_ID": "pipeline-42"}
	if got, want := DetectInvocationID(env), "pipeline-42"; got != want {
		t.Fatalf("unexpected invocation id %q, want %q", got, want)
	}
}

func TestDetectGitDependencyFromGitHubEnv(t *testing.T) {
	dependency, ok := detectGitDependency("", map[string]string{
		"GITHUB_REPOSITORY": "acme/images",
		"GITHUB_SHA":        "deadbeef",
		"GITHUB_REF":        "refs/heads/main",
	}, nil)
	if !ok {
		t.Fatalf("expected github dependency to be detected")
	}

	if got, want := dependency.URI, "git+https://github.com/acme/images@refs/heads/main"; got != want {
		t.Fatalf("unexpected uri %q, want %q", got, want)
	}
	if got, want := dependency.Digest["gitCommit"], "deadbeef"; got != want {
		t.Fatalf("unexpected commit %q, want %q", got, want)
	}
}

func TestDetectGitDependencyFallsBackToLocalGit(t *testing.T) {
	runner := func(_ string, args ...string) (string, error) {
		switch strings.Join(args, " ") {
		case "rev-parse HEAD":
			return "cafebabe", nil
		case "config --get remote.origin.url":
			return "", fmt.Errorf("missing remote")
		case "rev-parse --show-toplevel":
			return "/workspace/packer", nil
		case "rev-parse --abbrev-ref HEAD":
			return "main", nil
		default:
			return "", fmt.Errorf("unexpected git args %q", strings.Join(args, " "))
		}
	}

	dependency, ok := detectGitDependency("/workspace/packer", map[string]string{}, runner)
	if !ok {
		t.Fatalf("expected local git dependency to be detected")
	}

	if got, want := dependency.URI, "git+file:///workspace/packer@refs/heads/main"; got != want {
		t.Fatalf("unexpected uri %q, want %q", got, want)
	}
	if got, want := dependency.Digest["gitCommit"], "cafebabe"; got != want {
		t.Fatalf("unexpected commit %q, want %q", got, want)
	}
}

func TestDetectGitDependencyGracefullySkipsWhenUnavailable(t *testing.T) {
	runner := func(_ string, _ ...string) (string, error) {
		return "", fmt.Errorf("git unavailable")
	}

	if _, ok := detectGitDependency("/workspace/packer", map[string]string{}, runner); ok {
		t.Fatalf("expected dependency detection to skip unavailable git context")
	}
}
