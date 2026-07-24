// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package provenance

import "testing"

func TestBuildSLSAPredicateDefaults(t *testing.T) {
	predicate := BuildSLSAPredicate(PredicateInput{})

	if got, want := predicate.BuildDefinition.BuildType, DefaultBuildType; got != want {
		t.Fatalf("unexpected build type %q, want %q", got, want)
	}

	if got, want := predicate.RunDetails.Builder.ID, DefaultLocalBuilderID; got != want {
		t.Fatalf("unexpected builder id %q, want %q", got, want)
	}

	if predicate.BuildDefinition.InternalParameters["packerVersion"] == "" {
		t.Fatalf("expected packerVersion to be populated")
	}

	if predicate.RunDetails.Builder.Version["packer"] == "" {
		t.Fatalf("expected builder version to be populated")
	}
}

func TestBuildSLSAPredicateIncludesByproducts(t *testing.T) {
	predicate := BuildSLSAPredicate(PredicateInput{
		BuildType: "https://packer.io/buildtypes/json/v1",
		Byproducts: []Byproduct{{
			Name: "cloud-artifact-identity",
			Content: map[string]any{
				"builderId": "packer.null",
			},
		}},
	})

	if got, want := len(predicate.RunDetails.Byproducts), 1; got != want {
		t.Fatalf("unexpected byproduct count %d, want %d", got, want)
	}

	if got, want := predicate.BuildDefinition.BuildType, "https://packer.io/buildtypes/json/v1"; got != want {
		t.Fatalf("unexpected build type %q, want %q", got, want)
	}
	if got, want := predicate.RunDetails.Byproducts[0].Name, "cloud-artifact-identity"; got != want {
		t.Fatalf("unexpected byproduct name %q, want %q", got, want)
	}
}
