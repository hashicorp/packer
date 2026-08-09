// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package provenance

import packerversion "github.com/hashicorp/packer/version"

const (
	DefaultBuildType              = "https://packer.io/buildtypes/hcl2/v1"
	DefaultLocalBuilderID         = "https://packer.io/local-build"
	SLSAProvenanceV1PredicateType = "https://slsa.dev/provenance/v1"
)

type PredicateInput struct {
	BuildType            string
	ExternalParameters   map[string]interface{}
	InternalParameters   map[string]interface{}
	ResolvedDependencies []ResolvedDependency
	BuilderID            string
	Byproducts           []Byproduct
	InvocationID         string
	StartedOn            string
	FinishedOn           string
}

type SLSAProvenancePredicate struct {
	BuildDefinition BuildDefinition `json:"buildDefinition"`
	RunDetails      RunDetails      `json:"runDetails"`
}

type BuildDefinition struct {
	BuildType            string                 `json:"buildType"`
	ExternalParameters   map[string]interface{} `json:"externalParameters"`
	InternalParameters   map[string]interface{} `json:"internalParameters,omitempty"`
	ResolvedDependencies []ResolvedDependency   `json:"resolvedDependencies,omitempty"`
}

type ResolvedDependency struct {
	URI    string    `json:"uri"`
	Digest DigestSet `json:"digest,omitempty"`
}

type RunDetails struct {
	Builder    Builder     `json:"builder"`
	Metadata   Metadata    `json:"metadata,omitempty"`
	Byproducts []Byproduct `json:"byproducts,omitempty"`
}

type Builder struct {
	ID      string            `json:"id"`
	Version map[string]string `json:"version,omitempty"`
}

type Metadata struct {
	InvocationID string `json:"invocationId,omitempty"`
	StartedOn    string `json:"startedOn,omitempty"`
	FinishedOn   string `json:"finishedOn,omitempty"`
}

type Byproduct struct {
	Name    string      `json:"name"`
	Content interface{} `json:"content,omitempty"`
}

func BuildSLSAPredicate(input PredicateInput) SLSAProvenancePredicate {
	buildType := input.BuildType
	if buildType == "" {
		buildType = DefaultBuildType
	}

	builderID := input.BuilderID
	if builderID == "" {
		builderID = DefaultLocalBuilderID
	}

	externalParameters := map[string]interface{}{}
	for key, value := range input.ExternalParameters {
		externalParameters[key] = value
	}

	internalParameters := map[string]interface{}{
		"packerVersion": packerversion.String(),
	}
	for key, value := range input.InternalParameters {
		internalParameters[key] = value
	}

	predicate := SLSAProvenancePredicate{
		BuildDefinition: BuildDefinition{
			BuildType:            buildType,
			ExternalParameters:   externalParameters,
			InternalParameters:   internalParameters,
			ResolvedDependencies: input.ResolvedDependencies,
		},
		RunDetails: RunDetails{
			Builder: Builder{
				ID: builderID,
				Version: map[string]string{
					"packer": packerversion.String(),
				},
			},
			Byproducts: input.Byproducts,
		},
	}

	if input.InvocationID != "" || input.StartedOn != "" || input.FinishedOn != "" {
		predicate.RunDetails.Metadata = Metadata{
			InvocationID: input.InvocationID,
			StartedOn:    input.StartedOn,
			FinishedOn:   input.FinishedOn,
		}
	}

	return predicate
}
