package packer

import (
	"cgl.tideland.biz/asserts"
	"os"
	"testing"
)

func TestEnvironment_DefaultCli_Empty(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	defaultEnv := NewEnvironment()

	assert.Equal(defaultEnv.Cli([]string{}), 1, "CLI with no args")
}

func TestEnvironment_DefaultCli_Version(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	defaultEnv := NewEnvironment()

	// Test the basic version options
	assert.Equal(defaultEnv.Cli([]string{"version"}), 0, "version should work")
	assert.Equal(defaultEnv.Cli([]string{"--version"}), 0, "--version should work")
	assert.Equal(defaultEnv.Cli([]string{"-v"}), 0, "-v should work")

	// Test the --version and -v can appear anywhere
	assert.Equal(defaultEnv.Cli([]string{"bad", "-v"}), 0, "-v should work anywhere")
	assert.Equal(defaultEnv.Cli([]string{"bad", "--version"}), 0, "--version should work anywhere")

	// Test that "version" can't appear anywhere
	assert.Equal(defaultEnv.Cli([]string{"bad", "version"}), 1, "version should NOT work anywhere")
}

func TestEnvironment_DefaultUi(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	defaultEnv := NewEnvironment()
	assert.NotNil(defaultEnv.Ui(), "default UI should not be nil")

	rwUi, ok := defaultEnv.Ui().(*ReaderWriterUi)
	assert.True(ok, "default UI should be ReaderWriterUi")
	assert.Equal(rwUi.Writer, os.Stdout, "default UI should go to stdout")
	assert.Equal(rwUi.Reader, os.Stdin, "default UI should read from stdin")
}
