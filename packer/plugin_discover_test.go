package packer

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	pluginsdk "github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
)

func newPluginConfig() PluginConfig {
	var conf PluginConfig
	conf.PluginMinPort = 10000
	conf.PluginMaxPort = 25000
	return conf
}

func TestDiscoverReturnsIfMagicCookieSet(t *testing.T) {
	config := newPluginConfig()

	os.Setenv(pluginsdk.MagicCookieKey, pluginsdk.MagicCookieValue)
	defer os.Unsetenv(pluginsdk.MagicCookieKey)

	err := config.Discover()
	if err != nil {
		t.Fatalf("Should not have errored: %s", err)
	}

	if len(config.Builders.List()) != 0 {
		t.Fatalf("Should not have tried to find builders")
	}
}

func TestEnvVarPackerPluginPath(t *testing.T) {
	// Create a temporary directory to store plugins in
	dir, _, cleanUpFunc, err := generateFakePlugins("custom_plugin_dir",
		[]string{"packer-provisioner-partyparrot"})
	if err != nil {
		t.Fatalf("Error creating fake custom plugins: %s", err)
	}

	defer cleanUpFunc()

	// Add temp dir to path.
	os.Setenv("PACKER_PLUGIN_PATH", dir)
	defer os.Unsetenv("PACKER_PLUGIN_PATH")

	config := newPluginConfig()

	err = config.Discover()
	if err != nil {
		t.Fatalf("Should not have errored: %s", err)
	}

	if len(config.Provisioners.List()) == 0 {
		t.Fatalf("Should have found partyparrot provisioner")
	}
	if !config.Provisioners.Has("partyparrot") {
		t.Fatalf("Should have found partyparrot provisioner.")
	}
}

func TestEnvVarPackerPluginPath_MultiplePaths(t *testing.T) {
	// Create a temporary directory to store plugins in
	dir, _, cleanUpFunc, err := generateFakePlugins("custom_plugin_dir",
		[]string{"packer-provisioner-partyparrot"})
	if err != nil {
		t.Fatalf("Error creating fake custom plugins: %s", err)
	}

	defer cleanUpFunc()

	pathsep := ":"
	if runtime.GOOS == "windows" {
		pathsep = ";"
	}

	// Create a second dir to look in that will be empty
	decoyDir, err := ioutil.TempDir("", "decoy")
	if err != nil {
		t.Fatalf("Failed to create a temporary test dir.")
	}
	defer os.Remove(decoyDir)

	pluginPath := dir + pathsep + decoyDir

	// Add temp dir to path.
	os.Setenv("PACKER_PLUGIN_PATH", pluginPath)
	defer os.Unsetenv("PACKER_PLUGIN_PATH")

	config := newPluginConfig()

	err = config.Discover()
	if err != nil {
		t.Fatalf("Should not have errored: %s", err)
	}

	if len(config.Provisioners.List()) == 0 {
		t.Fatalf("Should have found partyparrot provisioner")
	}
	if !config.Provisioners.Has("partyparrot") {
		t.Fatalf("Should have found partyparrot provisioner.")
	}
}

func TestDiscoverDatasource(t *testing.T) {
	// Create a temporary directory to store plugins in
	dir, _, cleanUpFunc, err := generateFakePlugins("custom_plugin_dir",
		[]string{"packer-datasource-partyparrot"})
	if err != nil {
		t.Fatalf("Error creating fake custom plugins: %s", err)
	}

	defer cleanUpFunc()

	pathsep := ":"
	if runtime.GOOS == "windows" {
		pathsep = ";"
	}

	// Create a second dir to look in that will be empty
	decoyDir, err := ioutil.TempDir("", "decoy")
	if err != nil {
		t.Fatalf("Failed to create a temporary test dir.")
	}
	defer os.Remove(decoyDir)

	pluginPath := dir + pathsep + decoyDir

	// Add temp dir to path.
	os.Setenv("PACKER_PLUGIN_PATH", pluginPath)
	defer os.Unsetenv("PACKER_PLUGIN_PATH")

	config := newPluginConfig()

	err = config.Discover()
	if err != nil {
		t.Fatalf("Should not have errored: %s", err)
	}

	if len(config.DataSources.List()) == 0 {
		t.Fatalf("Should have found partyparrot datasource")
	}
	if !config.DataSources.Has("partyparrot") {
		t.Fatalf("Should have found partyparrot datasource.")
	}
}

func generateFakePlugins(dirname string, pluginNames []string) (string, []string, func(), error) {
	dir, err := ioutil.TempDir("", dirname)
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to create temporary test directory: %v", err)
	}

	cleanUpFunc := func() {
		os.RemoveAll(dir)
	}

	var suffix string
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}

	plugins := make([]string, len(pluginNames))
	for i, plugin := range pluginNames {
		plug := filepath.Join(dir, plugin+suffix)
		plugins[i] = plug
		_, err := os.Create(plug)
		if err != nil {
			cleanUpFunc()
			return "", nil, nil, fmt.Errorf("failed to create temporary plugin file (%s): %v", plug, err)
		}
	}

	return dir, plugins, cleanUpFunc, nil
}

// TestHelperProcess isn't a real test. It's used as a helper process
// for multi-component plugin tests.
func TestHelperPlugins(t *testing.T) {
	if os.Getenv("PKR_WANT_TEST_PLUGINS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	pluginName, args := args[0], args[1:]

	allMocks := []map[string]pluginsdk.Set{mockPlugins, defaultNameMock, doubleDefaultMock, badDefaultNameMock}
	for _, mock := range allMocks {
		plugin, found := mock[pluginName]
		if found {
			err := plugin.RunCommand(args...)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	fmt.Fprintf(os.Stderr, "No %q plugin found\n", pluginName)
	os.Exit(2)
}

// HasExec reports whether the current system can start new processes
// using os.StartProcess or (more commonly) exec.Command.
func HasExec() bool {
	switch runtime.GOOS {
	case "js":
		return false
	case "darwin":
		if runtime.GOARCH == "arm64" {
			return false
		}
	case "windows":
		// TODO(azr): Fix this once versioning is added and we know more
		return false
	}
	return true
}

// MustHaveExec checks that the current system can start new processes
// using os.StartProcess or (more commonly) exec.Command.
// If not, MustHaveExec calls t.Skip with an explanation.
func MustHaveExec(t testing.TB) {
	if !HasExec() {
		t.Skipf("skipping test: cannot exec subprocess on %s/%s", runtime.GOOS, runtime.GOARCH)
	}
}

func MustHaveCommand(t testing.TB, cmd string) string {
	path, err := exec.LookPath(cmd)
	if err != nil {
		t.Skipf("skipping test: cannot find the %q command: %v", cmd, err)
	}
	return path
}

func helperCommand(t *testing.T, s ...string) []string {
	MustHaveExec(t)

	cmd := []string{os.Args[0], "-test.run=TestHelperPlugins", "--"}
	return append(cmd, s...)
}

func createMockPlugins(t *testing.T, plugins map[string]pluginsdk.Set) {
	pluginDir, err := tmp.Dir("pkr-multi-component-plugin-test-*")
	{
		// create an exectutable file with a `sh` sheebang
		// this file will look like:
		// #!/bin/sh
		// PKR_WANT_TEST_PLUGINS=1 ...plugin/debug.test -test.run=TestHelperPlugins -- bird $@
		// 'bird' is the mock plugin we want to start
		// $@ just passes all passed arguments
		// This will allow to run the fake plugin from go tests which in turn
		// will run go tests callback to `TestHelperPlugins`, this one will be
		// transparently calling our mock multi-component plugins `mockPlugins`.
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("putting temporary mock plugins in %s", pluginDir)

		shPath := MustHaveCommand(t, "bash")
		for name := range plugins {
			plugin := path.Join(pluginDir, "packer-plugin-"+name)
			fileContent := ""
			fileContent = fmt.Sprintf("#!%s\n", shPath)
			fileContent += strings.Join(
				append([]string{"PKR_WANT_TEST_PLUGINS=1"}, helperCommand(t, name, "$@")...),
				" ")
			if err := ioutil.WriteFile(plugin, []byte(fileContent), os.ModePerm); err != nil {
				t.Fatalf("failed to create fake plugin binary: %v", err)
			}
		}
	}
	os.Setenv("PACKER_PLUGIN_PATH", pluginDir)
}

var (
	mockPlugins = map[string]pluginsdk.Set{
		"bird": pluginsdk.Set{
			Builders: map[string]packersdk.Builder{
				"feather":   nil,
				"guacamole": nil,
			},
		},
		"chimney": pluginsdk.Set{
			PostProcessors: map[string]packersdk.PostProcessor{
				"smoke": nil,
			},
		},
		"data": pluginsdk.Set{
			Datasources: map[string]packersdk.Datasource{
				"source": nil,
			},
		},
	}

	defaultNameMock = map[string]pluginsdk.Set{
		"foo": pluginsdk.Set{
			Builders: map[string]packersdk.Builder{
				"bar":                  nil,
				"baz":                  nil,
				pluginsdk.DEFAULT_NAME: nil,
			},
		},
	}

	doubleDefaultMock = map[string]pluginsdk.Set{
		"yolo": pluginsdk.Set{
			Builders: map[string]packersdk.Builder{
				"bar":                  nil,
				"baz":                  nil,
				pluginsdk.DEFAULT_NAME: nil,
			},
			PostProcessors: map[string]packersdk.PostProcessor{
				pluginsdk.DEFAULT_NAME: nil,
			},
		},
	}

	badDefaultNameMock = map[string]pluginsdk.Set{
		"foo": pluginsdk.Set{
			Builders: map[string]packersdk.Builder{
				"bar":                  nil,
				"baz":                  nil,
				pluginsdk.DEFAULT_NAME: nil,
			},
		},
	}
)

func Test_multiplugin_describe(t *testing.T) {
	createMockPlugins(t, mockPlugins)
	pluginDir := os.Getenv("PACKER_PLUGIN_PATH")
	defer os.RemoveAll(pluginDir)

	c := PluginConfig{}
	err := c.Discover()
	if err != nil {
		t.Fatalf("error discovering plugins; %s", err.Error())
	}

	for mockPluginName, plugin := range mockPlugins {
		for mockBuilderName := range plugin.Builders {
			expectedBuilderName := mockPluginName + "-" + mockBuilderName

			if !c.Builders.Has(expectedBuilderName) {
				t.Fatalf("expected to find builder %q", expectedBuilderName)
			}
		}
		for mockProvisionerName := range plugin.Provisioners {
			expectedProvisionerName := mockPluginName + "-" + mockProvisionerName
			if !c.Provisioners.Has(expectedProvisionerName) {
				t.Fatalf("expected to find builder %q", expectedProvisionerName)
			}
		}
		for mockPostProcessorName := range plugin.PostProcessors {
			expectedPostProcessorName := mockPluginName + "-" + mockPostProcessorName
			if !c.PostProcessors.Has(expectedPostProcessorName) {
				t.Fatalf("expected to find post-processor %q", expectedPostProcessorName)
			}
		}
		for mockDatasourceName := range plugin.Datasources {
			expectedDatasourceName := mockPluginName + "-" + mockDatasourceName
			if !c.DataSources.Has(expectedDatasourceName) {
				t.Fatalf("expected to find datasource %q", expectedDatasourceName)
			}
		}
	}
}

func Test_multiplugin_defaultName(t *testing.T) {
	createMockPlugins(t, defaultNameMock)
	pluginDir := os.Getenv("PACKER_PLUGIN_PATH")
	defer os.RemoveAll(pluginDir)

	c := PluginConfig{}
	err := c.Discover()
	if err != nil {
		t.Fatalf("error discovering plugins; %s ; mocks are %#v", err.Error(), defaultNameMock)
	}

	expectedBuilderNames := []string{"foo-bar", "foo-baz", "foo"}
	for _, mockBuilderName := range expectedBuilderNames {
		if !c.Builders.Has(mockBuilderName) {
			t.Fatalf("expected to find builder %q; builders is %#v", mockBuilderName, c.Builders)
		}
	}
}

func Test_only_one_multiplugin_defaultName_each_plugin_type(t *testing.T) {
	createMockPlugins(t, doubleDefaultMock)
	pluginDir := os.Getenv("PACKER_PLUGIN_PATH")
	defer os.RemoveAll(pluginDir)

	c := PluginConfig{}
	err := c.Discover()
	if err != nil {
		t.Fatal("Should not have error because pluginsdk.DEFAULT_NAME is used twice but only once per plugin type.")
	}
}
