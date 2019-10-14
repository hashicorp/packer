package hcl2template

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
) 

type Artifacts map[ArtifactRef]*Artifact

type Artifact struct {
	Type string
	Name string

	DeclRange hcl.Range

	Config hcl.Body
}

func (a *Artifact) Ref() ArtifactRef {
	return ArtifactRef{
		Type: a.Type,
		Name: a.Name,
	}
}

type ArtifactRef struct {
	Type string
	Name string
}

// NoArtifact is the zero value of ArtifactRef, representing the absense of an
// artifact.
var NoArtifact ArtifactRef

func artifactRefFromAbsTraversal(t hcl.Traversal) (ArtifactRef, hcl.Diagnostics) {
	var diags hcl.Diagnostics
	if len(t) != 3 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid artifact reference",
			Detail:   "An artifact reference must have three parts separated by periods: the keyword \"artifact\", the builder type name, and the artifact name.",
			Subject:  t.SourceRange().Ptr(),
		})
		return NoArtifact, diags
	}

	if t.RootName() != "artifact" {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid artifact reference",
			Detail:   "The first part of an artifact reference must be the keyword \"artifact\".",
			Subject:  t[0].SourceRange().Ptr(),
		})
		return NoArtifact, diags
	}
	btStep, ok := t[1].(hcl.TraverseAttr)
	if !ok {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid artifact reference",
			Detail:   "The second part of an artifact reference must be an identifier giving the builder type of the artifact.",
			Subject:  t[1].SourceRange().Ptr(),
		})
		return NoArtifact, diags
	}
	nameStep, ok := t[2].(hcl.TraverseAttr)
	if !ok {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid artifact reference",
			Detail:   "The third part of an artifact reference must be an identifier giving the name of the artifact.",
			Subject:  t[2].SourceRange().Ptr(),
		})
		return NoArtifact, diags
	}

	return ArtifactRef{
		Type: btStep.Name,
		Name: nameStep.Name,
	}, diags
}

func (r ArtifactRef) String() string {
	return fmt.Sprintf("%s.%s", r.Type, r.Name)
}

// decodeBodyWithoutSchema is a generic alternative to hcldec.Decode that
// just extracts whatever attributes are present and rejects any nested blocks,
// for compatibility with legacy builders that can't provide explicit schema.
func decodeBodyWithoutSchema(body hcl.Body, ctx *hcl.EvalContext) (cty.Value, hcl.Diagnostics) {
	attrs, diags := body.JustAttributes()
	vals := make(map[string]cty.Value)
	for name, attr := range attrs {
		val, moreDiags := attr.Expr.Value(ctx)
		diags = append(diags, moreDiags...)
		vals[name] = val
	}
	return cty.ObjectVal(vals), diags
}
