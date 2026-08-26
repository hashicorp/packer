// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package provenance

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
	"github.com/hashicorp/packer-plugin-sdk/template"
	filebuilder "github.com/hashicorp/packer/builder/file"
	nullbuilder "github.com/hashicorp/packer/builder/null"
)

func TestDeriveSubjectsFromFiles(t *testing.T) {
	artifact := buildFileArtifact(t)
	defer func() { _ = artifact.Destroy() }()

	subjects, err := deriveSubjects(artifact)
	if err != nil {
		t.Fatalf("derive subjects: %v", err)
	}

	if len(subjects) != 1 {
		t.Fatalf("expected one subject, got %d", len(subjects))
	}

	if got, want := subjects[0].Name, "package.txt"; got != want {
		t.Fatalf("unexpected subject name %q, want %q", got, want)
	}

	expectedDigest := sha256.Sum256([]byte("Hello world!"))
	if got, want := subjects[0].Digest["sha256"], hex.EncodeToString(expectedDigest[:]); got != want {
		t.Fatalf("unexpected digest %q, want %q", got, want)
	}
}

func TestDeriveSubjectsFromIdentity(t *testing.T) {
	artifact := new(nullbuilder.NullArtifact)

	subjects, err := deriveSubjects(artifact)
	if err != nil {
		t.Fatalf("derive subjects: %v", err)
	}

	if len(subjects) != 1 {
		t.Fatalf("expected one subject, got %d", len(subjects))
	}

	if got, want := subjects[0].Name, artifact.BuilderId()+":"+artifact.Id(); got != want {
		t.Fatalf("unexpected subject name %q, want %q", got, want)
	}

	identity, err := deriveIdentityRecord(artifact)
	if err != nil {
		t.Fatalf("derive identity: %v", err)
	}

	if _, ok := identity["state"]; !ok {
		t.Fatalf("expected identity state for cloud-style artifact")
	}

	encodedIdentity, err := json.Marshal(identity)
	if err != nil {
		t.Fatalf("marshal identity: %v", err)
	}

	expectedDigest := sha256.Sum256(encodedIdentity)
	if got, want := subjects[0].Digest["sha256"], hex.EncodeToString(expectedDigest[:]); got != want {
		t.Fatalf("unexpected digest %q, want %q", got, want)
	}

	state := artifact.State(registryimage.ArtifactStateURI)
	if state == nil {
		t.Fatalf("expected registry state")
	}
}

func buildFileArtifact(t *testing.T) packersdk.Artifact {
	t.Helper()

	target := filepath.Join(t.TempDir(), "package.txt")
	config := mustTemplateJSON(t, map[string]any{
		"builders": []map[string]string{{
			"type":    "file",
			"target":  target,
			"content": "Hello world!",
		}},
	})
	tpl, err := template.Parse(strings.NewReader(config))
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	var builder filebuilder.Builder
	_, warnings, err := builder.Prepare(tpl.Builders["file"].Config)
	if err != nil {
		t.Fatalf("prepare builder: %v", err)
	}
	if len(warnings) > 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}

	artifact, err := builder.Run(context.Background(), packersdk.TestUi(t), nil)
	if err != nil {
		t.Fatalf("run builder: %v", err)
	}

	return artifact
}

func mustTemplateJSON(t *testing.T, value any) string {
	t.Helper()

	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal template config: %v", err)
	}

	return string(encoded)
}
