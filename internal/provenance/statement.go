// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package provenance

const StatementType = "https://in-toto.io/Statement/v1"

type Statement struct {
	Type          string      `json:"_type"`
	Subject       []Subject   `json:"subject"`
	PredicateType string      `json:"predicateType"`
	Predicate     interface{} `json:"predicate"`
}

func WrapInToto(subjects []Subject, predicateType string, predicate interface{}) Statement {
	return Statement{
		Type:          StatementType,
		Subject:       subjects,
		PredicateType: predicateType,
		Predicate:     predicate,
	}
}
