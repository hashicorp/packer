package packer

import (
	"cgl.tideland.biz/asserts"
	"testing"
)

func TestBuildCommand_Run_NoArgs(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	command := new(buildCommand)
	result := command.Run(testEnvironment(), make([]string, 0))
	assert.Equal(result, 1, "no args should error")
}

func TestBuildCommand_Run_MoreThanOneArg(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	command := new(buildCommand)

	args := []string{"one", "two"}
	result := command.Run(testEnvironment(), args)
	assert.Equal(result, 1, "More than one arg should fail")
}

func TestBuildCommand_Run_MissingFile(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	command := new(buildCommand)

	args := []string{"i-better-not-exist"}
	result := command.Run(testEnvironment(), args)
	assert.Equal(result, 1, "a non-existent file should error")
}
