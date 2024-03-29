// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hashicorp/packer/fix"
)

var deprecatedOptsTemplate = template.Must(template.New("deprecatedOptsTemplate").Funcs(template.FuncMap{"StringsJoin": strings.Join}).Parse(`//<!-- Code generated by generate-fixer-deprecations; DO NOT EDIT MANUALLY -->

package config

var DeprecatedOptions = map[string][]string{
{{- range $key, $value := .DeprecatedOpts}}
	"{{$key}}": []string{"{{ StringsJoin . "\", \"" }}"},
{{- end}}
}
`))

type executeOpts struct {
	DeprecatedOpts map[string][]string
}

func main() {
	// Figure out location in directory structure
	args := flag.Args()
	if len(args) == 0 {
		// Default: process the file
		args = []string{os.Getenv("GOFILE")}
	}
	fname := args[0]

	absFilePath, err := filepath.Abs(fname)
	if err != nil {
		panic(err)
	}
	paths := strings.Split(absFilePath, "cmd"+string(os.PathSeparator)+
		"generate-fixer-deprecations"+string(os.PathSeparator)+"main.go")
	packerDir := paths[0]

	// Load all deprecated options from all active fixers
	allDeprecatedOpts := map[string][]string{}
	for _, name := range fix.FixerOrder {
		fixer, ok := fix.Fixers[name]
		if !ok {
			panic("fixer not found: " + name)
		}

		deprecated := fixer.DeprecatedOptions()
		for k, v := range deprecated {
			if allDeprecatedOpts[k] == nil {
				allDeprecatedOpts[k] = v
			} else {
				allDeprecatedOpts[k] = append(allDeprecatedOpts[k], v...)
			}
		}
	}

	deprecated_path := filepath.Join(packerDir, "packer-plugin-sdk", "template",
		"config", "deprecated_options.go")

	buf := bytes.Buffer{}

	// execute template into buffer
	deprecated := &executeOpts{DeprecatedOpts: allDeprecatedOpts}
	err = deprecatedOptsTemplate.Execute(&buf, deprecated)
	if err != nil {
		panic(err)
	}
	// we've written unformatted go code to the file. now we have to format it.
	out, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}

	outputFile, err := os.Create(deprecated_path)
	if err != nil {
		panic(err)
	}
	_, err = outputFile.Write(out)
	defer outputFile.Close()
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
}
