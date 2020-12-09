package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/tools/imports"
)

var (
	tmpl = template.Must(template.New("var").Parse(`
	// CODE GENERATED. DO NOT EDIT
	package {{.PkgName }}
	var (
		{{ .Name }} = ` + "`" + `{{.Value}}` + "`" + `
	)

	`))
)

type vars struct {
	PkgName string
	Name    string
	Value   string
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s file varname [output]", os.Args[0])
	}
	fname := os.Args[1]
	targetVar := os.Args[2]
	pkg := os.Getenv("GOPACKAGE")
	absFilePath, err := filepath.Abs(fname)

	targetFName := strings.ToLower(targetVar) + ".go"
	if len(os.Args) > 3 {
		targetFName = os.Args[3]
	}
	log.Println(absFilePath, "=>", targetFName)
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(absFilePath); err != nil {
		os.Remove(absFilePath)
	}
	buff := bytes.Buffer{}
	err = tmpl.Execute(&buff, vars{
		Name:    targetVar,
		Value:   string(b),
		PkgName: pkg,
	})
	if err != nil {
		log.Fatal(err)
	}

	data, err := imports.Process(targetFName, buff.Bytes(), nil)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(targetFName)
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}
