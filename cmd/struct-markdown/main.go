package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/camelcase"
	"github.com/fatih/structtag"
)

func main() {
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
	paths := strings.SplitAfter(absFilePath, "packer"+string(os.PathSeparator))
	packerDir := paths[0]
	builderName, _ := filepath.Split(paths[1])
	builderName = strings.Trim(builderName, string(os.PathSeparator))

	b, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Printf("ReadFile: %+v", err)
		os.Exit(1)
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fname, b, parser.ParseComments)
	if err != nil {
		fmt.Printf("ParseFile: %+v", err)
		os.Exit(1)
	}

	for _, decl := range f.Decls {
		typeDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		typeSpec, ok := typeDecl.Specs[0].(*ast.TypeSpec)
		if !ok {
			continue
		}
		structDecl, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			continue
		}

		fields := structDecl.Fields.List
		sourcePath := filepath.ToSlash(paths[1])
		header := Struct{
			SourcePath: sourcePath,
			Name:       typeSpec.Name.Name,
			Filename:   typeSpec.Name.Name + ".mdx",
			Header:     typeDecl.Doc.Text(),
		}
		required := Struct{
			SourcePath: sourcePath,
			Name:       typeSpec.Name.Name,
			Filename:   typeSpec.Name.Name + "-required.mdx",
		}
		notRequired := Struct{
			SourcePath: sourcePath,
			Name:       typeSpec.Name.Name,
			Filename:   typeSpec.Name.Name + "-not-required.mdx",
		}

		for _, field := range fields {
			if len(field.Names) == 0 || field.Tag == nil {
				continue
			}
			tag := field.Tag.Value[1:]
			tag = tag[:len(tag)-1]
			tags, err := structtag.Parse(tag)
			if err != nil {
				fmt.Printf("structtag.Parse(%s): err: %v", field.Tag.Value, err)
				os.Exit(1)
			}

			mstr, err := tags.Get("mapstructure")
			if err != nil {
				continue
			}
			name := mstr.Name

			if name == "" {
				continue
			}

			var docs string
			if field.Doc != nil {
				docs = field.Doc.Text()
			} else {
				docs = strings.Join(camelcase.Split(field.Names[0].Name), " ")
			}

			if strings.Contains(docs, "TODO") {
				continue
			}
			fieldType := string(b[field.Type.Pos()-1 : field.Type.End()-1])
			fieldType = strings.ReplaceAll(fieldType, "*", `\*`)
			switch fieldType {
			case "time.Duration":
				fieldType = `duration string | ex: "1h5m2s"`
			case "config.Trilean":
				fieldType = `boolean`
			case "hcl2template.NameValues":
				fieldType = `[]{name string, value string}`
			}

			field := Field{
				Name: name,
				Type: fieldType,
				Docs: docs,
			}
			if req, err := tags.Get("required"); err == nil && req.Value() == "true" {
				required.Fields = append(required.Fields, field)
			} else {
				notRequired.Fields = append(notRequired.Fields, field)
			}
		}

		dir := filepath.Join(packerDir, "website", "pages", "partials", builderName)
		os.MkdirAll(dir, 0755)

		for _, str := range []Struct{header, required, notRequired} {
			if len(str.Fields) == 0 && len(str.Header) == 0 {
				continue
			}
			outputPath := filepath.Join(dir, str.Filename)

			outputFile, err := os.Create(outputPath)
			if err != nil {
				panic(err)
			}
			defer outputFile.Close()

			err = structDocsTemplate.Execute(outputFile, str)
			if err != nil {
				fmt.Printf("%v", err)
				os.Exit(1)
			}
		}
	}

}
