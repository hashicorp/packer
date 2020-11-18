package plugin

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/version"
)

// Set is a plugin set. It's API is meant to be very close to what is returned
// by plugin.Server
// It can describe itself or run a single plugin using the CLI arguments.
type Set struct {
	version        string
	sdkVersion     string
	Builders       map[string]packer.Builder
	PostProcessors map[string]packer.PostProcessor
	Provisioners   map[string]packer.Provisioner
}

// Description describes a Set.
type Description struct {
	Version        string   `json:"version"`
	SDKVersion     string   `json:"sdk_version"`
	Builders       []string `json:"builders"`
	PostProcessors []string `json:"post_processors"`
	Provisioners   []string `json:"provisioners"`
}

////
// Setup
////

func New() *Set {
	return &Set{
		version:        version.String(),
		sdkVersion:     version.String(), // TODO: Set me after the split
		Builders:       map[string]packer.Builder{},
		PostProcessors: map[string]packer.PostProcessor{},
		Provisioners:   map[string]packer.Provisioner{},
	}
}

func (i *Set) RegisterBuilder(name string, builder packer.Builder) {
	if _, found := i.Builders[name]; found {
		panic(fmt.Errorf("registering duplicate %s builder", name))
	}
	i.Builders[name] = builder
}

func (i *Set) RegisterPostProcessor(name string, postProcessor packer.PostProcessor) {
	if _, found := i.PostProcessors[name]; found {
		panic(fmt.Errorf("registering duplicate %s post-processor", name))
	}
	i.PostProcessors[name] = postProcessor
}

func (i *Set) RegisterProvisioner(name string, provisioner packer.Provisioner) {
	if _, found := i.Provisioners[name]; found {
		panic(fmt.Errorf("registering duplicate %s provisioner", name))
	}
	i.Provisioners[name] = provisioner
}

// Run takes the os Args and runs a packer plugin command from it.
//  * "describe" command makes the plugin set describe itself.
//  * "start builder builder-name" starts the builder "builder-name"
//  * "start post-processor example" starts the post-processor "example"
func (i *Set) Run() error {
	args := os.Args[1:]
	return i.run(args)
}

func (i *Set) run(args []string) error {
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
	case "provisioners":
		err = server.RegisterProvisioner(i.Provisioners[name])
	default:
		err = fmt.Errorf("Unknown pluging type: %s", kind)
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

func (i *Set) description() Description {
	return Description{
		Version:        i.version,
		SDKVersion:     i.sdkVersion,
		Builders:       i.buildersDescription(),
		PostProcessors: i.postProcessorsDescription(),
		Provisioners:   i.provisionersDescription(),
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
