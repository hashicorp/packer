// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package shell

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"inline": []interface{}{"foo", "bar"},
	}
}

func TestProvisioner_Impl(t *testing.T) {
	var raw interface{}
	raw = &Provisioner{}
	if _, ok := raw.(packersdk.Provisioner); !ok {
		t.Fatalf("must be a Provisioner")
	}
}

func TestProvisionerPrepare_Defaults(t *testing.T) {
	var p Provisioner
	config := testConfig()

	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.ExpectDisconnect != false {
		t.Errorf("expected ExpectDisconnect to default to false")
	}

	if p.config.RemotePath == "" {
		t.Errorf("unexpected remote path: %s", p.config.RemotePath)
	}
}

func TestProvisionerPrepare_ExpectDisconnect(t *testing.T) {
	config := testConfig()
	p := new(Provisioner)
	config["expect_disconnect"] = false

	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.ExpectDisconnect != false {
		t.Errorf("expected ExpectDisconnect to be false")
	}

	config["expect_disconnect"] = true
	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.ExpectDisconnect != true {
		t.Errorf("expected ExpectDisconnect to be true")
	}
}

func TestProvisionerPrepare_InlineShebang(t *testing.T) {
	config := testConfig()

	delete(config, "inline_shebang")
	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if p.config.InlineShebang != "/bin/sh -e" {
		t.Fatalf("bad value: %s", p.config.InlineShebang)
	}

	// Test with a good one
	config["inline_shebang"] = "foo"
	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if p.config.InlineShebang != "foo" {
		t.Fatalf("bad value: %s", p.config.InlineShebang)
	}
}

func TestProvisionerPrepare_InvalidKey(t *testing.T) {
	var p Provisioner
	config := testConfig()

	// Add a random key
	config["i_should_not_be_valid"] = true
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestProvisionerPrepare_Script(t *testing.T) {
	config := testConfig()
	delete(config, "inline")

	config["script"] = "/this/should/not/exist"
	p := new(Provisioner)
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	tf, err := os.CreateTemp("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	config["script"] = tf.Name()
	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestProvisionerPrepare_ScriptAndInline(t *testing.T) {
	var p Provisioner
	config := testConfig()

	delete(config, "inline")
	delete(config, "script")
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with both
	tf, err := os.CreateTemp("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	config["inline"] = []interface{}{"foo"}
	config["script"] = tf.Name()
	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestProvisionerPrepare_ScriptAndScripts(t *testing.T) {
	var p Provisioner
	config := testConfig()

	// Test with both
	tf, err := os.CreateTemp("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	config["inline"] = []interface{}{"foo"}
	config["scripts"] = []string{tf.Name()}
	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestProvisionerPrepare_Scripts(t *testing.T) {
	config := testConfig()
	delete(config, "inline")

	config["scripts"] = []string{}
	p := new(Provisioner)
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	tf, err := os.CreateTemp("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	config["scripts"] = []string{tf.Name()}
	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestProvisionerPrepare_EnvironmentVars(t *testing.T) {
	config := testConfig()

	// Test with a bad case
	config["environment_vars"] = []string{"badvar", "good=var"}
	p := new(Provisioner)
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a trickier case
	config["environment_vars"] = []string{"=bad"}
	p = new(Provisioner)
	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good case
	// Note: baz= is a real env variable, just empty
	config["environment_vars"] = []string{"FOO=bar", "baz="}
	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test when the env variable value contains an equals sign
	config["environment_vars"] = []string{"good=withequals=true"}
	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test when the env variable value starts with an equals sign
	config["environment_vars"] = []string{"good==true"}
	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestProvisioner_createFlattenedEnvVars(t *testing.T) {
	var flattenedEnvVars string
	config := testConfig()

	userEnvVarTests := [][]string{
		{},                     // No user env var
		{"FOO=bar"},            // Single user env var
		{"FOO=bar's"},          // User env var with single quote in value
		{"FOO=bar", "BAZ=qux"}, // Multiple user env vars
		{"FOO=bar=baz"},        // User env var with value containing equals
		{"FOO==bar"},           // User env var with value starting with equals
	}
	userEnvVarEnvmapTests := []map[string]string{
		{},
		{
			"BAR": "foo",
		},
		{
			"BAR": "foo's",
		},
		{
			"BAR": "foo",
			"YAR": "yaa",
		},
		{
			"BAR": "foo=yar",
		},
		{
			"BAR": "=foo",
		},
	}
	expected := []string{
		`PACKER_BUILDER_TYPE='iso' PACKER_BUILD_NAME='vmware' `,
		`BAR='foo' FOO='bar' PACKER_BUILDER_TYPE='iso' PACKER_BUILD_NAME='vmware' `,
		`BAR='foo'"'"'s' FOO='bar'"'"'s' PACKER_BUILDER_TYPE='iso' PACKER_BUILD_NAME='vmware' `,
		`BAR='foo' BAZ='qux' FOO='bar' PACKER_BUILDER_TYPE='iso' PACKER_BUILD_NAME='vmware' YAR='yaa' `,
		`BAR='foo=yar' FOO='bar=baz' PACKER_BUILDER_TYPE='iso' PACKER_BUILD_NAME='vmware' `,
		`BAR='=foo' FOO='=bar' PACKER_BUILDER_TYPE='iso' PACKER_BUILD_NAME='vmware' `,
	}

	p := new(Provisioner)
	p.generatedData = generatedData()
	p.Prepare(config)

	// Defaults provided by Packer
	p.config.PackerBuildName = "vmware"
	p.config.PackerBuilderType = "iso"

	for i, expectedValue := range expected {
		p.config.Vars = userEnvVarTests[i]
		p.config.Env = userEnvVarEnvmapTests[i]
		flattenedEnvVars = p.createFlattenedEnvVars()
		if flattenedEnvVars != expectedValue {
			t.Fatalf("expected flattened env vars to be: %s, got %s.", expectedValue, flattenedEnvVars)
		}
	}
}

func TestProvisioner_createFlattenedEnvVars_withEnvVarFormat(t *testing.T) {
	var flattenedEnvVars string
	config := testConfig()

	userEnvVarTests := [][]string{
		{},                     // No user env var
		{"FOO=bar"},            // Single user env var
		{"FOO=bar's"},          // User env var with single quote in value
		{"FOO=bar", "BAZ=qux"}, // Multiple user env vars
		{"FOO=bar=baz"},        // User env var with value containing equals
		{"FOO==bar"},           // User env var with value starting with equals
	}
	userEnvVarEnvmapTests := []map[string]string{
		{},
		{
			"BAR": "foo",
		},
		{
			"BAR": "foo's",
		},
		{
			"BAR": "foo",
			"YAR": "yaa",
		},
		{
			"BAR": "foo=yar",
		},
		{
			"BAR": "=foo",
		},
	}
	expected := []string{
		`PACKER_BUILDER_TYPE=iso PACKER_BUILD_NAME=vmware `,
		`BAR=foo FOO=bar PACKER_BUILDER_TYPE=iso PACKER_BUILD_NAME=vmware `,
		`BAR=foo'"'"'s FOO=bar'"'"'s PACKER_BUILDER_TYPE=iso PACKER_BUILD_NAME=vmware `,
		`BAR=foo BAZ=qux FOO=bar PACKER_BUILDER_TYPE=iso PACKER_BUILD_NAME=vmware YAR=yaa `,
		`BAR=foo=yar FOO=bar=baz PACKER_BUILDER_TYPE=iso PACKER_BUILD_NAME=vmware `,
		`BAR==foo FOO==bar PACKER_BUILDER_TYPE=iso PACKER_BUILD_NAME=vmware `,
	}

	p := new(Provisioner)
	p.generatedData = generatedData()
	p.config.EnvVarFormat = "%s=%s "
	p.Prepare(config)

	// Defaults provided by Packer
	p.config.PackerBuildName = "vmware"
	p.config.PackerBuilderType = "iso"

	for i, expectedValue := range expected {
		p.config.Vars = userEnvVarTests[i]
		p.config.Env = userEnvVarEnvmapTests[i]
		flattenedEnvVars = p.createFlattenedEnvVars()
		if flattenedEnvVars != expectedValue {
			t.Fatalf("expected flattened env vars to be: %s, got %s.", expectedValue, flattenedEnvVars)
		}
	}
}

func TestProvisioner_createEnvVarFileContent(t *testing.T) {
	var flattenedEnvVars string
	config := testConfig()

	userEnvVarTests := [][]string{
		{},                     // No user env var
		{"FOO=bar"},            // Single user env var
		{"FOO=bar's"},          // User env var with single quote in value
		{"FOO=bar", "BAZ=qux"}, // Multiple user env vars
		{"FOO=bar=baz"},        // User env var with value containing equals
		{"FOO==bar"},           // User env var with value starting with equals
	}
	userEnvVarEnvmapTests := []map[string]string{
		{},
		{
			"BAR": "foo",
		},
		{
			"BAR": "foo's",
		},
		{
			"BAR": "foo",
			"YAR": "yaa",
		},
		{
			"BAR": "foo=yar",
		},
		{
			"BAR": "=foo",
		},
	}
	expected := []string{
		`export PACKER_BUILDER_TYPE='iso'
export PACKER_BUILD_NAME='vmware'
`,
		`export BAR='foo'
export FOO='bar'
export PACKER_BUILDER_TYPE='iso'
export PACKER_BUILD_NAME='vmware'
`,
		`export BAR='foo'"'"'s'
export FOO='bar'"'"'s'
export PACKER_BUILDER_TYPE='iso'
export PACKER_BUILD_NAME='vmware'
`,
		`export BAR='foo'
export BAZ='qux'
export FOO='bar'
export PACKER_BUILDER_TYPE='iso'
export PACKER_BUILD_NAME='vmware'
export YAR='yaa'
`,
		`export BAR='foo=yar'
export FOO='bar=baz'
export PACKER_BUILDER_TYPE='iso'
export PACKER_BUILD_NAME='vmware'
`,
		`export BAR='=foo'
export FOO='=bar'
export PACKER_BUILDER_TYPE='iso'
export PACKER_BUILD_NAME='vmware'
`,
	}

	p := new(Provisioner)
	p.generatedData = generatedData()
	p.config.UseEnvVarFile = true
	p.Prepare(config)

	// Defaults provided by Packer
	p.config.PackerBuildName = "vmware"
	p.config.PackerBuilderType = "iso"

	for i, expectedValue := range expected {
		p.config.Vars = userEnvVarTests[i]
		p.config.Env = userEnvVarEnvmapTests[i]
		flattenedEnvVars = p.createEnvVarFileContent()
		if flattenedEnvVars != expectedValue {
			t.Fatalf("expected flattened env vars to be: %s, got %s.", expectedValue, flattenedEnvVars)
		}
	}
}

func TestProvisioner_createEnvVarFileContent_withEnvVarFormat(t *testing.T) {
	var flattenedEnvVars string
	config := testConfig()

	userEnvVarTests := [][]string{
		{},                     // No user env var
		{"FOO=bar", "BAZ=qux"}, // Multiple user env vars
		{"FOO=bar=baz"},        // User env var with value containing equals
		{"FOO==bar"},           // User env var with value starting with equals
	}
	userEnvVarEnvmapTests := []map[string]string{
		{},
		{
			"BAR": "foo",
			"YAR": "yaa",
		},
		{
			"BAR": "foo=yar",
		},
		{
			"BAR": "=foo",
		},
	}
	expected := []string{
		`PACKER_BUILDER_TYPE=iso
PACKER_BUILD_NAME=vmware
`,
		`BAR=foo
BAZ=qux
FOO=bar
PACKER_BUILDER_TYPE=iso
PACKER_BUILD_NAME=vmware
YAR=yaa
`,
		`BAR=foo=yar
FOO=bar=baz
PACKER_BUILDER_TYPE=iso
PACKER_BUILD_NAME=vmware
`,
		`BAR==foo
FOO==bar
PACKER_BUILDER_TYPE=iso
PACKER_BUILD_NAME=vmware
`,
	}

	p := new(Provisioner)
	p.generatedData = generatedData()
	p.config.UseEnvVarFile = true
	//User provided env_var_format without export prefix
	p.config.EnvVarFormat = "%s=%s\n"
	p.Prepare(config)

	// Defaults provided by Packer
	p.config.PackerBuildName = "vmware"
	p.config.PackerBuilderType = "iso"

	for i, expectedValue := range expected {
		p.config.Vars = userEnvVarTests[i]
		p.config.Env = userEnvVarEnvmapTests[i]
		flattenedEnvVars = p.createEnvVarFileContent()
		if flattenedEnvVars != expectedValue {
			t.Fatalf("expected flattened env vars to be: %q, got %q.", expectedValue, flattenedEnvVars)
		}
	}
}

func TestProvisioner_RemoteFolderSetSuccessfully(t *testing.T) {
	config := testConfig()

	expectedRemoteFolder := "/example/path"
	config["remote_folder"] = expectedRemoteFolder

	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if !strings.Contains(p.config.RemotePath, expectedRemoteFolder) {
		t.Fatalf("remote path does not contain remote_folder")
	}
}

func TestProvisioner_RemoteFolderDefaultsToTmp(t *testing.T) {
	config := testConfig()

	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if p.config.RemoteFolder != "/tmp" {
		t.Fatalf("remote_folder did not default to /tmp")
	}

	if !strings.Contains(p.config.RemotePath, "/tmp") {
		t.Fatalf("remote path does not contain remote_folder")
	}
}

func TestProvisioner_RemoteFileSetSuccessfully(t *testing.T) {
	config := testConfig()

	expectedRemoteFile := "example.sh"
	config["remote_file"] = expectedRemoteFile

	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if !strings.Contains(p.config.RemotePath, expectedRemoteFile) {
		t.Fatalf("remote path does not contain remote_file")
	}
}

func TestProvisioner_RemoteFileDefaultsToScriptnnnn(t *testing.T) {
	config := testConfig()

	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	remoteFileRegex := regexp.MustCompile("script_[0-9]{1,4}.sh")

	if !remoteFileRegex.MatchString(p.config.RemoteFile) {
		t.Fatalf("remote_file did not default to script_nnnn.sh: %q", p.config.RemoteFile)
	}

	if !remoteFileRegex.MatchString(p.config.RemotePath) {
		t.Fatalf("remote_path did not match script_nnnn.sh: %q", p.config.RemotePath)
	}
}

func TestProvisioner_RemotePathSetViaRemotePathAndRemoteFile(t *testing.T) {
	config := testConfig()

	expectedRemoteFile := "example.sh"
	expectedRemoteFolder := "/example/path"
	config["remote_file"] = expectedRemoteFile
	config["remote_folder"] = expectedRemoteFolder

	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if p.config.RemotePath != expectedRemoteFolder+"/"+expectedRemoteFile {
		t.Fatalf("remote path does not contain remote_file: %q", p.config.RemotePath)
	}
}

func TestProvisioner_RemotePathOverridesRemotePathAndRemoteFile(t *testing.T) {
	config := testConfig()

	expectedRemoteFile := "example.sh"
	expectedRemoteFolder := "/example/path"
	expectedRemotePath := "/example/remote/path/script.sh"
	config["remote_file"] = expectedRemoteFile
	config["remote_folder"] = expectedRemoteFolder
	config["remote_path"] = expectedRemotePath

	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if p.config.RemotePath != expectedRemotePath {
		t.Fatalf("remote path does not contain remote_path: %q", p.config.RemotePath)
	}
}

func TestProvisionerRemotePathDefaultsSuccessfully(t *testing.T) {
	config := testConfig()

	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	remotePathRegex := regexp.MustCompile("/tmp/script_[0-9]{1,4}.sh")

	if !remotePathRegex.MatchString(p.config.RemotePath) {
		t.Fatalf("remote path does not match the expected default regex: %q", p.config.RemotePath)
	}
}

func generatedData() map[string]interface{} {
	return map[string]interface{}{
		"PackerHTTPAddr": commonsteps.HttpAddrNotImplemented,
		"PackerHTTPIP":   commonsteps.HttpIPNotImplemented,
		"PackerHTTPPort": commonsteps.HttpPortNotImplemented,
	}
}
