package common

import (
	"bytes"
	"fmt"
	"strconv"
	"text/template"
	"time"
)

// Template processes string data as a text/template with some common
// elements and functions available. Plugin creators should process as
// many fields as possible through this.
type Template struct {
	UserVars map[string]string

	root *template.Template
	i    int
}

// NewTemplate creates a new template processor.
func NewTemplate() (*Template, error) {
	result := &Template{
		UserVars: make(map[string]string),
	}

	result.root = template.New("configTemplateRoot")
	result.root.Funcs(template.FuncMap{
		"timestamp": templateTimestamp,
		"user":      result.templateUser,
	})

	return result, nil
}

// Process processes a single string, compiling and executing the template.
func (t *Template) Process(s string, data interface{}) (string, error) {
	tpl, err := t.root.New(t.nextTemplateName()).Parse(s)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Validate the template.
func (t *Template) Validate(s string) error {
	root, err := t.root.Clone()
	if err != nil {
		return err
	}

	_, err = root.New("template").Parse(s)
	return err
}

func (t *Template) nextTemplateName() string {
	name := fmt.Sprintf("tpl%d", t.i)
	t.i++
	return name
}

// User is the function exposed as "user" within the templates and
// looks up user variables.
func (t *Template) templateUser(n string) (string, error) {
	result, ok := t.UserVars[n]
	if !ok {
		return "", fmt.Errorf("uknown user var: %s", n)
	}

	return result, nil
}

func templateTimestamp() string {
	return strconv.FormatInt(time.Now().UTC().Unix(), 10)
}
