// The packer package contains the core components of Packer.
package packer

import (
	"errors"
	"fmt"
	"os"
)

// The function type used to lookup Builder implementations.
type BuilderFunc func(name string) (Builder, error)

// The function type used to lookup Hook implementations.
type HookFunc func(name string) (Hook, error)

// The function type used to lookup PostProcessor implementations.
type PostProcessorFunc func(name string) (PostProcessor, error)

// The function type used to lookup Provisioner implementations.
type ProvisionerFunc func(name string) (Provisioner, error)

// ComponentFinder is a struct that contains the various function
// pointers necessary to look up components of Packer such as builders,
// commands, etc.
type ComponentFinder struct {
	Builder       BuilderFunc
	Hook          HookFunc
	PostProcessor PostProcessorFunc
	Provisioner   ProvisionerFunc
}

// The environment interface provides access to the configuration and
// state of a single Packer run.
//
// It allows for things such as executing CLI commands, getting the
// list of available builders, and more.
type Environment interface {
	Builder(string) (Builder, error)
	Cache() Cache
	Hook(string) (Hook, error)
	PostProcessor(string) (PostProcessor, error)
	Provisioner(string) (Provisioner, error)
	Ui() Ui
}

// An implementation of an Environment that represents the Packer core
// environment.
type coreEnvironment struct {
	cache      Cache
	components ComponentFinder
	ui         Ui
}

// This struct configures new environments.
type EnvironmentConfig struct {
	Cache      Cache
	Components ComponentFinder
	Ui         Ui
}

// DefaultEnvironmentConfig returns a default EnvironmentConfig that can
// be used to create a new enviroment with NewEnvironment with sane defaults.
func DefaultEnvironmentConfig() *EnvironmentConfig {
	config := &EnvironmentConfig{}
	config.Ui = &BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stdout,
	}

	return config
}

// This creates a new environment
func NewEnvironment(config *EnvironmentConfig) (resultEnv Environment, err error) {
	if config == nil {
		err = errors.New("config must be given to initialize environment")
		return
	}

	env := &coreEnvironment{}
	env.cache = config.Cache
	env.components = config.Components
	env.ui = config.Ui

	// We want to make sure the components have valid function pointers.
	// If a function pointer was not given, we assume that the function
	// will just return a nil component.
	if env.components.Builder == nil {
		env.components.Builder = func(string) (Builder, error) { return nil, nil }
	}

	if env.components.Hook == nil {
		env.components.Hook = func(string) (Hook, error) { return nil, nil }
	}

	if env.components.PostProcessor == nil {
		env.components.PostProcessor = func(string) (PostProcessor, error) { return nil, nil }
	}

	if env.components.Provisioner == nil {
		env.components.Provisioner = func(string) (Provisioner, error) { return nil, nil }
	}

	// The default cache is just the system temporary directory
	if env.cache == nil {
		env.cache = &FileCache{CacheDir: os.TempDir()}
	}

	resultEnv = env
	return
}

// Returns a builder of the given name that is registered with this
// environment.
func (e *coreEnvironment) Builder(name string) (b Builder, err error) {
	b, err = e.components.Builder(name)
	if err != nil {
		return
	}

	if b == nil {
		err = fmt.Errorf("No builder returned for name: %s", name)
	}

	return
}

// Returns the cache for this environment
func (e *coreEnvironment) Cache() Cache {
	return e.cache
}

// Returns a hook of the given name that is registered with this
// environment.
func (e *coreEnvironment) Hook(name string) (h Hook, err error) {
	h, err = e.components.Hook(name)
	if err != nil {
		return
	}

	if h == nil {
		err = fmt.Errorf("No hook returned for name: %s", name)
	}

	return
}

// Returns a PostProcessor for the given name that is registered with this
// environment.
func (e *coreEnvironment) PostProcessor(name string) (p PostProcessor, err error) {
	p, err = e.components.PostProcessor(name)
	if err != nil {
		return
	}

	if p == nil {
		err = fmt.Errorf("No post processor found for name: %s", name)
	}

	return
}

// Returns a provisioner for the given name that is registered with this
// environment.
func (e *coreEnvironment) Provisioner(name string) (p Provisioner, err error) {
	p, err = e.components.Provisioner(name)
	if err != nil {
		return
	}

	if p == nil {
		err = fmt.Errorf("No provisioner returned for name: %s", name)
	}

	return
}

// Returns the UI for the environment. The UI is the interface that should
// be used for all communication with the outside world.
func (e *coreEnvironment) Ui() Ui {
	return e.ui
}
