package powershell

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	//"log"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/packer/packer"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"inline": []interface{}{"foo", "bar"},
	}
}

func init() {
	//log.SetOutput(ioutil.Discard)
}

func TestProvisionerPrepare_extractScript(t *testing.T) {
	config := testConfig()
	p := new(Provisioner)
	_ = p.Prepare(config)
	file, err := extractScript(p)
	if err != nil {
		t.Fatalf("Should not be error: %s", err)
	}
	t.Logf("File: %s", file)
	if strings.Index(file, os.TempDir()) != 0 {
		t.Fatalf("Temp file should reside in %s. File location: %s", os.TempDir(), file)
	}

	// File contents should contain 2 lines concatenated by newlines: foo\nbar
	readFile, err := ioutil.ReadFile(file)
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
	if _, ok := raw.(packer.Provisioner); !ok {
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

	if p.config.ExecuteCommand != `powershell -executionpolicy bypass "& { if (Test-Path variable:global:ProgressPreference){$ProgressPreference='SilentlyContinue'};. {{.Vars}}; &'{{.Path}}';exit $LastExitCode }"` {
		t.Fatalf(`Default command should be 'powershell -executionpolicy bypass "& { if (Test-Path variable:global:ProgressPreference){$ProgressPreference='SilentlyContinue'};. {{.Vars}}; &'{{.Path}}';exit $LastExitCode }"', but got '%s'`, p.config.ExecuteCommand)
	}

	if p.config.ElevatedExecuteCommand != `powershell -executionpolicy bypass "& { if (Test-Path variable:global:ProgressPreference){$ProgressPreference='SilentlyContinue'};. {{.Vars}}; &'{{.Path}}'; exit $LastExitCode }"` {
		t.Fatalf(`Default command should be 'powershell -executionpolicy bypass "& { if (Test-Path variable:global:ProgressPreference){$ProgressPreference='SilentlyContinue'};. {{.Vars}}; &'{{.Path}}'; exit $LastExitCode }"', but got '%s'`, p.config.ElevatedExecuteCommand)
	}

	if p.config.ValidExitCodes == nil {
		t.Fatalf("ValidExitCodes should not be nil")
	}
	if p.config.ValidExitCodes != nil {
		expCodes := []int{0}
		for i, v := range p.config.ValidExitCodes {
			if v != expCodes[i] {
				t.Fatalf("Expected ValidExitCodes don't match actual")
			}
		}
	}

	if p.config.ElevatedEnvVarFormat != `$env:%s="%s"; ` {
		t.Fatalf(`Default command should be powershell '$env:%%s="%%s"; ', but got %s`, p.config.ElevatedEnvVarFormat)
	}
}

func TestProvisionerPrepare_Config(t *testing.T) {
	config := testConfig()
	config["elevated_user"] = "{{user `user`}}"
	config["elevated_password"] = "{{user `password`}}"
	config[packer.UserVariablesConfigKey] = map[string]string{
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

	if err == nil {
		t.Fatal("should have error (only provided elevated_user)")
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
	tf, err := ioutil.TempFile("", "packer")
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
	tf, err := ioutil.TempFile("", "packer")
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
	tf, err := ioutil.TempFile("", "packer")
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
	tf, err := ioutil.TempFile("", "packer")
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

func testUi() *packer.BasicUi {
	return &packer.BasicUi{
		Reader:      new(bytes.Buffer),
		Writer:      new(bytes.Buffer),
		ErrorWriter: new(bytes.Buffer),
	}
}

func testObjects() (packer.Ui, packer.Communicator) {
	ui := testUi()
	return ui, new(packer.MockCommunicator)
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
	comm := new(packer.MockCommunicator)
	comm.StartExitStatus = 200
	p.Prepare(config)
	err := p.Provision(ui, comm)
	if err != nil {
		t.Fatal("should not have error")
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
	comm := new(packer.MockCommunicator)
	comm.StartExitStatus = 201 // Invalid!
	p.Prepare(config)
	err := p.Provision(ui, comm)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestProvisionerProvision_Inline(t *testing.T) {
	config := testConfig()
	delete(config, "inline")

	// Defaults provided by Packer
	config["remote_path"] = "c:/Windows/Temp/inlineScript.ps1"
	config["inline"] = []string{"whoami"}
	ui := testUi()
	p := new(Provisioner)

	// Defaults provided by Packer - env vars should not appear in cmd
	p.config.PackerBuildName = "vmware"
	p.config.PackerBuilderType = "iso"
	comm := new(packer.MockCommunicator)
	p.Prepare(config)
	err := p.Provision(ui, comm)
	if err != nil {
		t.Fatal("should not have error")
	}

	cmd := comm.StartCmd.Command
	re := regexp.MustCompile(`powershell -executionpolicy bypass "& { if \(Test-Path variable:global:ProgressPreference\){\$ProgressPreference='SilentlyContinue'};\. \${env:SYSTEMROOT}/Temp/packer-env-vars-[[:alnum:]]{8}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{12}\.ps1; &'c:/Windows/Temp/inlineScript.ps1';exit \$LastExitCode }"`)
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
	err = p.Provision(ui, comm)
	if err != nil {
		t.Fatal("should not have error")
	}

	cmd = comm.StartCmd.Command
	re = regexp.MustCompile(`powershell -executionpolicy bypass "& { if \(Test-Path variable:global:ProgressPreference\){\$ProgressPreference='SilentlyContinue'};\. \${env:SYSTEMROOT}/Temp/packer-env-vars-[[:alnum:]]{8}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{12}\.ps1; &'c:/Windows/Temp/inlineScript.ps1';exit \$LastExitCode }"`)
	matched = re.MatchString(cmd)
	if !matched {
		t.Fatalf("Got unexpected command: %s", cmd)
	}
}

func TestProvisionerProvision_Scripts(t *testing.T) {
	tempFile, _ := ioutil.TempFile("", "packer")
	defer os.Remove(tempFile.Name())
	config := testConfig()
	delete(config, "inline")
	config["scripts"] = []string{tempFile.Name()}
	config["packer_build_name"] = "foobuild"
	config["packer_builder_type"] = "footype"
	config["remote_path"] = "c:/Windows/Temp/script.ps1"
	ui := testUi()

	p := new(Provisioner)
	comm := new(packer.MockCommunicator)
	p.Prepare(config)
	err := p.Provision(ui, comm)
	if err != nil {
		t.Fatal("should not have error")
	}

	cmd := comm.StartCmd.Command
	re := regexp.MustCompile(`powershell -executionpolicy bypass "& { if \(Test-Path variable:global:ProgressPreference\){\$ProgressPreference='SilentlyContinue'};\. \${env:SYSTEMROOT}/Temp/packer-env-vars-[[:alnum:]]{8}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{12}\.ps1; &'c:/Windows/Temp/script.ps1';exit \$LastExitCode }"`)
	matched := re.MatchString(cmd)
	if !matched {
		t.Fatalf("Got unexpected command: %s", cmd)
	}
}

func TestProvisionerProvision_ScriptsWithEnvVars(t *testing.T) {
	tempFile, _ := ioutil.TempFile("", "packer")
	config := testConfig()
	ui := testUi()
	defer os.Remove(tempFile.Name())
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
	comm := new(packer.MockCommunicator)
	p.Prepare(config)
	err := p.Provision(ui, comm)
	if err != nil {
		t.Fatal("should not have error")
	}

	cmd := comm.StartCmd.Command
	re := regexp.MustCompile(`powershell -executionpolicy bypass "& { if \(Test-Path variable:global:ProgressPreference\){\$ProgressPreference='SilentlyContinue'};\. \${env:SYSTEMROOT}/Temp/packer-env-vars-[[:alnum:]]{8}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{12}\.ps1; &'c:/Windows/Temp/script.ps1';exit \$LastExitCode }"`)
	matched := re.MatchString(cmd)
	if !matched {
		t.Fatalf("Got unexpected command: %s", cmd)
	}
}

func TestProvisionerProvision_UISlurp(t *testing.T) {
	// UI should be called n times

	// UI should receive following messages / output
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
	expected := []string{
		`$env:PACKER_BUILDER_TYPE="iso"; $env:PACKER_BUILD_NAME="vmware"; `,
		`$env:FOO="bar"; $env:PACKER_BUILDER_TYPE="iso"; $env:PACKER_BUILD_NAME="vmware"; `,
		`$env:BAZ="qux"; $env:FOO="bar"; $env:PACKER_BUILDER_TYPE="iso"; $env:PACKER_BUILD_NAME="vmware"; `,
		`$env:FOO="bar=baz"; $env:PACKER_BUILDER_TYPE="iso"; $env:PACKER_BUILD_NAME="vmware"; `,
		`$env:FOO="=bar"; $env:PACKER_BUILDER_TYPE="iso"; $env:PACKER_BUILD_NAME="vmware"; `,
		"$env:FOO=\"bar`$baz\"; $env:PACKER_BUILDER_TYPE=\"iso\"; $env:PACKER_BUILD_NAME=\"vmware\"; ",
		"$env:FOO=\"bar`\"baz\"; $env:PACKER_BUILDER_TYPE=\"iso\"; $env:PACKER_BUILD_NAME=\"vmware\"; ",
		"$env:FOO=\"bar`'baz\"; $env:PACKER_BUILDER_TYPE=\"iso\"; $env:PACKER_BUILD_NAME=\"vmware\"; ",
		"$env:FOO=\"bar``baz\"; $env:PACKER_BUILDER_TYPE=\"iso\"; $env:PACKER_BUILD_NAME=\"vmware\"; ",
	}

	p := new(Provisioner)
	p.Prepare(config)

	// Defaults provided by Packer
	p.config.PackerBuildName = "vmware"
	p.config.PackerBuilderType = "iso"

	for i, expectedValue := range expected {
		p.config.Vars = userEnvVarTests[i]
		flattenedEnvVars = p.createFlattenedEnvVars(true)
		if flattenedEnvVars != expectedValue {
			t.Fatalf("expected flattened env vars to be: %s, got %s.", expectedValue, flattenedEnvVars)
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
	expected := []string{
		`$env:PACKER_BUILDER_TYPE="iso"; $env:PACKER_BUILD_NAME="vmware"; `,
		`$env:FOO="bar"; $env:PACKER_BUILDER_TYPE="iso"; $env:PACKER_BUILD_NAME="vmware"; `,
		`$env:BAZ="qux"; $env:FOO="bar"; $env:PACKER_BUILDER_TYPE="iso"; $env:PACKER_BUILD_NAME="vmware"; `,
		`$env:FOO="bar=baz"; $env:PACKER_BUILDER_TYPE="iso"; $env:PACKER_BUILD_NAME="vmware"; `,
		`$env:FOO="=bar"; $env:PACKER_BUILDER_TYPE="iso"; $env:PACKER_BUILD_NAME="vmware"; `,
		"$env:FOO=\"bar`$baz\"; $env:PACKER_BUILDER_TYPE=\"iso\"; $env:PACKER_BUILD_NAME=\"vmware\"; ",
		"$env:FOO=\"bar`\"baz\"; $env:PACKER_BUILDER_TYPE=\"iso\"; $env:PACKER_BUILD_NAME=\"vmware\"; ",
		"$env:FOO=\"bar`'baz\"; $env:PACKER_BUILDER_TYPE=\"iso\"; $env:PACKER_BUILD_NAME=\"vmware\"; ",
		"$env:FOO=\"bar``baz\"; $env:PACKER_BUILDER_TYPE=\"iso\"; $env:PACKER_BUILD_NAME=\"vmware\"; ",
	}

	p := new(Provisioner)
	p.Prepare(config)

	// Defaults provided by Packer
	p.config.PackerBuildName = "vmware"
	p.config.PackerBuilderType = "iso"

	for i, expectedValue := range expected {
		p.config.Vars = userEnvVarTests[i]
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
	comm := new(packer.MockCommunicator)
	p.communicator = comm
	_ = p.Prepare(config)

	// Defaults provided by Packer
	p.config.PackerBuildName = "vmware"
	p.config.PackerBuilderType = "iso"

	// Non-elevated
	cmd, _ := p.createCommandText()

	re := regexp.MustCompile(`powershell -executionpolicy bypass "& { if \(Test-Path variable:global:ProgressPreference\){\$ProgressPreference='SilentlyContinue'};\. \${env:SYSTEMROOT}/Temp/packer-env-vars-[[:alnum:]]{8}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{12}\.ps1; &'c:/Windows/Temp/script.ps1';exit \$LastExitCode }"`)
	matched := re.MatchString(cmd)
	if !matched {
		t.Fatalf("Got unexpected command: %s", cmd)
	}

	// Elevated
	p.config.ElevatedUser = "vagrant"
	p.config.ElevatedPassword = "vagrant"
	cmd, _ = p.createCommandText()
	re = regexp.MustCompile(`powershell -executionpolicy bypass -file "%TEMP%/packer-elevated-shell-[[:alnum:]]{8}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{12}\.ps1"`)
	matched = re.MatchString(cmd)
	if !matched {
		t.Fatalf("Got unexpected elevated command: %s", cmd)
	}
}

func TestProvision_uploadEnvVars(t *testing.T) {
	p := new(Provisioner)
	comm := new(packer.MockCommunicator)
	p.communicator = comm

	flattenedEnvVars := `$env:PACKER_BUILDER_TYPE="footype"; $env:PACKER_BUILD_NAME="foobuild";`

	envVarPath, err := p.uploadEnvVars(flattenedEnvVars)
	if err != nil {
		t.Fatalf("Did not expect error: %s", err.Error())
	}

	if comm.UploadCalled != true {
		t.Fatalf("Failed to upload env var file")
	}

	re := regexp.MustCompile(`\${env:SYSTEMROOT}/Temp/packer-env-vars-[[:alnum:]]{8}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{4}-[[:alnum:]]{12}\.ps1`)
	matched := re.MatchString(envVarPath)
	if !matched {
		t.Fatalf("Got unexpected path for env var file: %s", envVarPath)
	}
}

func TestProvision_generateElevatedShellRunner(t *testing.T) {

	// Non-elevated
	config := testConfig()
	p := new(Provisioner)
	p.Prepare(config)
	comm := new(packer.MockCommunicator)
	p.communicator = comm
	path, err := p.generateElevatedRunner("whoami")

	if err != nil {
		t.Fatalf("Did not expect error: %s", err.Error())
	}

	if comm.UploadCalled != true {
		t.Fatalf("Should have uploaded file")
	}

	matched, _ := regexp.MatchString("%TEMP%(.{1})packer-elevated-shell.*", path)
	if !matched {
		t.Fatalf("Got unexpected file: %s", path)
	}
}

func TestRetryable(t *testing.T) {
	config := testConfig()

	count := 0
	retryMe := func() error {
		t.Logf("RetryMe, attempt number %d", count)
		if count == 2 {
			return nil
		}
		count++
		return errors.New(fmt.Sprintf("Still waiting %d more times...", 2-count))
	}
	retryableSleep = 50 * time.Millisecond
	p := new(Provisioner)
	p.config.StartRetryTimeout = 155 * time.Millisecond
	err := p.Prepare(config)
	err = p.retryable(retryMe)
	if err != nil {
		t.Fatalf("should not have error retrying funuction")
	}

	count = 0
	p.config.StartRetryTimeout = 10 * time.Millisecond
	err = p.Prepare(config)
	err = p.retryable(retryMe)
	if err == nil {
		t.Fatalf("should have error retrying funuction")
	}
}

func TestCancel(t *testing.T) {
	// Don't actually call Cancel() as it performs an os.Exit(0)
	// which kills the 'go test' tool
}
