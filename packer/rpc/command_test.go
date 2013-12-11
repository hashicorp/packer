package rpc

import (
	"github.com/mitchellh/packer/packer"
	"reflect"
	"testing"
)

type TestCommand struct {
	runArgs   []string
	runCalled bool
	runEnv    packer.Environment
}

func (tc *TestCommand) Help() string {
	return "bar"
}

func (tc *TestCommand) Run(env packer.Environment, args []string) int {
	tc.runCalled = true
	tc.runArgs = args
	tc.runEnv = env
	return 0
}

func (tc *TestCommand) Synopsis() string {
	return "foo"
}

func TestRPCCommand(t *testing.T) {
	// Create the command
	command := new(TestCommand)

	// Start the server
	client, server := testClientServer(t)
	defer client.Close()
	defer server.Close()
	server.RegisterCommand(command)
	commClient := client.Command()

	//Test Help
	help := commClient.Help()
	if help != "bar" {
		t.Fatalf("bad: %s", help)
	}

	// Test run
	runArgs := []string{"foo", "bar"}
	testEnv := &testEnvironment{}
	exitCode := commClient.Run(testEnv, runArgs)
	if !reflect.DeepEqual(command.runArgs, runArgs) {
		t.Fatalf("bad: %#v", command.runArgs)
	}
	if exitCode != 0 {
		t.Fatalf("bad: %d", exitCode)
	}

	if command.runEnv == nil {
		t.Fatal("runEnv should not be nil")
	}

	// Test Synopsis
	synopsis := commClient.Synopsis()
	if synopsis != "foo" {
		t.Fatalf("bad: %#v", synopsis)
	}
}

func TestCommand_Implements(t *testing.T) {
	var _ packer.Command = new(command)
}
