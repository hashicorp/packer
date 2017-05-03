package interpolate

import (
	"bytes"
	"strings"
	"text/template"
)

const maxInterpolationRounds = 5

// Context is the context that an interpolation is done in. This defines
// things such as available variables.
type Context struct {
	// Data is the data for the template that is available
	Data interface{}

	// Funcs are extra functions available in the template
	Funcs map[string]interface{}

	// UserVariables is the mapping of user variables that the
	// "user" function reads from.
	UserVariables map[string]string

	// EnableEnv enables the env function
	EnableEnv bool

	// All the fields below are used for built-in functions.
	//
	// BuildName and BuildType are the name and type, respectively,
	// of the builder being used.
	//
	// TemplatePath is the path to the template that this is being
	// rendered within.
	BuildName    string
	BuildType    string
	TemplatePath string
}

// Render is shorthand for constructing an I and calling Render.
// Variable interpolation is done through several rounds in order to cover
// nested variables definitions (e.g. variables whose content is referencing
// another variable). Termination of substitution rounds is either max const or
// no delta between rounds.
func Render(v string, ctx *Context) (result string, err error) {
	for count := 0; count < maxInterpolationRounds; count++ {
		i, err := (&I{Value: v}).Render(ctx)
		if err != nil {
			return i, err
		}
		if strings.EqualFold(i, v) {
			result = i
			break
		} else {
			v = i
		}
	}
	return result, err
}

// Validate is shorthand for constructing an I and calling Validate.
func Validate(v string, ctx *Context) error {
	return (&I{Value: v}).Validate(ctx)
}

// I stands for "interpolation" and is the main interpolation struct
// in order to render values.
type I struct {
	Value string
}

// Render renders the interpolation with the given context.
func (i *I) Render(ctx *Context) (string, error) {
	tpl, err := i.template(ctx)
	if err != nil {
		return "", err
	}

	var result bytes.Buffer
	var data interface{}
	if ctx != nil {
		data = ctx.Data
	}
	if err := tpl.Execute(&result, data); err != nil {
		return "", err
	}

	return result.String(), nil
}

// Validate validates that the template is syntactically valid.
func (i *I) Validate(ctx *Context) error {
	_, err := i.template(ctx)
	return err
}

func (i *I) template(ctx *Context) (*template.Template, error) {
	return template.New("root").Funcs(Funcs(ctx)).Parse(i.Value)
}
