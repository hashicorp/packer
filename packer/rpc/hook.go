package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
)

// An implementation of packer.Hook where the hook is actually executed
// over an RPC connection.
type hook struct {
	client *rpc.Client
}

// HookServer wraps a packer.Hook implementation and makes it exportable
// as part of a Golang RPC server.
type HookServer struct {
	hook packer.Hook
}

type HookRunArgs struct {
	Name       string
	Data       interface{}
	RPCAddress string
}

func Hook(client *rpc.Client) *hook {
	return &hook{client}
}

func (h *hook) Run(name string, ui packer.Ui, comm packer.Communicator, data interface{}) error {
	server := rpc.NewServer()
	RegisterCommunicator(server, comm)
	RegisterUi(server, ui)
	address := serveSingleConn(server)

	args := &HookRunArgs{name, data, address}
	return h.client.Call("Hook.Run", args, new(interface{}))
}

func (h *HookServer) Run(args *HookRunArgs, reply *interface{}) error {
	client, err := rpc.Dial("tcp", args.RPCAddress)
	if err != nil {
		return err
	}

	if err := h.hook.Run(args.Name, &Ui{client}, Communicator(client), args.Data); err != nil {
		return NewBasicError(err)
	}

	*reply = nil
	return nil
}
