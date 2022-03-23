package addrs

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
)

// Reference describes a reference to an address with source location
// information.
type Reference struct {
	Subject     Referenceable
	SourceRange hcl.Range
	Remaining   hcl.Traversal
}

// ParseRef attempts to extract a referencable address from the prefix of the
// given traversal, which must be an absolute traversal or this function
// will panic.
//
// If no error diagnostics are returned, the returned reference includes the
// address that was extracted, the source range it was extracted from, and any
// remaining relative traversal that was not consumed as part of the
// reference.
//
// If error diagnostics are returned then the Reference value is invalid and
// must not be used.
func ParseRef(traversal hcl.Traversal) (*Reference, hcl.Diagnostics) {
	ref, diags := parseRef(traversal)

	// Normalize a little to make life easier for callers.
	if ref != nil {
		if len(ref.Remaining) == 0 {
			ref.Remaining = nil
		}
	}

	return ref, diags
}

func parseRef(traversal hcl.Traversal) (*Reference, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	root := traversal.RootName()
	rootRange := traversal[0].SourceRange()

	switch root {

	case "local":
		name, rng, remain, diags := parseSingleAttrRef(traversal)
		return &Reference{
			Subject:     LocalValue{Name: name},
			SourceRange: rng,
			Remaining:   remain,
		}, diags

	case "var":
		name, rng, remain, diags := parseSingleAttrRef(traversal)
		return &Reference{
			Subject:     InputVariable{Name: name},
			SourceRange: rng,
			Remaining:   remain,
		}, diags

	case "data":
		if len(traversal) < 3 {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid reference",
				Detail:   `The "data" object must be followed by two attribute names: the data source type and the resource name.`,
				Subject:  traversal.SourceRange().Ptr(),
			})
			return nil, diags
		}
		remain := traversal[1:] // trim off "data" so we can use our shared resource reference parser
		return parseResourceRef(DataResourceMode, rootRange, remain)

	default:
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unhandled reference type",
			Detail:   `Currently parseRef can only parse "var" references.`,
			Subject:  &rootRange,
		})
	}
	return nil, diags
}

// parseResourceRef parses any kind of resource reference that is not a local or
// a var. It is handy to tell what is being referenced in a datasource, and in
// the future for a build. This function was taken from terraform core, hence
// why it is already refactored.
func parseResourceRef(mode ResourceMode, startRange hcl.Range, traversal hcl.Traversal) (*Reference, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	var typeName, name string
	switch tt := traversal[0].(type) { // Could be either root or attr, depending on our resource mode
	case hcl.TraverseRoot:
		typeName = tt.Name
	case hcl.TraverseAttr:
		typeName = tt.Name
	default:
		// If it isn't a TraverseRoot then it must be a "data" reference.
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid reference",
			Detail:   `The "data" object does not support this operation.`,
			Subject:  traversal[0].SourceRange().Ptr(),
		})
		return nil, diags
	}

	attrTrav, ok := traversal[1].(hcl.TraverseAttr)
	if !ok {
		var what string
		switch mode {
		case DataResourceMode:
			what = "data source"
		default:
			what = "build type"
		}
		diags = diags.Append(&hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid reference",
			Detail:   fmt.Sprintf(`A reference to a %s must be followed by at least one attribute access, specifying the resource name.`, what),
			Subject:  traversal[1].SourceRange().Ptr(),
		})
		return nil, diags
	}
	name = attrTrav.Name
	rng := hcl.RangeBetween(startRange, attrTrav.SrcRange)
	remain := traversal[2:]

	resourceAddr := Resource{
		Mode: mode,
		Type: typeName,
		Name: name,
	}
	resourceInstAddr := ResourceInstance{
		Resource: resourceAddr,
		Key:      NoKey,
	}

	if len(remain) == 0 {
		// This might actually be a reference to the collection of all instances
		// of the resource, but we don't have enough context here to decide
		// so we'll let the caller resolve that ambiguity.
		return &Reference{
			Subject:     resourceAddr,
			SourceRange: rng,
		}, diags
	}

	if idxTrav, ok := remain[0].(hcl.TraverseIndex); ok {
		var err error
		resourceInstAddr.Key, err = ParseInstanceKey(idxTrav.Key)
		if err != nil {
			diags = diags.Append(&hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Invalid index key",
				Detail:   fmt.Sprintf("Invalid index for resource instance: %s.", err),
				Subject:  &idxTrav.SrcRange,
			})
			return nil, diags
		}
		remain = remain[1:]
		rng = hcl.RangeBetween(rng, idxTrav.SrcRange)
	}

	return &Reference{
		Subject:     resourceInstAddr,
		SourceRange: rng,
		Remaining:   remain,
	}, diags
}

func parseSingleAttrRef(traversal hcl.Traversal) (string, hcl.Range, hcl.Traversal, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	root := traversal.RootName()
	rootRange := traversal[0].SourceRange()

	if len(traversal) < 2 {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid reference",
			Detail:   fmt.Sprintf("The %q object cannot be accessed directly. Instead, access one of its attributes.", root),
			Subject:  &rootRange,
		})
		return "", hcl.Range{}, nil, diags
	}
	if attrTrav, ok := traversal[1].(hcl.TraverseAttr); ok {
		return attrTrav.Name, hcl.RangeBetween(rootRange, attrTrav.SrcRange), traversal[2:], diags
	}
	diags = diags.Append(&hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  "Invalid reference",
		Detail:   fmt.Sprintf("The %q object does not support this operation.", root),
		Subject:  traversal[1].SourceRange().Ptr(),
	})
	return "", hcl.Range{}, nil, diags
}
