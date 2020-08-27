package hcl2template

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/ext/typeexpr"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

// Local represents a single entry from a "locals" block in a module or file.
// The "locals" block itself is not represented, because it serves only to
// provide context for us to interpret its contents.
type LocalBlock struct {
	Name string
	Expr hcl.Expression
}

type Variable struct {
	// CmdValue, VarfileValue, EnvValue, DefaultValue are possible values of
	// the variable; The first value set from these will be the one used. If
	// none is set; an error will be returned if a user tries to use the
	// Variable.
	CmdValue     cty.Value
	VarfileValue cty.Value
	EnvValue     cty.Value
	DefaultValue cty.Value

	// Cty Type of the variable. If the default value or a collected value is
	// not of this type nor can be converted to this type an error diagnostic
	// will show up. This allows us to assume that values are valid later in
	// code.
	//
	// When a default value - and no type - is passed in the variable
	// declaration, the type of the default variable will be used. This will
	// allow to ensure that users set this variable correctly.
	Type cty.Type
	// Common name of the variable
	Name string
	// Description of the variable
	Description string
	// When Sensitive is set to true Packer will try it best to hide/obfuscate
	// the variable from the output stream. By replacing the text.
	Sensitive bool

	Range hcl.Range
}

func (v *Variable) GoString() string {
	return fmt.Sprintf("{Type:%s,CmdValue:%s,VarfileValue:%s,EnvValue:%s,DefaultValue:%s}",
		v.Type.GoString(),
		PrintableCtyValue(v.CmdValue),
		PrintableCtyValue(v.VarfileValue),
		PrintableCtyValue(v.EnvValue),
		PrintableCtyValue(v.DefaultValue),
	)
}

func (v *Variable) Value() (cty.Value, *hcl.Diagnostic) {
	for _, value := range []cty.Value{
		v.CmdValue,
		v.VarfileValue,
		v.EnvValue,
		v.DefaultValue,
	} {
		if value != cty.NilVal {
			return value, nil
		}
	}

	return cty.UnknownVal(v.Type), &hcl.Diagnostic{
		Severity: hcl.DiagError,
		Summary:  fmt.Sprintf("Unset variable %q", v.Name),
		Detail: "A used variable must be set or have a default value; see " +
			"https://packer.io/docs/configuration/from-1.5/syntax for " +
			"details.",
		Context: v.Range.Ptr(),
	}
}

type Variables map[string]*Variable

func (variables Variables) Values() (map[string]cty.Value, hcl.Diagnostics) {
	res := map[string]cty.Value{}
	var diags hcl.Diagnostics
	for k, v := range variables {
		value, diag := v.Value()
		if diag != nil {
			diags = append(diags, diag)
		}
		res[k] = value
	}
	return res, diags
}

// decodeVariable decodes a variable key and value into Variables
func (variables *Variables) decodeVariable(key string, attr *hcl.Attribute, ectx *hcl.EvalContext) hcl.Diagnostics {
	var diags hcl.Diagnostics

	if (*variables) == nil {
		(*variables) = Variables{}
	}

	if _, found := (*variables)[key]; found {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  "Duplicate variable",
			Detail:   "Duplicate " + key + " variable definition found.",
			Subject:  attr.NameRange.Ptr(),
		})
		return diags
	}

	value, moreDiags := attr.Expr.Value(ectx)
	diags = append(diags, moreDiags...)
	if moreDiags.HasErrors() {
		return diags
	}

	(*variables)[key] = &Variable{
		Name:         key,
		DefaultValue: value,
		Type:         value.Type(),
		Range:        attr.Range,
	}

	return diags
}

// decodeVariableBlock decodes a "variables" section the way packer 1 used to
func (variables *Variables) decodeVariableBlock(block *hcl.Block, ectx *hcl.EvalContext) hcl.Diagnostics {
	if (*variables) == nil {
		(*variables) = Variables{}
	}

	if _, found := (*variables)[block.Labels[0]]; found {

		return []*hcl.Diagnostic{{
			Severity: hcl.DiagError,
			Summary:  "Duplicate variable",
			Detail:   "Duplicate " + block.Labels[0] + " variable definition found.",
			Context:  block.DefRange.Ptr(),
		}}
	}

	var b struct {
		Description string   `hcl:"description,optional"`
		Sensitive   bool     `hcl:"sensitive,optional"`
		Rest        hcl.Body `hcl:",remain"`
	}
	diags := gohcl.DecodeBody(block.Body, nil, &b)

	if diags.HasErrors() {
		return diags
	}

	name := block.Labels[0]

	res := &Variable{
		Name:        name,
		Description: b.Description,
		Sensitive:   b.Sensitive,
		Range:       block.DefRange,
	}

	attrs, moreDiags := b.Rest.JustAttributes()
	diags = append(diags, moreDiags...)

	if t, ok := attrs["type"]; ok {
		delete(attrs, "type")
		tp, moreDiags := typeexpr.Type(t.Expr)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			return diags
		}

		res.Type = tp
	}

	if def, ok := attrs["default"]; ok {
		delete(attrs, "default")
		defaultValue, moreDiags := def.Expr.Value(ectx)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			return diags
		}

		if res.Type != cty.NilType {
			var err error
			defaultValue, err = convert.Convert(defaultValue, res.Type)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid default value for variable",
					Detail:   fmt.Sprintf("This default value is not compatible with the variable's type constraint: %s.", err),
					Subject:  def.Expr.Range().Ptr(),
				})
				defaultValue = cty.DynamicVal
			}
		}

		res.DefaultValue = defaultValue

		// It's possible no type attribute was assigned so lets make sure we
		// have a valid type otherwise there could be issues parsing the value.
		if res.Type == cty.NilType {
			res.Type = res.DefaultValue.Type()
		}
	}
	if len(attrs) > 0 {
		keys := []string{}
		for k := range attrs {
			keys = append(keys, k)
		}
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagWarning,
			Summary:  "Unknown keys",
			Detail:   fmt.Sprintf("unknown variable setting(s): %s", keys),
			Context:  block.DefRange.Ptr(),
		})
	}

	(*variables)[name] = res

	return diags
}

// Prefix your environment variables with VarEnvPrefix so that Packer can see
// them.
const VarEnvPrefix = "PKR_VAR_"

func (cfg *PackerConfig) collectInputVariableValues(env []string, files []*hcl.File, argv map[string]string) hcl.Diagnostics {
	var diags hcl.Diagnostics
	variables := cfg.InputVariables

	for _, raw := range env {
		if !strings.HasPrefix(raw, VarEnvPrefix) {
			continue
		}
		raw = raw[len(VarEnvPrefix):] // trim the prefix

		eq := strings.Index(raw, "=")
		if eq == -1 {
			// Seems invalid, so we'll ignore it.
			continue
		}

		name := raw[:eq]
		value := raw[eq+1:]

		variable, found := variables[name]
		if !found {
			// this variable was not defined in the hcl files, let's skip it !
			continue
		}

		fakeFilename := fmt.Sprintf("<value for var.%s from env>", name)
		expr, moreDiags := expressionFromVariableDefinition(fakeFilename, value, variable.Type)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			continue
		}

		val, valDiags := expr.Value(nil)
		diags = append(diags, valDiags...)
		if variable.Type != cty.NilType {
			var err error
			val, err = convert.Convert(val, variable.Type)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid value for variable",
					Detail:   fmt.Sprintf("The value for %s is not compatible with the variable's type constraint: %s.", name, err),
					Subject:  expr.Range().Ptr(),
				})
				val = cty.DynamicVal
			}
		}
		variable.EnvValue = val
	}

	// files will contain files found in the folder then files passed as
	// arguments.
	for _, file := range files {
		// Before we do our real decode, we'll probe to see if there are any
		// blocks of type "variable" in this body, since it's a common mistake
		// for new users to put variable declarations in pkrvars rather than
		// variable value definitions, and otherwise our error message for that
		// case is not so helpful.
		{
			content, _, _ := file.Body.PartialContent(&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type:       "variable",
						LabelNames: []string{"name"},
					},
				},
			})
			for _, block := range content.Blocks {
				name := block.Labels[0]
				diags = diags.Append(&hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Variable declaration in a .pkrvar file",
					Detail: fmt.Sprintf("A .pkrvar file is used to assign "+
						"values to variables that have already been declared "+
						"in .pkr files, not to declare new variables. To "+
						"declare variable %q, place this block in one of your"+
						" .pkr files, such as variables.pkr.hcl\n\nTo set a "+
						"value for this variable in %s, use the definition "+
						"syntax instead:\n    %s = <value>",
						name, block.TypeRange.Filename, name),
					Subject: &block.TypeRange,
				})
			}
			if diags.HasErrors() {
				// If we already found problems then JustAttributes below will find
				// the same problems with less-helpful messages, so we'll bail for
				// now to let the user focus on the immediate problem.
				return diags
			}
		}

		attrs, moreDiags := file.Body.JustAttributes()
		diags = append(diags, moreDiags...)

		for name, attr := range attrs {
			variable, found := variables[name]
			if !found {
				sev := hcl.DiagWarning
				if cfg.ValidationOptions.Strict {
					sev = hcl.DiagError
				}
				diags = append(diags, &hcl.Diagnostic{
					Severity: sev,
					Summary:  "Undefined variable",
					Detail: fmt.Sprintf("A %q variable was set but was "+
						"not found in known variables. To declare "+
						"variable %q, place this block in one of your "+
						".pkr files, such as variables.pkr.hcl",
						name, name),
					Context: attr.Range.Ptr(),
				})
				continue
			}

			val, moreDiags := attr.Expr.Value(nil)
			diags = append(diags, moreDiags...)

			if variable.Type != cty.NilType {
				var err error
				val, err = convert.Convert(val, variable.Type)
				if err != nil {
					diags = append(diags, &hcl.Diagnostic{
						Severity: hcl.DiagError,
						Summary:  "Invalid value for variable",
						Detail:   fmt.Sprintf("The value for %s is not compatible with the variable's type constraint: %s.", name, err),
						Subject:  attr.Expr.Range().Ptr(),
					})
					val = cty.DynamicVal
				}
			}

			variable.VarfileValue = val
		}
	}

	// Finally we process values given explicitly on the command line.
	for name, value := range argv {
		variable, found := variables[name]
		if !found {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Undefined -var variable",
				Detail: fmt.Sprintf("A %q variable was passed in the command "+
					"line but was not found in known variables. "+
					"To declare variable %q, place this block in one of your"+
					" .pkr files, such as variables.pkr.hcl",
					name, name),
			})
			continue
		}

		fakeFilename := fmt.Sprintf("<value for var.%s from arguments>", name)
		expr, moreDiags := expressionFromVariableDefinition(fakeFilename, value, variable.Type)
		diags = append(diags, moreDiags...)
		if moreDiags.HasErrors() {
			continue
		}

		val, valDiags := expr.Value(nil)
		diags = append(diags, valDiags...)

		if variable.Type != cty.NilType {
			var err error
			val, err = convert.Convert(val, variable.Type)
			if err != nil {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagError,
					Summary:  "Invalid argument value for -var variable",
					Detail:   fmt.Sprintf("The received arg value for %s is not compatible with the variable's type constraint: %s.", name, err),
					Subject:  expr.Range().Ptr(),
				})
				val = cty.DynamicVal
			}
		}

		variable.CmdValue = val
	}

	return diags
}

// expressionFromVariableDefinition creates an hclsyntax.Expression that is capable of evaluating the specified value for a given cty.Type.
// The specified filename is to identify the source of where value originated from in the diagnostics report, if there is an error.
func expressionFromVariableDefinition(filename string, value string, variableType cty.Type) (hclsyntax.Expression, hcl.Diagnostics) {
	switch variableType {
	case cty.String, cty.Number:
		return &hclsyntax.LiteralValueExpr{Val: cty.StringVal(value)}, nil
	default:
		return hclsyntax.ParseExpression([]byte(value), filename, hcl.Pos{Line: 1, Column: 1})
	}
}
