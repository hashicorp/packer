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
	Name string
	Data interface{}
	RPCAddress string
}

func Hook(client *rpc.Client) *hook {
	return &hook{client}
}

func (h *hook) Run(name string, data interface{}, ui packer.Ui) {
	server := rpc.NewServer()
	RegisterUi(server, ui)
	address := serveSingleConn(server)

	args := &HookRunArgs{name, data, address}
	h.client.Call("Hook.Run", args, new(interface{}))
	return
}

func (h *HookServer) Run(args *HookRunArgs, reply *interface{}) error {
	client, err := rpc.Dial("tcp", args.RPCAddress)
	if err != nil {
		return err
	}

	h.hook.Run(args.Name, args.Data, &Ui{client})

	*reply = nil
	return nil
}
