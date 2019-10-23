package hcl2template

import (
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type Communicator struct {
	// Type of communicator; ex: ssh
	Type string
	// Given name
	Name string

	Cfg interface{}

	HCL2Ref HCL2Ref
}

func (communicator *Communicator) Ref() CommunicatorRef {
	return CommunicatorRef{
		Type: communicator.Type,
		Name: communicator.Name,
	}
}

func (p *Parser) decodeCommunicatorConfig(block *hcl.Block) (*Communicator, hcl.Diagnostics) {

	output := &Communicator{}
	output.Type = block.Labels[0]
	output.Name = block.Labels[1]
	output.HCL2Ref.DeclRange = block.DefRange

	diags := hcl.Diagnostics{}

	communicator, found := p.CommunicatorSchemas[output.Type]
	if !found {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Unknown " + communicatorLabel + " type " + output.Type,
			Detail: "A " + communicatorLabel + " type must start with a letter and " +
				"may contain only letters, digits, underscores, and dashes.",
			Subject: &block.DefRange,
		})
		return output, diags
	}

	flatCommunicator, moreDiags := decodeDecodable(block, nil, communicator)
	diags = append(diags, moreDiags...)
	output.Cfg = flatCommunicator

	if !hclsyntax.ValidIdentifier(output.Name) {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Invalid " + communicatorLabel + " name",
			Detail: "A " + communicatorLabel + " type must start with a letter and " +
				"may contain only letters, digits, underscores, and dashes.",
			Subject: &block.DefRange,
		})
	}

	return output, diags
}

type CommunicatorRef struct {
	Type string
	Name string
}

// NoCommunicator is the zero value of CommunicatorRef, representing the
// absense of Communicator.
var NoCommunicator CommunicatorRef

func communicatorRefFromString(in string) CommunicatorRef {
	args := strings.Split(in, ".")
	if len(args) < 2 {
		return NoCommunicator
	}
	if len(args) > 2 {
		// comm.type.name
		args = args[1:]
	}
	return CommunicatorRef{
		Type: args[0],
		Name: args[1],
	}
}
