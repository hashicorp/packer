package packer

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func init() {
	// Disable log output for tests
	log.SetOutput(ioutil.Discard)
}

func testComponentFinder() *ComponentFinder {
	builderFactory := func(n string) (Builder, error) { return new(MockBuilder), nil }
	ppFactory := func(n string) (PostProcessor, error) { return new(TestPostProcessor), nil }
	provFactory := func(n string) (Provisioner, error) { return new(MockProvisioner), nil }
	return &ComponentFinder{
		Builder:       builderFactory,
		PostProcessor: ppFactory,
		Provisioner:   provFactory,
	}
}

func testEnvironment() Environment {
	config := DefaultEnvironmentConfig()
	config.Ui = &BasicUi{
		Reader:      new(bytes.Buffer),
		Writer:      new(bytes.Buffer),
		ErrorWriter: new(bytes.Buffer),
	}

	env, err := NewEnvironment(config)
	if err != nil {
		panic(err)
	}

	return env
}

func TestEnvironment_DefaultConfig_Ui(t *testing.T) {
	config := DefaultEnvironmentConfig()
	if config.Ui == nil {
		t.Fatal("config.Ui should not be nil")
	}

	rwUi, ok := config.Ui.(*BasicUi)
	if !ok {
		t.Fatal("default UI should be BasicUi")
	}
	if rwUi.Writer != os.Stdout {
		t.Fatal("default UI should go to stdout")
	}
	if rwUi.Reader != os.Stdin {
		t.Fatal("default UI reader should go to stdin")
	}
}

func TestNewEnvironment_NoConfig(t *testing.T) {
	env, err := NewEnvironment(nil)
	if env != nil {
		t.Fatal("env should be nil")
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestEnvironment_NilComponents(t *testing.T) {
	config := DefaultEnvironmentConfig()
	config.Components = *new(ComponentFinder)

	env, err := NewEnvironment(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// All of these should not cause panics... so we don't assert
	// anything but if there is a panic in the test then yeah, something
	// went wrong.
	env.Builder("foo")
	env.Hook("foo")
	env.PostProcessor("foo")
	env.Provisioner("foo")
}

func TestEnvironment_Builder(t *testing.T) {
	builder := &MockBuilder{}
	builders := make(map[string]Builder)
	builders["foo"] = builder

	config := DefaultEnvironmentConfig()
	config.Components.Builder = func(n string) (Builder, error) { return builders[n], nil }

	env, _ := NewEnvironment(config)
	returnedBuilder, err := env.Builder("foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if returnedBuilder != builder {
		t.Fatalf("bad: %#v", returnedBuilder)
	}
}

func TestEnvironment_Builder_NilError(t *testing.T) {
	config := DefaultEnvironmentConfig()
	config.Components.Builder = func(n string) (Builder, error) { return nil, nil }

	env, _ := NewEnvironment(config)
	returnedBuilder, err := env.Builder("foo")
	if err == nil {
		t.Fatal("should have error")
	}
	if returnedBuilder != nil {
		t.Fatalf("bad: %#v", returnedBuilder)
	}
}

func TestEnvironment_Builder_Error(t *testing.T) {
	config := DefaultEnvironmentConfig()
	config.Components.Builder = func(n string) (Builder, error) { return nil, errors.New("foo") }

	env, _ := NewEnvironment(config)
	returnedBuilder, err := env.Builder("foo")
	if err == nil {
		t.Fatal("should have error")
	}
	if err.Error() != "foo" {
		t.Fatalf("bad err: %s", err)
	}
	if returnedBuilder != nil {
		t.Fatalf("should be nil: %#v", returnedBuilder)
	}
}

func TestEnvironment_Cache(t *testing.T) {
	config := DefaultEnvironmentConfig()
	env, _ := NewEnvironment(config)
	if env.Cache() == nil {
		t.Fatal("cache should not be nil")
	}
}

func TestEnvironment_Hook(t *testing.T) {
	hook := &MockHook{}
	hooks := make(map[string]Hook)
	hooks["foo"] = hook

	config := DefaultEnvironmentConfig()
	config.Components.Hook = func(n string) (Hook, error) { return hooks[n], nil }

	env, _ := NewEnvironment(config)
	returned, err := env.Hook("foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if returned != hook {
		t.Fatalf("bad: %#v", returned)
	}
}

func TestEnvironment_Hook_NilError(t *testing.T) {
	config := DefaultEnvironmentConfig()
	config.Components.Hook = func(n string) (Hook, error) { return nil, nil }

	env, _ := NewEnvironment(config)
	returned, err := env.Hook("foo")
	if err == nil {
		t.Fatal("should have error")
	}
	if returned != nil {
		t.Fatalf("bad: %#v", returned)
	}
}

func TestEnvironment_Hook_Error(t *testing.T) {
	config := DefaultEnvironmentConfig()
	config.Components.Hook = func(n string) (Hook, error) { return nil, errors.New("foo") }

	env, _ := NewEnvironment(config)
	returned, err := env.Hook("foo")
	if err == nil {
		t.Fatal("should have error")
	}
	if err.Error() != "foo" {
		t.Fatalf("err: %s", err)
	}
	if returned != nil {
		t.Fatalf("bad: %#v", returned)
	}
}

func TestEnvironment_PostProcessor(t *testing.T) {
	pp := &TestPostProcessor{}
	pps := make(map[string]PostProcessor)
	pps["foo"] = pp

	config := DefaultEnvironmentConfig()
	config.Components.PostProcessor = func(n string) (PostProcessor, error) { return pps[n], nil }

	env, _ := NewEnvironment(config)
	returned, err := env.PostProcessor("foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if returned != pp {
		t.Fatalf("bad: %#v", returned)
	}
}

func TestEnvironment_PostProcessor_NilError(t *testing.T) {
	config := DefaultEnvironmentConfig()
	config.Components.PostProcessor = func(n string) (PostProcessor, error) { return nil, nil }

	env, _ := NewEnvironment(config)
	returned, err := env.PostProcessor("foo")
	if err == nil {
		t.Fatal("should have error")
	}
	if returned != nil {
		t.Fatalf("bad: %#v", returned)
	}
}

func TestEnvironment_PostProcessor_Error(t *testing.T) {
	config := DefaultEnvironmentConfig()
	config.Components.PostProcessor = func(n string) (PostProcessor, error) { return nil, errors.New("foo") }

	env, _ := NewEnvironment(config)
	returned, err := env.PostProcessor("foo")
	if err == nil {
		t.Fatal("should be an error")
	}
	if err.Error() != "foo" {
		t.Fatalf("bad err: %s", err)
	}
	if returned != nil {
		t.Fatalf("bad: %#v", returned)
	}
}

func TestEnvironmentProvisioner(t *testing.T) {
	p := &MockProvisioner{}
	ps := make(map[string]Provisioner)
	ps["foo"] = p

	config := DefaultEnvironmentConfig()
	config.Components.Provisioner = func(n string) (Provisioner, error) { return ps[n], nil }

	env, _ := NewEnvironment(config)
	returned, err := env.Provisioner("foo")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if returned != p {
		t.Fatalf("bad: %#v", returned)
	}
}

func TestEnvironmentProvisioner_NilError(t *testing.T) {
	config := DefaultEnvironmentConfig()
	config.Components.Provisioner = func(n string) (Provisioner, error) { return nil, nil }

	env, _ := NewEnvironment(config)
	returned, err := env.Provisioner("foo")
	if err == nil {
		t.Fatal("should have error")
	}
	if returned != nil {
		t.Fatalf("bad: %#v", returned)
	}
}

func TestEnvironmentProvisioner_Error(t *testing.T) {
	config := DefaultEnvironmentConfig()
	config.Components.Provisioner = func(n string) (Provisioner, error) {
		return nil, errors.New("foo")
	}

	env, _ := NewEnvironment(config)
	returned, err := env.Provisioner("foo")
	if err == nil {
		t.Fatal("should have error")
	}
	if err.Error() != "foo" {
		t.Fatalf("err: %s", err)
	}
	if returned != nil {
		t.Fatalf("bad: %#v", returned)
	}
}

func TestEnvironment_SettingUi(t *testing.T) {
	ui := &BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	}

	config := &EnvironmentConfig{}
	config.Ui = ui

	env, _ := NewEnvironment(config)

	if env.Ui() != ui {
		t.Fatalf("UI should be equal: %#v", env.Ui())
	}
}
