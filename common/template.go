package common

import (
	"bytes"
	"strconv"
	"text/template"
	"time"
)

type Template struct {
	root *template.Template
}

func NewTemplate() (*Template, error) {
	root := template.New("configTemplateRoot")
	root.Funcs(template.FuncMap{
		"timestamp": templateTimestamp,
	})

	result := &Template{
		root: root,
	}

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

func templateTimestamp() string {
	return strconv.FormatInt(time.Now().UTC().Unix(), 10)
}
