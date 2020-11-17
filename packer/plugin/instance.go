package plugin

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/version"
)

// Instance is a plugin instance it
type Instance struct {
	version        string
	sdkVersion     string
	Builders       map[string]packer.Builder
	PostProcessors map[string]packer.PostProcessor
	Provisioners   map[string]packer.Provisioner
}

// description is a plugin instance's description
type description struct {
	Version        string   `json:"version"`
	SDKVersion     string   `json:"sdk_version"`
	Builders       []string `json:"builders"`
	PostProcessors []string `json:"post_processors"`
	Provisioners   []string `json:"provisioners"`
}

////
// Setup
////

func New() *Instance {
	return &Instance{
		version:        version.String(),
		sdkVersion:     version.String(), // TODO: Set me after the split
		Builders:       map[string]packer.Builder{},
		PostProcessors: map[string]packer.PostProcessor{},
		Provisioners:   map[string]packer.Provisioner{},
	}
}

func (i *Instance) RegisterBuilder(name string, builder packer.Builder) {
	if _, found := i.Builders[name]; found {
		panic(fmt.Errorf("registering duplicate %s builder"))
	}
	i.Builders[name] = builder
}

func (i *Instance) RegisterPostProcessor(name string, postProcessor packer.PostProcessor) {
	if _, found := i.PostProcessors[name]; found {
		panic(fmt.Errorf("registering duplicate %s post-processor"))
	}
	i.PostProcessors[name] = postProcessor
}

func (i *Instance) RegisterProvisioner(name string, provisioner packer.Provisioner) {
	if _, found := i.Provisioners[name]; found {
		panic(fmt.Errorf("registering duplicate %s provisioner"))
	}
	i.Provisioners[name] = provisioner
}

////
// Run
////

func (i *Instance) Run() error {
	args := os.Args[1:]
	return i.run(args)
}

func (i *Instance) run(args []string) error {
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
func (i *Instance) start(kind, name string) error {
	server, err := Server()
	if err != nil {
		return err
	}

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

func (i *Instance) description() description {
	return description{
		Version:        i.version,
		SDKVersion:     i.sdkVersion,
		Builders:       i.buildersDescription(),
		PostProcessors: i.postProcessorsDescription(),
		Provisioners:   i.provisionersDescription(),
	}
}

func (i *Instance) jsonDescribe(out io.Writer) error {
	return json.NewEncoder(out).Encode(i.description())
}

func (i *Instance) buildersDescription() []string {
	out := []string{}
	for key := range i.Builders {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func (i *Instance) postProcessorsDescription() []string {
	out := []string{}
	for key := range i.PostProcessors {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

func (i *Instance) provisionersDescription() []string {
	out := []string{}
	for key := range i.Provisioners {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}
