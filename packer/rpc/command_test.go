package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
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
	server := rpc.NewServer()
	RegisterCommand(server, command)
	address := serveSingleConn(server)

	// Create the command client over RPC and run some methods to verify
	// we get the proper behavior.
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	clientComm := Command(client)

	//Test Help
	help := clientComm.Help()
	if help != "bar" {
		t.Fatalf("bad: %s", help)
	}

	// Test run
	runArgs := []string{"foo", "bar"}
	testEnv := &testEnvironment{}
	exitCode := clientComm.Run(testEnv, runArgs)
	if !reflect.DeepEqual(command.runArgs, runArgs) {
		t.Fatalf("bad: %#v", command.runArgs)
	}
	if exitCode != 0 {
		t.Fatalf("bad: %d", exitCode)
	}

	if command.runEnv == nil {
		t.Fatal("runEnv should not be nil")
	}

	command.runEnv.Ui()
	if !testEnv.uiCalled {
		t.Fatal("ui should be called")
	}

	// Test Synopsis
	synopsis := clientComm.Synopsis()
	if synopsis != "foo" {
		t.Fatalf("bad: %#v", synopsis)
	}
}

func TestCommand_Implements(t *testing.T) {
	var _ packer.Command = Command(nil)
}
