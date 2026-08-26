// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package provenance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
)

type DigestSet map[string]string

type Subject struct {
	Name   string    `json:"name"`
	Digest DigestSet `json:"digest"`
}

func DeriveSubjects(artifact packersdk.Artifact) ([]Subject, error) {
	return deriveSubjects(artifact)
}

func DeriveIdentityRecord(artifact packersdk.Artifact) (map[string]any, error) {
	return deriveIdentityRecord(artifact)
}

func deriveSubjects(artifact packersdk.Artifact) ([]Subject, error) {
	if artifact == nil {
		return nil, fmt.Errorf("artifact is nil")
	}

	files := artifact.Files()
	if len(files) > 0 {
		subjects := make([]Subject, 0, len(files))
		for _, file := range files {
			digest, err := sha256File(file)
			if err != nil {
				return nil, fmt.Errorf("hash %q: %w", file, err)
			}

			subjects = append(subjects, Subject{
				Name: filepath.Base(file),
				Digest: DigestSet{
					"sha256": digest,
				},
			})
		}

		return subjects, nil
	}

	identity, err := deriveIdentityRecord(artifact)
	if err != nil {
		return nil, err
	}

	canonicalIdentity, err := json.Marshal(identity)
	if err != nil {
		return nil, fmt.Errorf("marshal artifact identity: %w", err)
	}

	digest := sha256.Sum256(canonicalIdentity)

	return []Subject{{
		Name: fmt.Sprintf("%s:%s", artifact.BuilderId(), artifact.Id()),
		Digest: DigestSet{
			"sha256": hex.EncodeToString(digest[:]),
		},
	}}, nil
}

func deriveIdentityRecord(artifact packersdk.Artifact) (map[string]any, error) {
	if artifact == nil {
		return nil, fmt.Errorf("artifact is nil")
	}

	record := map[string]any{
		"builderId": artifact.BuilderId(),
		"id":        artifact.Id(),
	}

	state := artifact.State(registryimage.ArtifactStateURI)
	if state == nil {
		return record, nil
	}

	normalizedState, err := normalizeJSONValue(state)
	if err != nil {
		return record, nil
	}

	record["state"] = normalizedState
	return record, nil
}

func normalizeJSONValue(value any) (any, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	var decoded any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return nil, err
	}

	return decoded, nil
}

func sha256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
