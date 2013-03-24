// This is the main package for the `packer` application.
package main

import "github.com/mitchellh/packer/packer"
import "os"

type RawTemplate struct {
	Name         string
	Builders     []map[string]interface{}
	Provisioners []map[string]interface{}
	Outputs      []map[string]interface{}
}

type Builder interface {
	ConfigInterface() interface{}
	Prepare()
	Build()
}

type Build interface {
	Hook(name string)
}

func main() {
	env := packer.NewEnvironment()
	os.Exit(env.Cli(os.Args[1:]))
	/*
		file, _ := ioutil.ReadFile("example.json")

		var tpl RawTemplate
		json.Unmarshal(file, &tpl)
		fmt.Printf("%#v\n", tpl)

		builderType, ok := tpl.Builders[0]["type"].(Build)
		if !ok {
			panic("OH NOES")
		}
		fmt.Printf("TYPE: %v\n", builderType)
	*/
}
