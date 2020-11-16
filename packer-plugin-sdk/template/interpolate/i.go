package interpolate

import (
	"bytes"
	"regexp"
	"strings"
	"text/template"

	"github.com/google/uuid"
)

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

	// SensitiveVariables is a list of variables to sanitize.
	SensitiveVariables []string

	// EnableEnv enables the env function
	EnableEnv bool

	// All the fields below are used for built-in functions.
	//
	// BuildName and BuildType are the name and type, respectively,
	// of the builder being used.
	//
	// TemplatePath is the path to the template that this is being
	// rendered within.
	BuildName               string
	BuildType               string
	CorePackerVersionString string
	TemplatePath            string
}

// NewContext returns an initialized empty context.
func NewContext() *Context {
	return &Context{}
}

// RenderOnce is shorthand for constructing an I and calling Render one time.
func RenderOnce(v string, ctx *Context) (string, error) {
	return (&I{Value: v}).Render(ctx)
}

// Render is shorthand for constructing an I and calling Render until all variables are rendered.
func Render(v string, ctx *Context) (rendered string, err error) {
	// Keep interpolating until all variables are done
	// Sometimes a variable can been inside another one
	for {
		rendered, err = (&I{Value: v}).Render(ctx)
		if err != nil || rendered == v {
			break
		}
		v = rendered
	}
	return
}

// Render is shorthand for constructing an I and calling Render.
// Use regex to filter variables that are not supposed to be interpolated now
func RenderRegex(v string, ctx *Context, regex string) (string, error) {
	re := regexp.MustCompile(regex)
	matches := re.FindAllStringSubmatch(v, -1)

	// Replace variables to be excluded with a unique UUID
	excluded := make(map[string]string)
	for _, value := range matches {
		id := uuid.New().String()
		excluded[id] = value[0]
		v = strings.ReplaceAll(v, value[0], id)
	}

	rendered, err := (&I{Value: v}).Render(ctx)
	if err != nil {
		return rendered, err
	}

	// Replace back by the UUID the previously excluded values
	for id, value := range excluded {
		rendered = strings.ReplaceAll(rendered, id, value)
	}

	return rendered, nil
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
func (i *I) Render(ictx *Context) (string, error) {
	tpl, err := i.template(ictx)
	if err != nil {
		return "", err
	}

	var result bytes.Buffer
	var data interface{}
	if ictx != nil {
		data = ictx.Data
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
