package common

import (
	"bytes"
	"fmt"
	"strconv"
	"text/template"
	"time"
)

type Template struct {
	UserData map[string]string

	root *template.Template
}

func NewTemplate() (*Template, error) {
	result := &Template{
		UserData: make(map[string]string),
	}

	result.root = template.New("configTemplateRoot")
	result.root.Funcs(template.FuncMap{
		"timestamp": templateTimestamp,
		"user":      result.templateUser,
	})

	return result, nil
}

func (t *Template) Process(s string, data interface{}) (string, error) {
	tpl, err := t.root.New("tpl").Parse(s)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err := tpl.Execute(buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// User is the function exposed as "user" within the templates and
// looks up user variables.
func (t *Template) templateUser(n string) (string, error) {
	result, ok := t.UserData[n]
	if !ok {
		return "", fmt.Errorf("uknown user var: %s", n)
	}

	return result, nil
}

func templateTimestamp() string {
	return strconv.FormatInt(time.Now().UTC().Unix(), 10)
}
