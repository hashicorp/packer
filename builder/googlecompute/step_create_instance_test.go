package googlecompute

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/stretchr/testify/assert"
)

func TestStepCreateInstance_impl(t *testing.T) {
	var _ multistep.Step = new(StepCreateInstance)
}

func TestStepCreateInstance(t *testing.T) {
	state := testState(t)
	step := new(StepCreateInstance)
	defer step.Cleanup(state)

	state.Put("ssh_public_key", "key")

	c := state.Get("config").(*Config)
	d := state.Get("driver").(*DriverMock)
	d.GetImageResult = StubImage("test-image", "test-project", []string{}, 100)

	// run the step
	assert.Equal(t, step.Run(context.Background(), state), multistep.ActionContinue, "Step should have passed and continued.")

	// Verify state
	nameRaw, ok := state.GetOk("instance_name")
	assert.True(t, ok, "State should have an instance name.")

	// cleanup
	step.Cleanup(state)

	// Check args passed to the driver.
	assert.Equal(t, d.DeleteInstanceName, nameRaw.(string), "Incorrect instance name passed to driver.")
	assert.Equal(t, d.DeleteInstanceZone, c.Zone, "Incorrect instance zone passed to driver.")
	assert.Equal(t, d.DeleteDiskName, c.InstanceName, "Incorrect disk name passed to driver.")
	assert.Equal(t, d.DeleteDiskZone, c.Zone, "Incorrect disk zone passed to driver.")
}

func TestStepCreateInstance_fromFamily(t *testing.T) {
	cases := []struct {
		Name   string
		Family string
		Expect bool
	}{
		{"test-image", "", false},
		{"test-image", "test-family", false}, // name trumps family
		{"", "test-family", true},
	}

	for _, tc := range cases {
		state := testState(t)
		step := new(StepCreateInstance)
		defer step.Cleanup(state)

		state.Put("ssh_public_key", "key")

		c := state.Get("config").(*Config)
		c.SourceImage = tc.Name
		c.SourceImageFamily = tc.Family
		d := state.Get("driver").(*DriverMock)
		d.GetImageResult = StubImage("test-image", "test-project", []string{}, 100)

		// run the step
		assert.Equal(t, step.Run(context.Background(), state), multistep.ActionContinue, "Step should have passed and continued.")

		// cleanup
		step.Cleanup(state)

		// Check args passed to the driver.
		if tc.Expect {
			assert.True(t, d.GetImageFromFamily, "Driver wasn't instructed to use an image family")
		} else {
			assert.False(t, d.GetImageFromFamily, "Driver was unexpectedly instructed to use an image family")
		}
	}
}

func TestStepCreateInstance_windowsNeedsPassword(t *testing.T) {

	state := testState(t)
	step := new(StepCreateInstance)
	defer step.Cleanup(state)

	state.Put("ssh_public_key", "key")
	c := state.Get("config").(*Config)
	d := state.Get("driver").(*DriverMock)
	d.GetImageResult = StubImage("test-image", "test-project", []string{"windows"}, 100)
	c.Comm.Type = "winrm"
	// run the step
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify state
	nameRaw, ok := state.GetOk("instance_name")
	if !ok {
		t.Fatal("should have instance name")
	}

	createPassword, ok := state.GetOk("create_windows_password")

	if !ok || !createPassword.(bool) {
		t.Fatal("should need to create a windows password")
	}

	// cleanup
	step.Cleanup(state)

	if d.DeleteInstanceName != nameRaw.(string) {
		t.Fatal("should've deleted instance")
	}
	if d.DeleteInstanceZone != c.Zone {
		t.Fatalf("bad instance zone: %#v", d.DeleteInstanceZone)
	}

	if d.DeleteDiskName != c.InstanceName {
		t.Fatal("should've deleted disk")
	}
	if d.DeleteDiskZone != c.Zone {
		t.Fatalf("bad disk zone: %#v", d.DeleteDiskZone)
	}
}

func TestStepCreateInstance_windowsPasswordSet(t *testing.T) {

	state := testState(t)
	step := new(StepCreateInstance)
	defer step.Cleanup(state)

	state.Put("ssh_public_key", "key")

	config := state.Get("config").(*Config)
	driver := state.Get("driver").(*DriverMock)
	driver.GetImageResult = StubImage("test-image", "test-project", []string{"windows"}, 100)
	config.Comm.Type = "winrm"
	config.Comm.WinRMPassword = "password"

	// run the step
	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}

	// Verify state
	nameRaw, ok := state.GetOk("instance_name")
	if !ok {
		t.Fatal("should have instance name")
	}

	_, ok = state.GetOk("create_windows_password")

	if ok {
		t.Fatal("should not need to create windows password")
	}

	// cleanup
	step.Cleanup(state)

	if driver.DeleteInstanceName != nameRaw.(string) {
		t.Fatal("should've deleted instance")
	}
	if driver.DeleteInstanceZone != config.Zone {
		t.Fatalf("bad instance zone: %#v", driver.DeleteInstanceZone)
	}

	if driver.DeleteDiskName != config.InstanceName {
		t.Fatal("should've deleted disk")
	}
	if driver.DeleteDiskZone != config.Zone {
		t.Fatalf("bad disk zone: %#v", driver.DeleteDiskZone)
	}
}

func TestStepCreateInstance_error(t *testing.T) {
	state := testState(t)
	step := new(StepCreateInstance)
	defer step.Cleanup(state)

	state.Put("ssh_public_key", "key")

	d := state.Get("driver").(*DriverMock)
	d.RunInstanceErr = errors.New("error")
	d.GetImageResult = StubImage("test-image", "test-project", []string{}, 100)

	// run the step
	assert.Equal(t, step.Run(context.Background(), state), multistep.ActionHalt, "Step should have failed and halted.")

	// Verify state
	_, ok := state.GetOk("error")
	assert.True(t, ok, "State should have an error.")
	_, ok = state.GetOk("instance_name")
	assert.False(t, ok, "State should not have an instance name.")
}

func TestStepCreateInstance_errorOnChannel(t *testing.T) {
	state := testState(t)
	step := new(StepCreateInstance)
	defer step.Cleanup(state)

	state.Put("ssh_public_key", "key")

	errCh := make(chan error, 1)
	errCh <- errors.New("error")

	d := state.Get("driver").(*DriverMock)
	d.RunInstanceErrCh = errCh
	d.GetImageResult = StubImage("test-image", "test-project", []string{}, 100)

	// run the step
	assert.Equal(t, step.Run(context.Background(), state), multistep.ActionHalt, "Step should have failed and halted.")

	// Verify state
	_, ok := state.GetOk("error")
	assert.True(t, ok, "State should have an error.")
	_, ok = state.GetOk("instance_name")
	assert.False(t, ok, "State should not have an instance name.")
}

func TestStepCreateInstance_errorTimeout(t *testing.T) {
	state := testState(t)
	step := new(StepCreateInstance)
	defer step.Cleanup(state)

	state.Put("ssh_public_key", "key")

	errCh := make(chan error, 1)

	config := state.Get("config").(*Config)
	config.StateTimeout = 1 * time.Millisecond

	d := state.Get("driver").(*DriverMock)
	d.RunInstanceErrCh = errCh
	d.GetImageResult = StubImage("test-image", "test-project", []string{}, 100)

	// run the step
	assert.Equal(t, step.Run(context.Background(), state), multistep.ActionHalt, "Step should have failed and halted.")

	// Verify state
	_, ok := state.GetOk("error")
	assert.True(t, ok, "State should have an error.")
	_, ok = state.GetOk("instance_name")
	assert.False(t, ok, "State should not have an instance name.")
}

func TestStepCreateInstance_noServiceAccount(t *testing.T) {
	state := testState(t)
	step := new(StepCreateInstance)
	defer step.Cleanup(state)

	state.Put("ssh_public_key", "key")

	c := state.Get("config").(*Config)
	c.DisableDefaultServiceAccount = true
	c.ServiceAccountEmail = ""
	d := state.Get("driver").(*DriverMock)
	d.GetImageResult = StubImage("test-image", "test-project", []string{}, 100)

	// run the step
	assert.Equal(t, step.Run(context.Background(), state), multistep.ActionContinue, "Step should have passed and continued.")

	// cleanup
	step.Cleanup(state)

	// Check args passed to the driver.
	assert.Equal(t, d.RunInstanceConfig.DisableDefaultServiceAccount, c.DisableDefaultServiceAccount, "Incorrect value for DisableDefaultServiceAccount passed to driver.")
	assert.Equal(t, d.RunInstanceConfig.ServiceAccountEmail, c.ServiceAccountEmail, "Incorrect value for ServiceAccountEmail passed to driver.")
}

func TestStepCreateInstance_customServiceAccount(t *testing.T) {
	state := testState(t)
	step := new(StepCreateInstance)
	defer step.Cleanup(state)

	state.Put("ssh_public_key", "key")

	c := state.Get("config").(*Config)
	c.DisableDefaultServiceAccount = true
	c.ServiceAccountEmail = "custom-service-account"
	d := state.Get("driver").(*DriverMock)
	d.GetImageResult = StubImage("test-image", "test-project", []string{}, 100)

	// run the step
	assert.Equal(t, step.Run(context.Background(), state), multistep.ActionContinue, "Step should have passed and continued.")

	// cleanup
	step.Cleanup(state)

	// Check args passed to the driver.
	assert.Equal(t, d.RunInstanceConfig.DisableDefaultServiceAccount, c.DisableDefaultServiceAccount, "Incorrect value for DisableDefaultServiceAccount passed to driver.")
	assert.Equal(t, d.RunInstanceConfig.ServiceAccountEmail, c.ServiceAccountEmail, "Incorrect value for ServiceAccountEmail passed to driver.")
}

func TestCreateInstanceMetadata(t *testing.T) {
	state := testState(t)
	c := state.Get("config").(*Config)
	image := StubImage("test-image", "test-project", []string{}, 100)
	key := "abcdefgh12345678"

	// create our metadata
	_, metadataSSHKeys, err := c.createInstanceMetadata(image, key)

	assert.True(t, err == nil, "Metadata creation should have succeeded.")

	// ensure our key is listed
	assert.True(t, strings.Contains(metadataSSHKeys["ssh-keys"], key), "Instance metadata should contain provided key")
}

func TestCreateInstanceMetadata_noPublicKey(t *testing.T) {
	state := testState(t)
	c := state.Get("config").(*Config)
	image := StubImage("test-image", "test-project", []string{}, 100)
	sshKeys := c.Metadata["ssh-keys"]

	// create our metadata
	_, metadataSSHKeys, err := c.createInstanceMetadata(image, "")

	assert.True(t, err == nil, "Metadata creation should have succeeded.")

	// ensure the ssh metadata hasn't changed
	assert.Equal(t, metadataSSHKeys["ssh-keys"], sshKeys, "Instance metadata should not have been modified")
}

func TestCreateInstanceMetadata_metadataFile(t *testing.T) {
	state := testState(t)
	c := state.Get("config").(*Config)
	image := StubImage("test-image", "test-project", []string{}, 100)
	content := testMetadataFileContent
	fileName := testMetadataFile(t)
	c.MetadataFiles["user-data"] = fileName

	// create our metadata
	metadataNoSSHKeys, _, err := c.createInstanceMetadata(image, "")

	assert.True(t, err == nil, "Metadata creation should have succeeded.")

	// ensure the user-data key in metadata is updated with file content
	assert.Equal(t, metadataNoSSHKeys["user-data"], content, "user-data field of the instance metadata should have been updated.")
}

func TestCreateInstanceMetadata_withWrapStartupScript(t *testing.T) {
	tt := []struct {
		WrapStartupScript            config.Trilean
		StartupScriptContents        string
		WrappedStartupScriptContents string
		WrappedStartupScriptStatus   string
	}{
		{
			WrapStartupScript:     config.TriUnset,
			StartupScriptContents: testMetadataFileContent,
		},
		{
			WrapStartupScript:     config.TriFalse,
			StartupScriptContents: testMetadataFileContent,
		},
		{
			WrapStartupScript:            config.TriTrue,
			StartupScriptContents:        StartupScriptLinux,
			WrappedStartupScriptContents: testMetadataFileContent,
			WrappedStartupScriptStatus:   StartupScriptStatusNotDone,
		},
	}

	for _, tc := range tt {
		tc := tc
		state := testState(t)
		image := StubImage("test-image", "test-project", []string{}, 100)
		c := state.Get("config").(*Config)
		c.StartupScriptFile = testMetadataFile(t)
		c.WrapStartupScriptFile = tc.WrapStartupScript

		// create our metadata
		metadataNoSSHKeys, _, err := c.createInstanceMetadata(image, "")

		assert.True(t, err == nil, "Metadata creation should have succeeded.")
		assert.Equal(t, tc.StartupScriptContents, metadataNoSSHKeys[StartupScriptKey], fmt.Sprintf("Instance metadata for startup script should be %q.", tc.StartupScriptContents))
		assert.Equal(t, tc.WrappedStartupScriptContents, metadataNoSSHKeys[StartupWrappedScriptKey], fmt.Sprintf("Instance metadata for wrapped startup script should be %q.", tc.WrappedStartupScriptContents))
		assert.Equal(t, tc.WrappedStartupScriptStatus, metadataNoSSHKeys[StartupScriptStatusKey], fmt.Sprintf("Instance metadata startup script status should be %q.", tc.WrappedStartupScriptStatus))
	}
}

func TestCreateInstanceMetadataWaitToAddSSHKeys(t *testing.T) {
	state := testState(t)
	c := state.Get("config").(*Config)
	image := StubImage("test-image", "test-project", []string{}, 100)
	key := "abcdefgh12345678"

	var waitTime int = 4
	c.WaitToAddSSHKeys = time.Duration(waitTime) * time.Second
	c.Metadata = map[string]string{
		"metadatakey1": "xyz",
		"metadatakey2": "123",
	}

	// create our metadata
	metadataNoSSHKeys, metadataSSHKeys, err := c.createInstanceMetadata(image, key)

	assert.True(t, err == nil, "Metadata creation should have succeeded.")

	// ensure our metadata is listed
	assert.True(t, strings.Contains(metadataSSHKeys["ssh-keys"], key), "Instance metadata should contain provided SSH key")
	assert.True(t, strings.Contains(metadataNoSSHKeys["metadatakey1"], "xyz"), "Instance metadata should contain provided key: metadatakey1")
	assert.True(t, strings.Contains(metadataNoSSHKeys["metadatakey2"], "123"), "Instance metadata should contain provided key: metadatakey2")
}

func TestStepCreateInstanceWaitToAddSSHKeys(t *testing.T) {
	state := testState(t)
	step := new(StepCreateInstance)
	defer step.Cleanup(state)

	state.Put("ssh_public_key", "key")

	c := state.Get("config").(*Config)
	d := state.Get("driver").(*DriverMock)
	d.GetImageResult = StubImage("test-image", "test-project", []string{}, 100)

	key := "abcdefgh12345678"

	var waitTime int = 5
	c.WaitToAddSSHKeys = time.Duration(waitTime) * time.Second
	c.Comm.SSHPublicKey = []byte(key)

	c.Metadata = map[string]string{
		"metadatakey1": "xyz",
		"metadatakey2": "123",
	}

	// run the step
	assert.Equal(t, step.Run(context.Background(), state), multistep.ActionContinue, "Step should have passed and continued.")

	// Verify state
	_, ok := state.GetOk("instance_name")
	assert.True(t, ok, "State should have an instance name.")

	// cleanup
	step.Cleanup(state)
}

func TestStepCreateInstanceNoWaitToAddSSHKeys(t *testing.T) {
	state := testState(t)
	step := new(StepCreateInstance)
	defer step.Cleanup(state)

	state.Put("ssh_public_key", "key")

	c := state.Get("config").(*Config)
	d := state.Get("driver").(*DriverMock)
	d.GetImageResult = StubImage("test-image", "test-project", []string{}, 100)

	key := "abcdefgh12345678"

	c.Comm.SSHPublicKey = []byte(key)

	c.Metadata = map[string]string{
		"metadatakey1": "xyz",
		"metadatakey2": "123",
	}

	// run the step
	assert.Equal(t, step.Run(context.Background(), state), multistep.ActionContinue, "Step should have passed and continued.")

	// Verify state
	_, ok := state.GetOk("instance_name")
	assert.True(t, ok, "State should have an instance name.")

	// cleanup
	step.Cleanup(state)
}
