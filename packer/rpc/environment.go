package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
)

// A EnvironmentClient is an implementation of the packer.Environment interface
// where the actual environment is executed over an RPC connection.
type EnvironmentClient struct {
	client *rpc.Client
}

// A EnvironmentServer wraps a packer.Environment and makes it exportable
// as part of a Golang RPC server.
type EnvironmentServer struct {
	env packer.Environment
}
