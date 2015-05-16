package interpolate

import (
	"bytes"
	"text/template"
)

// Context is the context that an interpolation is done in. This defines
// things such as available variables.
type Context struct {
	DisableEnv bool
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
	data := map[string]interface{}{}
	if err := tpl.Execute(&result, data); err != nil {
		return "", err
	}

	return result.String(), nil
}

func (i *I) template(ctx *Context) (*template.Template, error) {
	return template.New("root").Funcs(Funcs(ctx)).Parse(i.Value)
}
