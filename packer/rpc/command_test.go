package rpc

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"net"
	"net/rpc"
	"testing"
)

type TestCommand struct {
	runArgs []string
	runCalled bool
	runEnv packer.Environment
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

// This starts a RPC server for the given command listening on the
// given address. The RPC server is ready when "readyChan" receives a message
// and the RPC server will quit when "stopChan" receives a message.
//
// This function should be run in a goroutine.
func testCommandRPCServer(laddr string, command interface{}, readyChan chan int, stopChan <-chan int) {
	listener, err := net.Listen("tcp", laddr)
	if err != nil {
		panic(err)
	}

	// Close the listener when we exit so that the RPC server ends
	defer listener.Close()

	// Start the RPC server
	server := rpc.NewServer()
	server.RegisterName("Command", command)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				// If there is an error, just ignore it.
				break
			}

			go server.ServeConn(conn)
		}
	}()

	// We're ready!
	readyChan <- 1

	// Block on waiting to receive from the channel
	<-stopChan
}

func TestRPCCommand(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	// Create the command
	command := new(TestCommand)
	serverCommand := &ServerCommand{command}

	// Start the RPC server, and make sure to exit it at the end
	// of the test.
	readyChan := make(chan int)
	stopChan := make(chan int)
	defer func() { stopChan <- 1 }()
	go testCommandRPCServer(":1234", serverCommand, readyChan, stopChan)
	<-readyChan

	// Create the command client over RPC and run some methods to verify
	// we get the proper behavior.
	client, err := rpc.Dial("tcp", ":1234")
	if err != nil {
		panic(err)
	}

	clientComm := &ClientCommand{client}
	runArgs := []string{"foo", "bar"}
	testEnv := &testEnvironment{}
	exitCode := clientComm.Run(testEnv, runArgs)
	synopsis := clientComm.Synopsis()

	assert.Equal(command.runArgs, runArgs, "Correct args should be sent")
	assert.Equal(exitCode, 0, "Exit code should be correct")
	assert.Equal(synopsis, "foo", "Synopsis should be correct")
}
