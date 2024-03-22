// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package packer

import (
	"crypto/sha256"
	"fmt"
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
	"github.com/hashicorp/packer-plugin-sdk/version"
	plugingetter "github.com/hashicorp/packer/packer/plugin-getter"
)

func newPluginConfig() PluginConfig {
	var conf PluginConfig
	conf.PluginMinPort = 10000
	conf.PluginMaxPort = 25000
	return conf
}

func TestDiscoverReturnsIfMagicCookieSet(t *testing.T) {
	config := newPluginConfig()

	t.Setenv(pluginsdk.MagicCookieKey, pluginsdk.MagicCookieValue)

	err := config.Discover()
	if err != nil {
		t.Fatalf("Should not have errored: %s", err)
	}

	if len(config.Builders.List()) != 0 {
		t.Fatalf("Should not have tried to find builders")
	}
}

func TestMultiPlugin_describe(t *testing.T) {
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
				t.Errorf("expected to find builder %q", expectedBuilderName)
			}
		}
		for mockProvisionerName := range plugin.Provisioners {
			expectedProvisionerName := mockPluginName + "-" + mockProvisionerName
			if !c.Provisioners.Has(expectedProvisionerName) {
				t.Errorf("expected to find builder %q", expectedProvisionerName)
			}
		}
		for mockPostProcessorName := range plugin.PostProcessors {
			expectedPostProcessorName := mockPluginName + "-" + mockPostProcessorName
			if !c.PostProcessors.Has(expectedPostProcessorName) {
				t.Errorf("expected to find post-processor %q", expectedPostProcessorName)
			}
		}
		for mockDatasourceName := range plugin.Datasources {
			expectedDatasourceName := mockPluginName + "-" + mockDatasourceName
			if !c.DataSources.Has(expectedDatasourceName) {
				t.Errorf("expected to find datasource %q", expectedDatasourceName)
			}
		}
	}
}

func TestMultiPlugin_describe_installed(t *testing.T) {
	createMockInstalledPlugins(t, mockInstalledPlugins, createMockChecksumFile)
	pluginDir := os.Getenv("PACKER_PLUGIN_PATH")
	defer os.RemoveAll(pluginDir)

	c := PluginConfig{}
	err := c.Discover()
	if err != nil {
		t.Fatalf("error discovering plugins; %s", err.Error())
	}

	for mockPluginName, plugin := range mockInstalledPlugins {
		mockPluginName = strings.Split(mockPluginName, "_")[0]
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

func TestMultiPlugin_describe_installed_for_invalid(t *testing.T) {
	tc := []struct {
		desc                 string
		installedPluginsMock map[string]pluginsdk.Set
		createMockFn         func(*testing.T, map[string]pluginsdk.Set)
	}{
		{
			desc:                 "Incorrectly named plugins",
			installedPluginsMock: invalidInstalledPluginsMock,
			createMockFn: func(t *testing.T, mocks map[string]pluginsdk.Set) {
				createMockInstalledPlugins(t, mocks, createMockChecksumFile)
			},
		},
		{
			desc:                 "Plugins missing checksums",
			installedPluginsMock: mockInstalledPlugins,
			createMockFn: func(t *testing.T, mocks map[string]pluginsdk.Set) {
				createMockInstalledPlugins(t, mocks)
			},
		},
	}

	for _, tt := range tc {
		t.Run(tt.desc, func(t *testing.T) {
			tt.createMockFn(t, tt.installedPluginsMock)
			pluginDir := os.Getenv("PACKER_PLUGIN_PATH")
			defer os.RemoveAll(pluginDir)

			c := PluginConfig{}
			err := c.Discover()
			if err != nil {
				t.Fatalf("error discovering plugins; %s", err.Error())
			}
			if c.Builders.Has("feather") {
				t.Fatalf("expected to not find builder %q", "feather")
			}
			for mockPluginName, plugin := range tt.installedPluginsMock {
				mockPluginName = strings.Split(mockPluginName, "_")[0]
				for mockBuilderName := range plugin.Builders {
					expectedBuilderName := mockPluginName + "-" + mockBuilderName
					if c.Builders.Has(expectedBuilderName) {
						t.Fatalf("expected to not find builder %q", expectedBuilderName)
					}
				}
				for mockProvisionerName := range plugin.Provisioners {
					expectedProvisionerName := mockPluginName + "-" + mockProvisionerName
					if c.Provisioners.Has(expectedProvisionerName) {
						t.Fatalf("expected to not find builder %q", expectedProvisionerName)
					}
				}
				for mockPostProcessorName := range plugin.PostProcessors {
					expectedPostProcessorName := mockPluginName + "-" + mockPostProcessorName
					if c.PostProcessors.Has(expectedPostProcessorName) {
						t.Fatalf("expected to not find post-processor %q", expectedPostProcessorName)
					}
				}
				for mockDatasourceName := range plugin.Datasources {
					expectedDatasourceName := mockPluginName + "-" + mockDatasourceName
					if c.DataSources.Has(expectedDatasourceName) {
						t.Fatalf("expected to not find datasource %q", expectedDatasourceName)
					}
				}
			}
		})
	}
}

func TestMultiPlugin_defaultName(t *testing.T) {
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

func TestMultiPlugin_IgnoreChecksumFile(t *testing.T) {
	createMockPlugins(t, defaultNameMock)
	pluginDir := os.Getenv("PACKER_PLUGIN_PATH")
	defer os.RemoveAll(pluginDir)

	fooPluginName := fmt.Sprintf("packer-plugin-foo_v1.0.0_x5.0_%s_%s", runtime.GOOS, runtime.GOARCH)
	fooPluginPath := filepath.Join(pluginDir, "github.com", "hashicorp", "foo", fooPluginName)
	csFile, err := generateMockChecksumFile(fooPluginPath)
	if err != nil {
		t.Fatal(err.Error())
	}

	// Copy plugin contents into checksum file to validate that it is not only skipped but that it never gets loaded
	if err := os.Rename(fooPluginPath, csFile); err != nil {
		t.Fatalf("failed to rename plugin bin file to checkfum file needed for test: %s", err)
	}

	c := PluginConfig{}
	err = c.Discover()
	if err != nil {
		t.Fatalf("error discovering plugins; %s ; mocks are %#v", err.Error(), defaultNameMock)
	}
	expectedBuilderNames := []string{"foo-bar", "foo-baz", "foo"}
	for _, mockBuilderName := range expectedBuilderNames {
		if c.Builders.Has(mockBuilderName) {
			t.Fatalf("expected to not find builder %q; builders is %#v", mockBuilderName, c.Builders)
		}
	}
}

func TestMultiPlugin_defaultName_each_plugin_type(t *testing.T) {
	createMockPlugins(t, doubleDefaultMock)
	pluginDir := os.Getenv("PACKER_PLUGIN_PATH")
	defer os.RemoveAll(pluginDir)

	c := PluginConfig{}
	err := c.Discover()
	if err != nil {
		t.Fatal("Should not have error because pluginsdk.DEFAULT_NAME is used twice but only once per plugin type.")
	}
}

func generateFakePlugins(dirname string, pluginNames []string) (string, []string, func(), error) {
	dir, err := os.MkdirTemp("", dirname)
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
			plugin.SetVersion(version.InitializePluginVersion("1.0.0", ""))
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
			pluginName := fmt.Sprintf("packer-plugin-%s_v1.0.0_x5.0_%s_%s", name, runtime.GOOS, runtime.GOARCH)
			pluginSubDir := fmt.Sprintf("github.com/hashicorp/%s", name)
			err := os.MkdirAll(path.Join(pluginDir, pluginSubDir), 0755)
			if err != nil {
				t.Fatalf("failed to create plugin hierarchy: %s", err)
			}
			plugin := path.Join(pluginDir, pluginSubDir, pluginName)
			t.Logf("creating fake plugin %s", plugin)
			fileContent := ""
			fileContent = fmt.Sprintf("#!%s\n", shPath)
			fileContent += strings.Join(
				append([]string{"PKR_WANT_TEST_PLUGINS=1"}, helperCommand(t, name, "$@")...),
				" ")
			if err := os.WriteFile(plugin, []byte(fileContent), os.ModePerm); err != nil {
				t.Fatalf("failed to create fake plugin binary: %v", err)
			}

			if _, err := generateMockChecksumFile(plugin); err != nil {
				t.Fatalf("failed to create fake plugin binary checksum file: %v", err)
			}
		}
	}
	t.Setenv("PACKER_PLUGIN_PATH", pluginDir)
}

func createMockChecksumFile(t testing.TB, filePath string) {
	t.Helper()
	cs, err := generateMockChecksumFile(filePath)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("created fake plugin checksum file %s", cs)
}

func generateMockChecksumFile(filePath string) (string, error) {
	cs := plugingetter.Checksummer{
		Type: "sha256",
		Hash: sha256.New(),
	}

	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open fake plugin binary: %v", err)
	}
	defer f.Close()

	sum, err := cs.Sum(f)
	if err != nil {
		return "", fmt.Errorf("failed to checksum fake plugin binary: %v", err)
	}

	sumfile := filePath + cs.FileExt()
	if err := os.WriteFile(sumfile, []byte(fmt.Sprintf("%x", sum)), os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to write checksum fake plugin binary: %v", err)
	}
	return sumfile, nil
}

func createMockInstalledPlugins(t *testing.T, plugins map[string]pluginsdk.Set, opts ...func(tb testing.TB, filePath string)) {
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
		dir, err := os.MkdirTemp(pluginDir, "github.com")
		if err != nil {
			t.Fatalf("failed to create temporary test directory: %v", err)
		}
		dir, err = os.MkdirTemp(dir, "hashicorp")
		if err != nil {
			t.Fatalf("failed to create temporary test directory: %v", err)
		}
		dir, err = os.MkdirTemp(dir, "plugin")
		if err != nil {
			t.Fatalf("failed to create temporary test directory: %v", err)
		}
		t.Logf("putting temporary mock installed plugins in %s", dir)

		shPath := MustHaveCommand(t, "bash")
		for name := range plugins {
			plugin := path.Join(dir, "packer-plugin-"+name)
			t.Logf("creating fake plugin %s", plugin)
			fileContent := ""
			fileContent = fmt.Sprintf("#!%s\n", shPath)
			fileContent += strings.Join(
				append([]string{"PKR_WANT_TEST_PLUGINS=1"}, helperCommand(t, strings.Split(name, "_")[0], "$@")...),
				" ")
			if err := os.WriteFile(plugin, []byte(fileContent), os.ModePerm); err != nil {
				t.Fatalf("failed to create fake plugin binary: %v", err)
			}

			for _, opt := range opts {
				opt(t, plugin)
			}
		}
	}
	t.Setenv("PACKER_PLUGIN_PATH", pluginDir)
}

func getFormattedInstalledPluginSuffix() string {
	return fmt.Sprintf("v1.0.0_x5.0_%s_%s", runtime.GOOS, runtime.GOARCH)
}

var (
	mockPlugins                 = map[string]pluginsdk.Set{}
	mockInstalledPlugins        = map[string]pluginsdk.Set{}
	invalidInstalledPluginsMock = map[string]pluginsdk.Set{}
	defaultNameMock             = map[string]pluginsdk.Set{}
	doubleDefaultMock           = map[string]pluginsdk.Set{}
	badDefaultNameMock          = map[string]pluginsdk.Set{}
)

func init() {
	mockPluginsBird := pluginsdk.NewSet()
	mockPluginsBird.Builders = map[string]packersdk.Builder{
		"feather":   nil,
		"guacamole": nil,
	}
	mockPluginsChim := pluginsdk.NewSet()
	mockPluginsChim.PostProcessors = map[string]packersdk.PostProcessor{
		"smoke": nil,
	}
	mockPluginsData := pluginsdk.NewSet()
	mockPluginsData.Datasources = map[string]packersdk.Datasource{
		"source": nil,
	}
	mockPlugins["bird"] = *mockPluginsBird
	mockPlugins["chimney"] = *mockPluginsChim
	mockPlugins["data"] = *mockPluginsData

	mockInstalledPluginsBird := pluginsdk.NewSet()
	mockInstalledPluginsBird.Builders = map[string]packersdk.Builder{
		"feather":   nil,
		"guacamole": nil,
	}
	mockInstalledPluginsChim := pluginsdk.NewSet()
	mockInstalledPluginsChim.PostProcessors = map[string]packersdk.PostProcessor{
		"smoke": nil,
	}
	mockInstalledPluginsData := pluginsdk.NewSet()
	mockInstalledPluginsData.Datasources = map[string]packersdk.Datasource{
		"source": nil,
	}
	mockInstalledPlugins[fmt.Sprintf("bird_%s", getFormattedInstalledPluginSuffix())] = *mockInstalledPluginsBird
	mockInstalledPlugins[fmt.Sprintf("chimney_%s", getFormattedInstalledPluginSuffix())] = *mockInstalledPluginsChim
	mockInstalledPlugins[fmt.Sprintf("data_%s", getFormattedInstalledPluginSuffix())] = *mockInstalledPluginsData

	invalidInstalledPluginsMockBird := pluginsdk.NewSet()
	invalidInstalledPluginsMockBird.Builders = map[string]packersdk.Builder{
		"feather":   nil,
		"guacamole": nil,
	}
	invalidInstalledPluginsMockChimney := pluginsdk.NewSet()
	invalidInstalledPluginsMockChimney.PostProcessors = map[string]packersdk.PostProcessor{
		"smoke": nil,
	}
	invalidInstalledPluginsMockData := pluginsdk.NewSet()
	invalidInstalledPluginsMockData.Datasources = map[string]packersdk.Datasource{
		"source": nil,
	}
	invalidInstalledPluginsMock["bird_v0.1.1_x5.0_wrong_architecture"] = *invalidInstalledPluginsMockBird
	invalidInstalledPluginsMock["chimney_cool_ranch"] = *invalidInstalledPluginsMockChimney
	invalidInstalledPluginsMock["data"] = *invalidInstalledPluginsMockData

	defaultNameFooSet := pluginsdk.NewSet()
	defaultNameFooSet.Builders = map[string]packersdk.Builder{
		"bar":                  nil,
		"baz":                  nil,
		pluginsdk.DEFAULT_NAME: nil,
	}
	defaultNameMock["foo"] = *defaultNameFooSet

	doubleDefaultYoloSet := pluginsdk.NewSet()
	doubleDefaultYoloSet.Builders = map[string]packersdk.Builder{
		"bar":                  nil,
		"baz":                  nil,
		pluginsdk.DEFAULT_NAME: nil,
	}
	doubleDefaultYoloSet.PostProcessors = map[string]packersdk.PostProcessor{
		pluginsdk.DEFAULT_NAME: nil,
	}
	doubleDefaultMock["yolo"] = *doubleDefaultYoloSet

	badDefaultSet := pluginsdk.NewSet()
	badDefaultSet.Builders = map[string]packersdk.Builder{
		"bar":                  nil,
		"baz":                  nil,
		pluginsdk.DEFAULT_NAME: nil,
	}
	badDefaultNameMock["foo"] = *badDefaultSet
}
