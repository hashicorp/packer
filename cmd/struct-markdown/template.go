package main

import (
	"strings"
	"text/template"
)

type Field struct {
	Name string
	Type string
	Docs string
}

type Struct struct {
	SourcePath string
	Name       string
	Filename   string
	Header     string
	Fields     []Field
}

var structDocsTemplate = template.Must(template.New("structDocsTemplate").
	Funcs(template.FuncMap{
		"indent": indent,
	}).
	Parse(`<!-- Code generated from the comments of the {{ .Name }} struct in {{ .SourcePath }}; DO NOT EDIT MANUALLY -->
{{ if .Header  }}
{{ .Header }}
{{ end -}}
{{ range .Fields }}
- ` + "`" + `{{ .Name}}` + "`" + ` ({{ .Type }}) - {{ .Docs | indent 2 }}
{{ end }}`))

func indent(spaces int, v string) string {
	pad := strings.Repeat(" ", spaces)
	return strings.TrimSpace(strings.Replace(v, "\n", "\n"+pad, -1))
}
