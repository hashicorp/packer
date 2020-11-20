package yandex

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testMetadataFileContent = `meta data value`

func testMetadataFile(t *testing.T) string {
	tf, err := ioutil.TempFile("", "packer")
	require.NoErrorf(t, err, "create temporary file failed")
	defer tf.Close()

	_, err = tf.Write([]byte(testMetadataFileContent))
	require.NoErrorf(t, err, "write to file failed")

	return tf.Name()
}

func TestCreateInstanceMetadata(t *testing.T) {
	state := testState(t)
	c := state.Get("config").(*Config)
	pubKey := "abcdefgh123456789"

	// create our metadata
	metadata, err := c.createInstanceMetadata(pubKey)
	require.NoError(t, err, "Metadata creation should have succeeded.")

	// ensure our pubKey is listed
	assert.True(t, strings.Contains(metadata["ssh-keys"], pubKey), "Instance metadata should contain provided public key")
}

func TestCreateInstanceMetadata_noPublicKey(t *testing.T) {
	state := testState(t)
	c := state.Get("config").(*Config)
	sshKeys := c.Metadata["sshKeys"]

	// create our metadata
	metadata, err := c.createInstanceMetadata("")

	// ensure the metadata created without err
	require.NoError(t, err, "Metadata creation should have succeeded.")

	// ensure the ssh metadata hasn't changed
	assert.Equal(t, metadata["sshKeys"], sshKeys, "Instance metadata should not have been modified")
}

func TestCreateInstanceMetadata_fromFile(t *testing.T) {
	state := testState(t)
	metadataFile := testMetadataFile(t)
	defer os.Remove(metadataFile)

	state.Put("config", testConfigStruct(t))
	c := state.Get("config").(*Config)
	c.MetadataFromFile = map[string]string{
		"test-key": metadataFile,
	}

	// create our metadata
	metadata, err := c.createInstanceMetadata("")
	require.NoError(t, err, "Metadata creation should have succeeded.")

	// ensure the metadata from file hasn't changed
	assert.Equal(t, testMetadataFileContent, metadata["test-key"], "Instance metadata should not have been modified")
}

func TestCreateInstanceMetadata_fromFileAndTemplate(t *testing.T) {
	state := testState(t)
	metadataFile := testMetadataFile(t)
	defer os.Remove(metadataFile)

	state.Put("config", testConfigStruct(t))
	c := state.Get("config").(*Config)
	c.MetadataFromFile = map[string]string{
		"test-key": metadataFile,
	}
	c.Metadata = map[string]string{
		"test-key":   "override value",
		"test-key-2": "new-value",
	}

	// create our metadata
	metadata, err := c.createInstanceMetadata("")
	require.NoError(t, err, "Metadata creation should have succeeded.")

	// ensure the metadata merged
	assert.Equal(t, "override value", metadata["test-key"], "Instance metadata should not have been modified")
	assert.Equal(t, "new-value", metadata["test-key-2"], "Instance metadata should not have been modified")
}

func TestCreateInstanceMetadata_fromNotExistFile(t *testing.T) {
	state := testState(t)
	metadataFile := "not-exist-file"

	state.Put("config", testConfigStruct(t))
	c := state.Get("config").(*Config)
	c.MetadataFromFile = map[string]string{
		"test-key": metadataFile,
	}

	// create our metadata
	_, err := c.createInstanceMetadata("")

	assert.True(t, err != nil, "Metadata creation should have an error.")
}

func testState(t *testing.T) multistep.StateBag {
	state := new(multistep.BasicStateBag)
	state.Put("config", testConfigStruct(t))
	state.Put("hook", &packersdk.MockHook{})
	state.Put("ui", &packersdk.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	return state
}
