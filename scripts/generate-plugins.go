// Generate Plugins is a small program that updates the lists of plugins in
// command/plugin.go so they will be compiled into the main packer binary.
//
// See https://github.com/mitchellh/packer/pull/2608 for details.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const target = "command/plugin.go"

func main() {
	wd, _ := os.Getwd()
	if filepath.Base(wd) != "packer" {
		log.Fatalf("This program must be invoked in the packer project root; in %s", wd)
	}

	// Collect all of the data we need about plugins we have in the project
	builders, err := discoverBuilders()
	if err != nil {
		log.Fatalf("Failed to discover builders: %s", err)
	}

	provisioners, err := discoverProvisioners()
	if err != nil {
		log.Fatalf("Failed to discover provisioners: %s", err)
	}

	postProcessors, err := discoverPostProcessors()
	if err != nil {
		log.Fatalf("Failed to discover post processors: %s", err)
	}

	// Do some simple code generation and templating
	output := source
	output = strings.Replace(output, "IMPORTS", makeImports(builders, provisioners, postProcessors), 1)
	output = strings.Replace(output, "BUILDERS", makeMap("Builders", "Builder", builders), 1)
	output = strings.Replace(output, "PROVISIONERS", makeMap("Provisioners", "Provisioner", provisioners), 1)
	output = strings.Replace(output, "POSTPROCESSORS", makeMap("PostProcessors", "PostProcessor", postProcessors), 1)

	// TODO sort the lists of plugins so we are not subjected to random OS ordering of the plugin lists
	// TODO format the file

	// Write our generated code to the command/plugin.go file
	file, err := os.Create(target)
	defer file.Close()
	if err != nil {
		log.Fatalf("Failed to open %s for writing: %s", target, err)
	}

	_, err = file.WriteString(output)
	if err != nil {
		log.Fatalf("Failed writing to %s: %s", target, err)
	}

	log.Printf("Generated %s", target)
}

type plugin struct {
	Package    string // This plugin's package name (iso)
	PluginName string // Name of plugin (vmware-iso)
	TypeName   string // Type of plugin (builder)
	Path       string // Path relative to packer root (builder/vmware/iso)
	ImportName string // PluginName+TypeName (vmwareisobuilder)
}

// makeMap creates a map named Name with type packer.Name that looks something
// like this:
//
// var Builders = map[string]packer.Builder{
// 	"amazon-chroot":   new(chroot.Builder),
// 	"amazon-ebs":      new(ebs.Builder),
// 	"amazon-instance": new(instance.Builder),
func makeMap(varName, varType string, items []plugin) string {
	output := ""

	output += fmt.Sprintf("var %s = map[string]packer.%s{\n", varName, varType)
	for _, item := range items {
		output += fmt.Sprintf("\t\"%s\":   new(%s.%s),\n", item.PluginName, item.ImportName, item.TypeName)
	}
	output += "}\n"
	return output
}

func makeImports(builders, provisioners, postProcessors []plugin) string {
	plugins := []string{}

	for _, builder := range builders {
		plugins = append(plugins, fmt.Sprintf("\t%s \"github.com/mitchellh/packer/%s\"\n", builder.ImportName, filepath.ToSlash(builder.Path)))
	}

	for _, provisioner := range provisioners {
		plugins = append(plugins, fmt.Sprintf("\t%s \"github.com/mitchellh/packer/%s\"\n", provisioner.ImportName, filepath.ToSlash(provisioner.Path)))
	}

	for _, postProcessor := range postProcessors {
		plugins = append(plugins, fmt.Sprintf("\t%s \"github.com/mitchellh/packer/%s\"\n", postProcessor.ImportName, filepath.ToSlash(postProcessor.Path)))
	}

	// Make things pretty
	sort.Strings(plugins)

	return strings.Join(plugins, "")
}

// listDirectories recursively lists directories under the specified path
func listDirectories(path string) ([]string, error) {
	names := []string{}
	items, err := ioutil.ReadDir(path)
	if err != nil {
		return names, err
	}

	for _, item := range items {
		// We only want directories
		if item.IsDir() {
			currentDir := filepath.Join(path, item.Name())
			names = append(names, currentDir)

			// Do some recursion
			subNames, err := listDirectories(currentDir)
			if err == nil {
				names = append(names, subNames...)
			}
		}
	}

	return names, nil
}

// deriveName determines the name of the plugin (what you'll see in a packer
// template) based on the filesystem path. We use two rules:
//
// Start with                     -> builder/virtualbox/iso
//
// 1. Strip the root              -> virtualbox/iso
// 2. Switch slash / to dash -    -> virtualbox-iso
func deriveName(root, full string) string {
	short, _ := filepath.Rel(root, full)
	bits := strings.Split(short, string(os.PathSeparator))
	return strings.Join(bits, "-")
}

// deriveImport will build a unique import identifier based on packageName and
// the result of deriveName()
//
// This will be something like    -> virtualboxisobuilder
//
// Which is long, but deterministic and unique.
func deriveImport(typeName, derivedName string) string {
	return strings.Replace(derivedName, "-", "", -1) + strings.ToLower(typeName)
}

// discoverTypesInPath searches for types of typeID in path and returns a list
// of plugins it finds.
func discoverTypesInPath(path, typeID string) ([]plugin, error) {
	postProcessors := []plugin{}

	dirs, err := listDirectories(path)
	if err != nil {
		return postProcessors, err
	}

	for _, dir := range dirs {
		fset := token.NewFileSet()
		goPackages, err := parser.ParseDir(fset, dir, nil, parser.AllErrors)
		if err != nil {
			return postProcessors, fmt.Errorf("Failed parsing directory %s: %s", dir, err)
		}

		for _, goPackage := range goPackages {
			ast.PackageExports(goPackage)
			ast.Inspect(goPackage, func(n ast.Node) bool {
				switch x := n.(type) {
				case *ast.TypeSpec:
					if x.Name.Name == typeID {
						derivedName := deriveName(path, dir)
						postProcessors = append(postProcessors, plugin{
							Package:    goPackage.Name,
							PluginName: derivedName,
							ImportName: deriveImport(x.Name.Name, derivedName),
							TypeName:   x.Name.Name,
							Path:       dir,
						})
						// The AST stops parsing when we return false. Once we
						// find the symbol we want we can stop parsing.

						// DEBUG:
						// fmt.Printf("package %#v\n", goPackage)
						return false
					}
				}
				return true
			})
		}
	}

	return postProcessors, nil
}

func discoverBuilders() ([]plugin, error) {
	path := "./builder"
	typeID := "Builder"
	return discoverTypesInPath(path, typeID)
}

func discoverProvisioners() ([]plugin, error) {
	path := "./provisioner"
	typeID := "Provisioner"
	return discoverTypesInPath(path, typeID)
}

func discoverPostProcessors() ([]plugin, error) {
	path := "./post-processor"
	typeID := "PostProcessor"
	return discoverTypesInPath(path, typeID)
}

const source = `//
// This file is automatically generated by scripts/generate-plugins.go -- Do not edit!
//

package command

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/packer/plugin"

IMPORTS
)

type PluginCommand struct {
	Meta
}

BUILDERS

PROVISIONERS

POSTPROCESSORS

var pluginRegexp = regexp.MustCompile("packer-(builder|post-processor|provisioner)-(.+)")

func (c *PluginCommand) Run(args []string) int {
	// This is an internal call (users should not call this directly) so we're
	// not going to do much input validation. If there's a problem we'll often
	// just crash. Error handling should be added to facilitate debugging.
	log.Printf("args: %#v", args)
	if len(args) != 1 {
		c.Ui.Error("Wrong number of args")
		return 1
	}

	// Plugin will match something like "packer-builder-amazon-ebs"
	parts := pluginRegexp.FindStringSubmatch(args[0])
	if len(parts) != 3 {
		c.Ui.Error(fmt.Sprintf("Error parsing plugin argument [DEBUG]: %#v", parts))
		return 1
	}
	pluginType := parts[1] // capture group 1 (builder|post-processor|provisioner)
	pluginName := parts[2] // capture group 2 (.+)

	server, err := plugin.Server()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error starting plugin server: %s", err))
		return 1
	}

	switch pluginType {
	case "builder":
		builder, found := Builders[pluginName]
		if !found {
			c.Ui.Error(fmt.Sprintf("Could not load builder: %s", pluginName))
			return 1
		}
		server.RegisterBuilder(builder)
	case "provisioner":
		provisioner, found := Provisioners[pluginName]
		if !found {
			c.Ui.Error(fmt.Sprintf("Could not load provisioner: %s", pluginName))
			return 1
		}
		server.RegisterProvisioner(provisioner)
	case "post-processor":
		postProcessor, found := PostProcessors[pluginName]
		if !found {
			c.Ui.Error(fmt.Sprintf("Could not load post-processor: %s", pluginName))
			return 1
		}
		server.RegisterPostProcessor(postProcessor)
	}

	server.Serve()

	return 0
}

func (*PluginCommand) Help() string {
	helpText := ` + "`" + `
Usage: packer plugin PLUGIN

  Runs an internally-compiled version of a plugin from the packer binary.

  NOTE: this is an internal command and you should not call it yourself.
` + "`" + `

	return strings.TrimSpace(helpText)
}

func (c *PluginCommand) Synopsis() string {
	return "internal plugin command"
}
`
