package rpc

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"net/rpc"
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
	assert := asserts.NewTestingAsserts(t, true)

	// Create the command
	command := new(TestCommand)

	// Start the server
	server := rpc.NewServer()
	RegisterCommand(server, command)
	address := serveSingleConn(server)

	// Create the command client over RPC and run some methods to verify
	// we get the proper behavior.
	client, err := rpc.Dial("tcp", address)
	assert.Nil(err, "should be no error")

	clientComm := Command(client)

	//Test Help
	help := clientComm.Help()
	assert.Equal(help, "bar", "helps hould be correct")

	// Test run
	runArgs := []string{"foo", "bar"}
	testEnv := &testEnvironment{}
	exitCode := clientComm.Run(testEnv, runArgs)
	assert.Equal(command.runArgs, runArgs, "Correct args should be sent")
	assert.Equal(exitCode, 0, "Exit code should be correct")

	assert.NotNil(command.runEnv, "should have an env")
	if command.runEnv != nil {
		command.runEnv.Ui()
		assert.True(testEnv.uiCalled, "UI should be called on env")
	}

	// Test Synopsis
	synopsis := clientComm.Synopsis()
	assert.Equal(synopsis, "foo", "Synopsis should be correct")
}

func TestCommand_Implements(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var r packer.Command
	c := Command(nil)

	assert.Implementor(c, &r, "should be a Builder")
}
