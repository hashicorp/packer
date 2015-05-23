package interpolate

import (
	"bytes"
	"text/template"
)

// Context is the context that an interpolation is done in. This defines
// things such as available variables.
type Context struct {
	// Data is the data for the template that is available
	Data interface{}

	// UserVariables is the mapping of user variables that the
	// "user" function reads from.
	UserVariables map[string]string

	// EnableEnv enables the env function
	EnableEnv bool
}

// Render is shorthand for constructing an I and calling Render.
func Render(v string, ctx *Context) (string, error) {
	return (&I{Value: v}).Render(ctx)
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

func (i *I) template(ctx *Context) (*template.Template, error) {
	return template.New("root").Funcs(Funcs(ctx)).Parse(i.Value)
}
