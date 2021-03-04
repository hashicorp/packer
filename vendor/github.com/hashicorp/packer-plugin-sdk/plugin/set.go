package plugin

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	pluginVersion "github.com/hashicorp/packer-plugin-sdk/version"
)

// Use this name to make the name of the plugin in the packer template match
// the multiplugin suffix, instead of requiring a second part.
// For example, calling :
//  pps.RegisterProvisioner(plugin.DEFAULT_NAME, new(CommentProvisioner))
// On a plugin named `packer-plugin-foo`, will make the `foo` provisioner available
// with your CommentProvisioner doing that. There can only be one unnamed
// plugin per plugin type.
const DEFAULT_NAME = "-packer-default-plugin-name-"

// Set is a plugin set. It's API is meant to be very close to what is returned
// by plugin.Server
// It can describe itself or run a single plugin using the CLI arguments.
type Set struct {
	version        string
	sdkVersion     string
	apiVersion     string
	Builders       map[string]packersdk.Builder
	PostProcessors map[string]packersdk.PostProcessor
	Provisioners   map[string]packersdk.Provisioner
	Datasources    map[string]packersdk.Datasource
}

// SetDescription describes a Set.
type SetDescription struct {
	Version        string   `json:"version"`
	SDKVersion     string   `json:"sdk_version"`
	APIVersion     string   `json:"api_version"`
	Builders       []string `json:"builders"`
	PostProcessors []string `json:"post_processors"`
	Provisioners   []string `json:"provisioners"`
	Datasources    []string `json:"datasources"`
}

////
// Setup
////

func NewSet() *Set {
	return &Set{
		sdkVersion:     pluginVersion.SDKVersion.String(),
		apiVersion:     "x" + APIVersionMajor + "." + APIVersionMinor,
		Builders:       map[string]packersdk.Builder{},
		PostProcessors: map[string]packersdk.PostProcessor{},
		Provisioners:   map[string]packersdk.Provisioner{},
		Datasources:    map[string]packersdk.Datasource{},
	}
}

func (i *Set) SetVersion(version *pluginVersion.PluginVersion) {
	i.version = version.String()
}

func (i *Set) RegisterBuilder(name string, builder packersdk.Builder) {
	if _, found := i.Builders[name]; found {
		panic(fmt.Errorf("registering duplicate %s builder", name))
	}
	i.Builders[name] = builder
}

func (i *Set) RegisterPostProcessor(name string, postProcessor packersdk.PostProcessor) {
	if _, found := i.PostProcessors[name]; found {
		panic(fmt.Errorf("registering duplicate %s post-processor", name))
	}
	i.PostProcessors[name] = postProcessor
}

func (i *Set) RegisterProvisioner(name string, provisioner packersdk.Provisioner) {
	if _, found := i.Provisioners[name]; found {
		panic(fmt.Errorf("registering duplicate %s provisioner", name))
	}
	i.Provisioners[name] = provisioner
}

func (i *Set) RegisterDatasource(name string, datasource packersdk.Datasource) {
	if _, found := i.Datasources[name]; found {
		panic(fmt.Errorf("registering duplicate %s datasource", name))
	}
	i.Datasources[name] = datasource
}

// Run takes the os Args and runs a packer plugin command from it.
//  * "describe" command makes the plugin set describe itself.
//  * "start builder builder-name" starts the builder "builder-name"
//  * "start post-processor example" starts the post-processor "example"
func (i *Set) Run() error {
	args := os.Args[1:]
	return i.RunCommand(args...)
}

func (i *Set) RunCommand(args ...string) error {
	if len(args) < 1 {
		return fmt.Errorf("needs at least one argument")
	}

	switch args[0] {
	case "describe":
		return i.jsonDescribe(os.Stdout)
	case "start":
		args = args[1:]
		if len(args) != 2 {
			return fmt.Errorf("start takes two arguments, for example 'start builder example-builder'. Found: %v", args)
		}
		return i.start(args[0], args[1])
	default:
		return fmt.Errorf("Unknown command: %q", args[0])
	}
}

func (i *Set) start(kind, name string) error {
	server, err := Server()
	if err != nil {
		return err
	}

	log.Printf("[TRACE] starting %s %s", kind, name)

	switch kind {
	case "builder":
		err = server.RegisterBuilder(i.Builders[name])
	case "post-processor":
		err = server.RegisterPostProcessor(i.PostProcessors[name])
	case "provisioner":
		err = server.RegisterProvisioner(i.Provisioners[name])
	case "datasource":
		err = server.RegisterDatasource(i.Datasources[name])
	default:
		err = fmt.Errorf("Unknown plugin type: %s", kind)
	}
	if err != nil {
		return err
	}
	server.Serve()
	return nil
}

////
// Describe
////

func (i *Set) description() SetDescription {
	return SetDescription{
		Version:        i.version,
		SDKVersion:     i.sdkVersion,
		APIVersion:     i.apiVersion,
		Builders:       i.buildersDescription(),
		PostProcessors: i.postProcessorsDescription(),
		Provisioners:   i.provisionersDescription(),
		Datasources:    i.datasourceDescription(),
	}
}

func (i *Set) jsonDescribe(out io.Writer) error {
	return json.NewEncoder(out).Encode(i.description())
}

func (i *Set) buildersDescription() []string {
	out := []string{}
	for key := range i.Builders {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func (i *Set) postProcessorsDescription() []string {
	out := []string{}
	for key := range i.PostProcessors {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func (i *Set) provisionersDescription() []string {
	out := []string{}
	for key := range i.Provisioners {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func (i *Set) datasourceDescription() []string {
	out := []string{}
	for key := range i.Datasources {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}
