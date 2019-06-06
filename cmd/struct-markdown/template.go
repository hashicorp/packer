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
	Fields     []Field
}

var structDocsTemplate = template.Must(template.New("structDocsTemplate").
	Funcs(template.FuncMap{
		"indent": indent,
	}).
	Parse(`<!-- Code generated from the comments of the {{ .Name }} struct in {{ .SourcePath }}; DO NOT EDIT MANUALLY -->
{{range .Fields}}
-   ` + "`" + `{{ .Name}}` + "`" + ` ({{ .Type }}) - {{ .Docs | indent 4 }}
{{- end -}}`))

func indent(spaces int, v string) string {
	pad := strings.Repeat(" ", spaces)
	return strings.Replace(v, "\n", "\n"+pad, -1)
}
