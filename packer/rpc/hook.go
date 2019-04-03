package rpc

import (
	"context"
	"log"
	"net/rpc"

	"github.com/hashicorp/packer/packer"
)

// An implementation of packer.Hook where the hook is actually executed
// over an RPC connection.
type hook struct {
	client *rpc.Client
	mux    *muxBroker
}

// HookServer wraps a packer.Hook implementation and makes it exportable
// as part of a Golang RPC server.
type HookServer struct {
	hook packer.Hook
	mux  *muxBroker
}

type HookRunArgs struct {
	Name     string
	Data     interface{}
	StreamId uint32
}

func (h *hook) Run(ctx context.Context, name string, ui packer.Ui, comm packer.Communicator, data interface{}) error {
	nextId := h.mux.NextId()
	server := newServerWithMux(h.mux, nextId)
	server.RegisterCommunicator(comm)
	server.RegisterUi(ui)
	go server.Serve()

	args := HookRunArgs{
		Name:     name,
		Data:     data,
		StreamId: nextId,
	}

	return h.client.Call("Hook.Run", &args, new(interface{}))
}

func (h *hook) Cancel() {
	err := h.client.Call("Hook.Cancel", new(interface{}), new(interface{}))
	if err != nil {
		log.Printf("Hook.Cancel error: %s", err)
	}
}

func (h *HookServer) Run(ctx context.Context, args *HookRunArgs, reply *interface{}) error {
	client, err := newClientWithMux(h.mux, args.StreamId)
	if err != nil {
		return NewBasicError(err)
	}
	defer client.Close()

	if err := h.hook.Run(ctx, args.Name, client.Ui(), client.Communicator(), args.Data); err != nil {
		return NewBasicError(err)
	}

	*reply = nil
	return nil
}
