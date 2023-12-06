// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package shell

import (
	"bytes"
	"context"
	"log"
	"os"
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

func TestProvisionerPrepare_extractScript(t *testing.T) {
	config := testConfig()
	p := new(Provisioner)
	_ = p.Prepare(config)
	file, err := extractScript(p)
	defer os.Remove(file)
	if err != nil {
		t.Fatalf("Should not be error: %s", err)
	}
	log.Printf("File: %s", file)
	if strings.Index(file, os.TempDir()) != 0 {
		t.Fatalf("Temp file should reside in %s. File location: %s", os.TempDir(), file)
	}

	// File contents should contain 2 lines concatenated by newlines: foo\nbar
	readFile, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("Should not be error: %s", err)
	}
	expectedContents := "foo\nbar\n"
	s := string(readFile[:])
	if s != expectedContents {
		t.Fatalf("Expected generated inlineScript to equal '%s', got '%s'", expectedContents, s)
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

	if p.config.RemotePath != DefaultRemotePath {
		t.Errorf("unexpected remote path: %s", p.config.RemotePath)
	}

	if p.config.ExecuteCommand != "{{.Vars}}\"{{.Path}}\"" {
		t.Fatalf("Default command should be powershell {{.Vars}}\"{{.Path}}\", but got %s", p.config.ExecuteCommand)
	}
}

func TestProvisionerPrepare_Config(t *testing.T) {

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
	defer tf.Close()

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
	defer tf.Close()

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
	defer tf.Close()

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
	defer tf.Close()

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

func TestProvisionerQuote_EnvironmentVars(t *testing.T) {
	config := testConfig()

	config["environment_vars"] = []string{
		"keyone=valueone",
		"keytwo=value\ntwo",
		"keythree='valuethree'",
		"keyfour='value\nfour'",
		"keyfive='value=five'",
		"keysix='=six'",
	}

	expected := []string{
		"keyone=valueone",
		"keytwo=value\ntwo",
		"keythree='valuethree'",
		"keyfour='value\nfour'",
		"keyfive='value=five'",
		"keysix='=six'",
	}

	p := new(Provisioner)
	p.Prepare(config)

	for i, expectedValue := range expected {
		if p.config.Vars[i] != expectedValue {
			t.Fatalf("%s should be equal to %s", p.config.Vars[i], expectedValue)
		}
	}

}

func testUi() *packersdk.BasicUi {
	return &packersdk.BasicUi{
		Reader:      new(bytes.Buffer),
		Writer:      new(bytes.Buffer),
		ErrorWriter: new(bytes.Buffer),
	}
}

func TestProvisionerProvision_Inline(t *testing.T) {
	config := testConfig()
	delete(config, "inline")

	// Defaults provided by Packer
	config["remote_path"] = "c:/Windows/Temp/inlineScript.bat"
	config["inline"] = []string{"whoami"}
	ui := testUi()
	p := new(Provisioner)

	// Defaults provided by Packer
	p.config.PackerBuildName = "vmware"
	p.config.PackerBuilderType = "iso"
	comm := new(packersdk.MockCommunicator)
	p.Prepare(config)

	err := p.Provision(context.Background(), ui, comm, generatedData())
	if err != nil {
		t.Fatal("should not have error")
	}

	expectedCommand := `set "PACKER_BUILDER_TYPE=iso" && set "PACKER_BUILD_NAME=vmware" && "c:/Windows/Temp/inlineScript.bat"`

	// Should run the command without alteration
	if comm.StartCmd.Command != expectedCommand {
		t.Fatalf("Expect command to be: %s, got %s", expectedCommand, comm.StartCmd.Command)
	}

	envVars := make([]string, 2)
	envVars[0] = "FOO=BAR"
	envVars[1] = "BAR=BAZ"
	config["environment_vars"] = envVars
	config["remote_path"] = "c:/Windows/Temp/inlineScript.bat"

	p.Prepare(config)
	err = p.Provision(context.Background(), ui, comm, generatedData())
	if err != nil {
		t.Fatal("should not have error")
	}

	expectedCommand = `set "BAR=BAZ" && set "FOO=BAR" && set "PACKER_BUILDER_TYPE=iso" && set "PACKER_BUILD_NAME=vmware" && "c:/Windows/Temp/inlineScript.bat"`

	// Should run the command without alteration
	if comm.StartCmd.Command != expectedCommand {
		t.Fatalf("Expect command to be: %s, got: %s", expectedCommand, comm.StartCmd.Command)
	}
}

func TestProvisionerProvision_Scripts(t *testing.T) {
	tf, err := os.CreateTemp("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()

	config := testConfig()
	delete(config, "inline")
	config["scripts"] = []string{tf.Name()}
	config["packer_build_name"] = "foobuild"
	config["packer_builder_type"] = "footype"
	ui := testUi()

	p := new(Provisioner)
	comm := new(packersdk.MockCommunicator)
	p.Prepare(config)
	err = p.Provision(context.Background(), ui, comm, generatedData())
	if err != nil {
		t.Fatal("should not have error")
	}

	//powershell -Command "$env:PACKER_BUILDER_TYPE=''"; powershell -Command "$env:PACKER_BUILD_NAME='foobuild'";  powershell -Command c:/Windows/Temp/script.ps1
	expectedCommand := `set "PACKER_BUILDER_TYPE=footype" && set "PACKER_BUILD_NAME=foobuild" && "c:/Windows/Temp/script.bat"`

	// Should run the command without alteration
	if comm.StartCmd.Command != expectedCommand {
		t.Fatalf("Expect command to be %s NOT %s", expectedCommand, comm.StartCmd.Command)
	}
}

func TestProvisionerProvision_ScriptsWithEnvVars(t *testing.T) {
	tf, err := os.CreateTemp("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()

	config := testConfig()
	ui := testUi()
	delete(config, "inline")

	config["scripts"] = []string{tf.Name()}
	config["packer_build_name"] = "foobuild"
	config["packer_builder_type"] = "footype"

	// Env vars - currently should not effect them
	envVars := make([]string, 2)
	envVars[0] = "FOO=BAR"
	envVars[1] = "BAR=BAZ"
	config["environment_vars"] = envVars

	p := new(Provisioner)
	comm := new(packersdk.MockCommunicator)
	p.Prepare(config)
	err = p.Provision(context.Background(), ui, comm, generatedData())
	if err != nil {
		t.Fatal("should not have error")
	}

	expectedCommand := `set "BAR=BAZ" && set "FOO=BAR" && set "PACKER_BUILDER_TYPE=footype" && set "PACKER_BUILD_NAME=foobuild" && "c:/Windows/Temp/script.bat"`

	// Should run the command without alteration
	if comm.StartCmd.Command != expectedCommand {
		t.Fatalf("Expect command to be %s NOT %s", expectedCommand, comm.StartCmd.Command)
	}
}

func TestProvisioner_createFlattenedEnvVars_windows(t *testing.T) {
	var flattenedEnvVars string
	config := testConfig()

	userEnvVarTests := [][]string{
		{},                     // No user env var
		{"FOO=bar"},            // Single user env var
		{"FOO=bar", "BAZ=qux"}, // Multiple user env vars
		{"FOO=bar=baz"},        // User env var with value containing equals
		{"FOO==bar"},           // User env var with value starting with equals
	}
	userEnvVarmapTests := []map[string]string{
		{},
		{
			"BAR": "foo",
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
		`set "PACKER_BUILDER_TYPE=iso" && set "PACKER_BUILD_NAME=vmware" && `,
		`set "BAR=foo" && set "FOO=bar" && set "PACKER_BUILDER_TYPE=iso" && set "PACKER_BUILD_NAME=vmware" && `,
		`set "BAR=foo" && set "BAZ=qux" && set "FOO=bar" && set "PACKER_BUILDER_TYPE=iso" && set "PACKER_BUILD_NAME=vmware" && set "YAR=yaa" && `,
		`set "BAR=foo=yar" && set "FOO=bar=baz" && set "PACKER_BUILDER_TYPE=iso" && set "PACKER_BUILD_NAME=vmware" && `,
		`set "BAR==foo" && set "FOO==bar" && set "PACKER_BUILDER_TYPE=iso" && set "PACKER_BUILD_NAME=vmware" && `,
	}

	p := new(Provisioner)
	p.generatedData = generatedData()
	p.Prepare(config)

	// Defaults provided by Packer
	p.config.PackerBuildName = "vmware"
	p.config.PackerBuilderType = "iso"

	for i, expectedValue := range expected {
		p.config.Vars = userEnvVarTests[i]
		p.config.Env = userEnvVarmapTests[i]
		flattenedEnvVars = p.createFlattenedEnvVars()
		if flattenedEnvVars != expectedValue {
			t.Fatalf("expected flattened env vars to be: %s, got %s.", expectedValue, flattenedEnvVars)
		}
	}
}

func TestCancel(t *testing.T) {
	// Don't actually call Cancel() as it performs an os.Exit(0)
	// which kills the 'go test' tool
}
func generatedData() map[string]interface{} {
	return map[string]interface{}{
		"PackerHTTPAddr": commonsteps.HttpAddrNotImplemented,
		"PackerHTTPIP":   commonsteps.HttpIPNotImplemented,
		"PackerHTTPPort": commonsteps.HttpPortNotImplemented,
	}
}
