// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package provenance

import "testing"

func TestWrapInToto(t *testing.T) {
	statement := WrapInToto(
		[]Subject{{Name: "artifact.bin", Digest: DigestSet{"sha256": "abc123"}}},
		SLSAProvenanceV1PredicateType,
		BuildSLSAPredicate(PredicateInput{}),
	)

	if got, want := statement.Type, StatementType; got != want {
		t.Fatalf("unexpected statement type %q, want %q", got, want)
	}

	if got, want := statement.PredicateType, SLSAProvenanceV1PredicateType; got != want {
		t.Fatalf("unexpected predicate type %q, want %q", got, want)
	}

	if got, want := len(statement.Subject), 1; got != want {
		t.Fatalf("unexpected subject count %d, want %d", got, want)
	}
}
