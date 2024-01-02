// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package powershell

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/assert"
)

func TestProvisionerPrepare_extractScript(t *testing.T) {
	config := testConfig()
	p := new(Provisioner)
	_ = p.Prepare(config)
	file, err := extractScript(p)
	defer os.Remove(file)
	if err != nil {
		t.Fatalf("Should not be error: %s", err)
	}
	t.Logf("File: %s", file)
	if strings.Index(file, os.TempDir()) != 0 {
		t.Fatalf("Temp file should reside in %s. File location: %s", os.TempDir(), file)
	}

	// File contents should contain 2 lines concatenated by newlines: foo\nbar
	readFile, err := os.ReadFile(file)
	expectedContents := "foo\nbar\n"
	if err != nil {
		t.Fatalf("Should not be error: %s", err)
	}
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

	matched, _ := regexp.MatchString("c:/Windows/Temp/script-.*.ps1", p.config.RemotePath)
	if !matched {
		t.Errorf("unexpected remote path: %s", p.config.RemotePath)
	}

	if p.config.ElevatedUser != "" {
		t.Error("expected elevated_user to be empty")
	}
	if p.config.ElevatedPassword != "" {
		t.Error("expected elevated_password to be empty")
	}

	if p.config.ExecuteCommand != `powershell -executionpolicy bypass "& { if (Test-Path variable:global:ProgressPreference){set-variable -name variable:global:ProgressPreference -value 'SilentlyContinue'};. {{.Vars}}; &'{{.Path}}'; exit $LastExitCode }"` {
		t.Fatalf(`Default command should be 'powershell -executionpolicy bypass "& { if (Test-Path variable:global:ProgressPreference){set-variable -name variable:global:ProgressPreference -value 'SilentlyContinue'};. {{.Vars}}; &'{{.Path}}'; exit $LastExitCode }"', but got '%s'`, p.config.ExecuteCommand)
	}

	if p.config.ElevatedExecuteCommand != `powershell -executionpolicy bypass "& { if (Test-Path variable:global:ProgressPreference){set-variable -name variable:global:ProgressPreference -value 'SilentlyContinue'};. {{.Vars}}; &'{{.Path}}'; exit $LastExitCode }"` {
		t.Fatalf(`Default command should be 'powershell -executionpolicy bypass "& { if (Test-Path variable:global:ProgressPreference){set-variable -name variable:global:ProgressPreference -value 'SilentlyContinue'};. {{.Vars}}; &'{{.Path}}'; exit $LastExitCode }"', but got '%s'`, p.config.ElevatedExecuteCommand)
	}

	if p.config.ElevatedEnvVarFormat != `$env:%s="%s"; ` {
		t.Fatalf(`Default command should be powershell '$env:%%s="%%s"; ', but got %s`, p.config.ElevatedEnvVarFormat)
	}
}

func TestProvisionerPrepare_Config(t *testing.T) {
	config := testConfig()
	config["elevated_user"] = "{{user `user`}}"
	config["elevated_password"] = "{{user `password`}}"
	config[common.UserVariablesConfigKey] = map[string]string{
		"user":     "myusername",
		"password": "mypassword",
	}

	var p Provisioner
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.ElevatedUser != "myusername" {
		t.Fatalf("Expected 'myusername' for key `elevated_user`: %s", p.config.ElevatedUser)
	}
	if p.config.ElevatedPassword != "mypassword" {
		t.Fatalf("Expected 'mypassword' for key `elevated_password`: %s", p.config.ElevatedPassword)
	}
}

func TestProvisionerPrepare_DebugMode(t *testing.T) {
	config := testConfig()
	config["debug_mode"] = 1

	var p Provisioner
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	command := `powershell -executionpolicy bypass "& { if (Test-Path variable:global:ProgressPreference){set-variable -name variable:global:ProgressPreference -value 'SilentlyContinue'};Set-PsDebug -Trace 1;. {{.Vars}}; &'{{.Path}}'; exit $LastExitCode }"`
	if p.config.ExecuteCommand != command {
		t.Fatalf(fmt.Sprintf(`Expected command should be '%s' but got '%s'`, command, p.config.ExecuteCommand))
	}
}

func TestProvisionerPrepare_InvalidDebugMode(t *testing.T) {
	config := testConfig()
	config["debug_mode"] = -1

	var p Provisioner
	err := p.Prepare(config)
	if err == nil {
		t.Fatalf("should have error")
	}

	message := "invalid Trace level for `debug_mode`; valid values are 0, 1, and 2"
	if !strings.Contains(err.Error(), message) {
		t.Fatalf("expected Prepare() error %q to contain %q", err.Error(), message)
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

func TestProvisionerPrepare_Elevated(t *testing.T) {
	var p Provisioner
	config := testConfig()

	// Add a random key
	config["elevated_user"] = "vagrant"
	err := p.Prepare(config)

	if err != nil {
		t.Fatal("should not have error")
	}

	config["elevated_password"] = "vagrant"
	err = p.Prepare(config)

	if err != nil {
		t.Fatal("should not have error")
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

func TestProvisionerPrepare_Pwsh(t *testing.T) {

	config := testConfig()

	config["use_pwsh"] = true

	p := new(Provisioner)
	err := p.Prepare(config)

	if err != nil {
		t.Fatalf("Should not be error: %s", err)
	}

	if !p.config.UsePwsh {
		t.Fatalf("Expected 'pwsh' to be: true")
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

func TestProvisionerProvision_ValidExitCodes(t *testing.T) {
	config := testConfig()
	delete(config, "inline")

	// Defaults provided by Packer
	config["remote_path"] = "c:/Windows/Temp/inlineScript.ps1"
	config["inline"] = []string{"whoami"}
	ui := testUi()
	p := new(Provisioner)

	// Defaults provided by Packer
	p.config.PackerBuildName = "vmware"
	p.config.PackerBuilderType = "iso"
	p.config.ValidExitCodes = []int{0, 200}
	comm := new(packersdk.MockCommunicator)
	comm.StartExitStatus = 200
	p.Prepare(config)
	err := p.Provision(context.Background(), ui, comm, generatedData())
	if err != nil {
		t.Fatal("should not have error")
	}
}

func TestProvisionerProvision_PauseAfter(t *testing.T) {
	config := testConfig()
	delete(config, "inline")

	// Defaults provided by Packer
	config["remote_path"] = "c:/Windows/Temp/inlineScript.ps1"
	config["inline"] = []string{"whoami"}
	ui := testUi()
	p := new(Provisioner)

	// Defaults provided by Packer
	p.config.PackerBuildName = "vmware"
	p.config.PackerBuilderType = "iso"
	p.config.ValidExitCodes = []int{0, 200}
	pause_amount := time.Second
	p.config.PauseAfter = pause_amount
	comm := new(packersdk.MockCommunicator)
	comm.StartExitStatus = 200
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("Prepar failed: %s", err)
	}

	start_time := time.Now()
	err = p.Provision(context.Background(), ui, comm, generatedData())
	end_time := time.Now()

	if err != nil {
		t.Fatal("should not have error")
	}

	if end_time.Sub(start_time) < pause_amount {
		t.Fatal("Didn't wait pause_amount")
	}
}

func TestProvisionerProvision_InvalidExitCodes(t *testing.T) {
	config := testConfig()
	delete(config, "inline")

	// Defaults provided by Packer
	config["remote_path"] = "c:/Windows/Temp/inlineScript.ps1"
	config["inline"] = []string{"whoami"}
	ui := testUi()
	p := new(Provisioner)

	// Defaults provided by Packer
	p.config.PackerBuildName = "vmware"
	p.config.PackerBuilderType = "iso"
	p.config.ValidExitCodes = []int{0, 200}
	comm := new(packersdk.MockCommunicator)
	comm.StartExitStatus = 201 // Invalid!
	p.Prepare(config)
	err := p.Provision(context.Background(), ui, comm, generatedData())
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestProvisionerProvision_Inline(t *testing.T) {
	// skip_clean is set to true otherwise the last command executed by the provisioner is the cleanup.
	config := testConfigWithSkipClean()
	delete(config, "inline")

	// Defaults provided by Packer
	config["remote_path"] = "c:/Windows/Temp/inlineScript.ps1"
	config["inline"] = []string{"whoami"}
	ui := testUi()
	p := new(Provisioner)

	// Defaults provided by Packer - env vars should not appear in cmd
	p.config.PackerBuildName = "vmware"
	p.config.PackerBuilderType = "iso"
	comm := new(packersdk.MockCommunicator)
	_ = p.Prepare(config)

	err := p.Provision(context.Background(), ui, comm, generatedData())
	if err != nil {
		t.Fatal("should not have error")
	}

	cmd := comm.StartCmd.Command
	re := regexp.MustCompile(`powershell -executionpolicy bypass "& { if \(Test-Path variable:global:ProgressPreference\){set-variable -name variable:global:ProgressPreference -value 'SilentlyContinue'};\. c:/Windows/Temp/packer-ps-env-vars-[[:alnum:]]{8}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{12}\.ps1; &'c:/Windows/Temp/inlineScript.ps1'; exit \$LastExitCode }"`)
	matched := re.MatchString(cmd)
	if !matched {
		t.Fatalf("Got unexpected command: %s", cmd)
	}

	// User supplied env vars should not change things
	envVars := make([]string, 2)
	envVars[0] = "FOO=BAR"
	envVars[1] = "BAR=BAZ"
	config["environment_vars"] = envVars
	config["remote_path"] = "c:/Windows/Temp/inlineScript.ps1"

	p.Prepare(config)
	err = p.Provision(context.Background(), ui, comm, generatedData())
	if err != nil {
		t.Fatal("should not have error")
	}

	cmd = comm.StartCmd.Command
	re = regexp.MustCompile(`powershell -executionpolicy bypass "& { if \(Test-Path variable:global:ProgressPreference\){set-variable -name variable:global:ProgressPreference -value 'SilentlyContinue'};\. c:/Windows/Temp/packer-ps-env-vars-[[:alnum:]]{8}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{12}\.ps1; &'c:/Windows/Temp/inlineScript.ps1'; exit \$LastExitCode }"`)
	matched = re.MatchString(cmd)
	if !matched {
		t.Fatalf("Got unexpected command: %s", cmd)
	}
}

func TestProvisionerProvision_Scripts(t *testing.T) {
	tempFile, _ := os.CreateTemp("", "packer")
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// skip_clean is set to true otherwise the last command executed by the provisioner is the cleanup.
	config := testConfigWithSkipClean()
	delete(config, "inline")
	config["scripts"] = []string{tempFile.Name()}
	config["packer_build_name"] = "foobuild"
	config["packer_builder_type"] = "footype"
	config["remote_path"] = "c:/Windows/Temp/script.ps1"
	ui := testUi()

	p := new(Provisioner)
	comm := new(packersdk.MockCommunicator)
	p.Prepare(config)
	err := p.Provision(context.Background(), ui, comm, generatedData())
	if err != nil {
		t.Fatal("should not have error")
	}

	cmd := comm.StartCmd.Command
	re := regexp.MustCompile(`powershell -executionpolicy bypass "& { if \(Test-Path variable:global:ProgressPreference\){set-variable -name variable:global:ProgressPreference -value 'SilentlyContinue'};\. c:/Windows/Temp/packer-ps-env-vars-[[:alnum:]]{8}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{12}\.ps1; &'c:/Windows/Temp/script.ps1'; exit \$LastExitCode }"`)
	matched := re.MatchString(cmd)
	if !matched {
		t.Fatalf("Got unexpected command: %s", cmd)
	}
}

func TestProvisionerProvision_ScriptsWithEnvVars(t *testing.T) {
	tempFile, _ := os.CreateTemp("", "packer")
	ui := testUi()
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// skip_clean is set to true otherwise the last command executed by the provisioner is the cleanup.
	config := testConfigWithSkipClean()
	delete(config, "inline")

	config["scripts"] = []string{tempFile.Name()}
	config["packer_build_name"] = "foobuild"
	config["packer_builder_type"] = "footype"

	// Env vars - currently should not effect them
	envVars := make([]string, 2)
	envVars[0] = "FOO=BAR"
	envVars[1] = "BAR=BAZ"
	config["environment_vars"] = envVars
	config["remote_path"] = "c:/Windows/Temp/script.ps1"

	p := new(Provisioner)
	comm := new(packersdk.MockCommunicator)
	p.Prepare(config)
	err := p.Provision(context.Background(), ui, comm, generatedData())
	if err != nil {
		t.Fatal("should not have error")
	}

	cmd := comm.StartCmd.Command
	re := regexp.MustCompile(`powershell -executionpolicy bypass "& { if \(Test-Path variable:global:ProgressPreference\){set-variable -name variable:global:ProgressPreference -value 'SilentlyContinue'};\. c:/Windows/Temp/packer-ps-env-vars-[[:alnum:]]{8}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{12}\.ps1; &'c:/Windows/Temp/script.ps1'; exit \$LastExitCode }"`)
	matched := re.MatchString(cmd)
	if !matched {
		t.Fatalf("Got unexpected command: %s", cmd)
	}
}

func TestProvisionerProvision_SkipClean(t *testing.T) {
	tempFile, _ := os.CreateTemp("", "packer")
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	config := map[string]interface{}{
		"scripts":     []string{tempFile.Name()},
		"remote_path": "c:/Windows/Temp/script.ps1",
	}

	tt := []struct {
		SkipClean                bool
		LastExecutedCommandRegex string
	}{
		{
			SkipClean:                true,
			LastExecutedCommandRegex: `powershell -executionpolicy bypass "& { if \(Test-Path variable:global:ProgressPreference\){set-variable -name variable:global:ProgressPreference -value 'SilentlyContinue'};\. c:/Windows/Temp/packer-ps-env-vars-[[:alnum:]]{8}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{12}\.ps1; &'c:/Windows/Temp/script.ps1'; exit \$LastExitCode }"`,
		},
		{
			SkipClean:                false,
			LastExecutedCommandRegex: `powershell -executionpolicy bypass "& { if \(Test-Path variable:global:ProgressPreference\){set-variable -name variable:global:ProgressPreference -value 'SilentlyContinue'};\. c:/Windows/Temp/packer-ps-env-vars-[[:alnum:]]{8}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{12}\.ps1; &'c:/Windows/Temp/packer-cleanup-[[:alnum:]]{8}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{12}\.ps1'; exit \$LastExitCode }"`,
		},
	}

	for _, tc := range tt {
		tc := tc
		p := new(Provisioner)
		ui := testUi()
		comm := new(packersdk.MockCommunicator)

		config["skip_clean"] = tc.SkipClean
		if err := p.Prepare(config); err != nil {
			t.Fatalf("failed to prepare config when SkipClean is %t: %s", tc.SkipClean, err)
		}
		err := p.Provision(context.Background(), ui, comm, generatedData())
		if err != nil {
			t.Fatal("should not have error")
		}

		// When SkipClean is false the last executed command should be the clean up command;
		// otherwise it will be the execution command for the provisioning script.
		cmd := comm.StartCmd.Command
		re := regexp.MustCompile(tc.LastExecutedCommandRegex)
		matched := re.MatchString(cmd)
		if !matched {
			t.Fatalf(`Got unexpected command when SkipClean is %t: %s`, tc.SkipClean, cmd)
		}
	}
}

func TestProvisionerProvision_UploadFails(t *testing.T) {
	config := testConfig()
	ui := testUi()

	p := new(Provisioner)
	comm := new(packersdk.ScriptUploadErrorMockCommunicator)
	p.Prepare(config)
	p.config.StartRetryTimeout = 1 * time.Second
	err := p.Provision(context.Background(), ui, comm, generatedData())
	if !strings.Contains(err.Error(), packersdk.ScriptUploadErrorMockCommunicatorError.Error()) {
		t.Fatalf("expected Provision() error %q to contain %q",
			err.Error(),
			packersdk.ScriptUploadErrorMockCommunicatorError.Error())
	}
}

func TestProvisioner_createFlattenedElevatedEnvVars_windows(t *testing.T) {
	var flattenedEnvVars string
	config := testConfig()

	userEnvVarTests := [][]string{
		{},                     // No user env var
		{"FOO=bar"},            // Single user env var
		{"FOO=bar", "BAZ=qux"}, // Multiple user env vars
		{"FOO=bar=baz"},        // User env var with value containing equals
		{"FOO==bar"},           // User env var with value starting with equals
		// Test escaping of characters special to PowerShell
		{"FOO=bar$baz"},  // User env var with value containing dollar
		{"FOO=bar\"baz"}, // User env var with value containing a double quote
		{"FOO=bar'baz"},  // User env var with value containing a single quote
		{"FOO=bar`baz"},  // User env var with value containing a backtick

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
			"BAR": "foo=yaa",
		},
		{
			"BAR": "=foo",
		},
		{
			"BAR": "foo$yaa",
		},
		{
			"BAR": "foo\"yaa",
		},
		{
			"BAR": "foo'yaa",
		},
		{
			"BAR": "foo`yaa",
		},
	}
	expected := []string{
		`$env:PACKER_BUILDER_TYPE="iso"; $env:PACKER_BUILD_NAME="vmware"; `,
		`$env:BAR="foo"; $env:FOO="bar"; $env:PACKER_BUILDER_TYPE="iso"; $env:PACKER_BUILD_NAME="vmware"; `,
		`$env:BAR="foo"; $env:BAZ="qux"; $env:FOO="bar"; $env:PACKER_BUILDER_TYPE="iso"; $env:PACKER_BUILD_NAME="vmware"; $env:YAR="yaa"; `,
		`$env:BAR="foo=yaa"; $env:FOO="bar=baz"; $env:PACKER_BUILDER_TYPE="iso"; $env:PACKER_BUILD_NAME="vmware"; `,
		`$env:BAR="=foo"; $env:FOO="=bar"; $env:PACKER_BUILDER_TYPE="iso"; $env:PACKER_BUILD_NAME="vmware"; `,
		"$env:BAR=\"foo`$yaa\"; $env:FOO=\"bar`$baz\"; $env:PACKER_BUILDER_TYPE=\"iso\"; $env:PACKER_BUILD_NAME=\"vmware\"; ",
		"$env:BAR=\"foo`\"yaa\"; $env:FOO=\"bar`\"baz\"; $env:PACKER_BUILDER_TYPE=\"iso\"; $env:PACKER_BUILD_NAME=\"vmware\"; ",
		"$env:BAR=\"foo`'yaa\"; $env:FOO=\"bar`'baz\"; $env:PACKER_BUILDER_TYPE=\"iso\"; $env:PACKER_BUILD_NAME=\"vmware\"; ",
		"$env:BAR=\"foo``yaa\"; $env:FOO=\"bar``baz\"; $env:PACKER_BUILDER_TYPE=\"iso\"; $env:PACKER_BUILD_NAME=\"vmware\"; ",
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
		flattenedEnvVars = p.createFlattenedEnvVars(true)
		if flattenedEnvVars != expectedValue {
			t.Fatalf("expected flattened env vars to be: %s, got %s.", expectedValue, flattenedEnvVars)
		}
	}
}

func TestProvisionerCorrectlyInterpolatesValidExitCodes(t *testing.T) {
	type testCases struct {
		Input    interface{}
		Expected []int
	}
	validExitCodeTests := []testCases{
		{"0", []int{0}},
		{[]string{"0"}, []int{0}},
		{[]int{0, 12345}, []int{0, 12345}},
		{[]string{"0", "12345"}, []int{0, 12345}},
		{"0,12345", []int{0, 12345}},
	}

	for _, tc := range validExitCodeTests {
		p := new(Provisioner)
		config := testConfig()
		config["valid_exit_codes"] = tc.Input
		err := p.Prepare(config)

		if err != nil {
			t.Fatalf("Shouldn't have had error interpolating exit codes")
		}
		assert.ElementsMatchf(t, p.config.ValidExitCodes, tc.Expected,
			fmt.Sprintf("expected exit codes to be: %#v, got %#v.", p.config.ValidExitCodes, tc.Expected))
	}
}

func TestProvisionerCorrectlyInterpolatesExecutionPolicy(t *testing.T) {
	type testCases struct {
		Input       interface{}
		Expected    ExecutionPolicy
		ErrExpected bool
	}
	tests := []testCases{
		{
			Input:       "bypass",
			Expected:    ExecutionPolicy(0),
			ErrExpected: false,
		},
		{
			Input:       "allsigned",
			Expected:    ExecutionPolicy(1),
			ErrExpected: false,
		},
		{
			Input:       "default",
			Expected:    ExecutionPolicy(2),
			ErrExpected: false,
		},
		{
			Input:       "remotesigned",
			Expected:    ExecutionPolicy(3),
			ErrExpected: false,
		},
		{
			Input:       "restricted",
			Expected:    ExecutionPolicy(4),
			ErrExpected: false,
		},
		{
			Input:       "undefined",
			Expected:    ExecutionPolicy(5),
			ErrExpected: false,
		},
		{
			Input:       "unrestricted",
			Expected:    ExecutionPolicy(6),
			ErrExpected: false,
		},
		{
			Input:       "none",
			Expected:    ExecutionPolicy(7),
			ErrExpected: false,
		},
		{
			Input:       "0", // User can supply a valid number for policy, too
			Expected:    0,
			ErrExpected: false,
		},
		{
			Input:       "invalid",
			Expected:    0,
			ErrExpected: true,
		},
		{
			Input:       "100", // If number is invalid policy, reject.
			Expected:    100,
			ErrExpected: true,
		},
	}

	for _, tc := range tests {
		p := new(Provisioner)
		config := testConfig()
		config["execution_policy"] = tc.Input
		err := p.Prepare(config)

		if (err != nil) != tc.ErrExpected {
			t.Fatalf("Either err was expected, or shouldn't have happened: %#v", tc)
		}
		if err == nil {
			assert.Equal(t, p.config.ExecutionPolicy, tc.Expected,
				fmt.Sprintf("expected %#v, got %#v.", p.config.ExecutionPolicy, tc.Expected))
		}
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
		// Test escaping of characters special to PowerShell
		{"FOO=bar$baz"},  // User env var with value containing dollar
		{"FOO=bar\"baz"}, // User env var with value containing a double quote
		{"FOO=bar'baz"},  // User env var with value containing a single quote
		{"FOO=bar`baz"},  // User env var with value containing a backtick
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
			"BAR": "foo=yaa",
		},
		{
			"BAR": "=foo",
		},
		{
			"BAR": "foo$yaa",
		},
		{
			"BAR": "foo\"yaa",
		},
		{
			"BAR": "foo'yaa",
		},
		{
			"BAR": "foo`yaa",
		},
	}
	expected := []string{
		`$env:PACKER_BUILDER_TYPE="iso"; $env:PACKER_BUILD_NAME="vmware"; `,
		`$env:BAR="foo"; $env:FOO="bar"; $env:PACKER_BUILDER_TYPE="iso"; $env:PACKER_BUILD_NAME="vmware"; `,
		`$env:BAR="foo"; $env:BAZ="qux"; $env:FOO="bar"; $env:PACKER_BUILDER_TYPE="iso"; $env:PACKER_BUILD_NAME="vmware"; $env:YAR="yaa"; `,
		`$env:BAR="foo=yaa"; $env:FOO="bar=baz"; $env:PACKER_BUILDER_TYPE="iso"; $env:PACKER_BUILD_NAME="vmware"; `,
		`$env:BAR="=foo"; $env:FOO="=bar"; $env:PACKER_BUILDER_TYPE="iso"; $env:PACKER_BUILD_NAME="vmware"; `,
		"$env:BAR=\"foo`$yaa\"; $env:FOO=\"bar`$baz\"; $env:PACKER_BUILDER_TYPE=\"iso\"; $env:PACKER_BUILD_NAME=\"vmware\"; ",
		"$env:BAR=\"foo`\"yaa\"; $env:FOO=\"bar`\"baz\"; $env:PACKER_BUILDER_TYPE=\"iso\"; $env:PACKER_BUILD_NAME=\"vmware\"; ",
		"$env:BAR=\"foo`'yaa\"; $env:FOO=\"bar`'baz\"; $env:PACKER_BUILDER_TYPE=\"iso\"; $env:PACKER_BUILD_NAME=\"vmware\"; ",
		"$env:BAR=\"foo``yaa\"; $env:FOO=\"bar``baz\"; $env:PACKER_BUILDER_TYPE=\"iso\"; $env:PACKER_BUILD_NAME=\"vmware\"; ",
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
		flattenedEnvVars = p.createFlattenedEnvVars(false)
		if flattenedEnvVars != expectedValue {
			t.Fatalf("expected flattened env vars to be: %s, got %s.", expectedValue, flattenedEnvVars)
		}
	}
}

func TestProvision_createCommandText(t *testing.T) {
	config := testConfig()
	config["remote_path"] = "c:/Windows/Temp/script.ps1"
	p := new(Provisioner)
	comm := new(packersdk.MockCommunicator)
	p.communicator = comm
	_ = p.Prepare(config)

	// Defaults provided by Packer
	p.config.PackerBuildName = "vmware"
	p.config.PackerBuilderType = "iso"

	// Non-elevated
	p.generatedData = make(map[string]interface{})
	cmd, _ := p.createCommandText()

	re := regexp.MustCompile(`powershell -executionpolicy bypass "& { if \(Test-Path variable:global:ProgressPreference\){set-variable -name variable:global:ProgressPreference -value 'SilentlyContinue'};\. c:/Windows/Temp/packer-ps-env-vars-[[:alnum:]]{8}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{12}\.ps1; &'c:/Windows/Temp/script.ps1'; exit \$LastExitCode }"`)
	matched := re.MatchString(cmd)
	if !matched {
		t.Fatalf("Got unexpected command: %s", cmd)
	}

	// Elevated
	p.config.ElevatedUser = "vagrant"
	p.config.ElevatedPassword = "vagrant"
	cmd, _ = p.createCommandText()
	re = regexp.MustCompile(`powershell -executionpolicy bypass -file "C:/Windows/Temp/packer-elevated-shell-[[:alnum:]]{8}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{12}\.ps1"`)
	matched = re.MatchString(cmd)
	if !matched {
		t.Fatalf("Got unexpected elevated command: %s", cmd)
	}
}

func TestProvision_uploadEnvVars(t *testing.T) {
	p := new(Provisioner)
	comm := new(packersdk.MockCommunicator)
	p.communicator = comm

	flattenedEnvVars := `$env:PACKER_BUILDER_TYPE="footype"; $env:PACKER_BUILD_NAME="foobuild";`

	err := p.uploadEnvVars(flattenedEnvVars)
	if err != nil {
		t.Fatalf("Did not expect error: %s", err.Error())
	}

	if comm.UploadCalled != true {
		t.Fatalf("Failed to upload env var file")
	}
}

func TestCancel(t *testing.T) {
	// Don't actually call Cancel() as it performs an os.Exit(0)
	// which kills the 'go test' tool
}

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"inline": []interface{}{"foo", "bar"},
	}
}

func testConfigWithSkipClean() map[string]interface{} {
	return map[string]interface{}{
		"inline":     []interface{}{"foo", "bar"},
		"skip_clean": true,
	}
}

func generatedData() map[string]interface{} {
	return map[string]interface{}{
		"PackerHTTPAddr": commonsteps.HttpAddrNotImplemented,
		"PackerHTTPIP":   commonsteps.HttpIPNotImplemented,
		"PackerHTTPPort": commonsteps.HttpPortNotImplemented,
	}
}
