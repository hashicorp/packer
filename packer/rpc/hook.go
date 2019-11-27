package rpc

import (
	"context"
	"log"
	"sync"

	"github.com/hashicorp/packer/packer"
)

// An implementation of packer.Hook where the hook is actually executed
// over an RPC connection.
type hook struct {
	commonClient
}

// HookServer wraps a packer.Hook implementation and makes it exportable
// as part of a Golang RPC server.
type HookServer struct {
	context       context.Context
	contextCancel func()

	hook packer.Hook
	lock sync.Mutex
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

	done := make(chan interface{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			log.Printf("Cancelling hook after context cancellation %v", ctx.Err())
			if err := h.client.Call(h.endpoint+".Cancel", new(interface{}), new(interface{})); err != nil {
				log.Printf("Error cancelling builder: %s", err)
			}
		case <-done:
		}
	}()

	args := HookRunArgs{
		Name:     name,
		Data:     data,
		StreamId: nextId,
	}

	return h.client.Call(h.endpoint+".Run", &args, new(interface{}))
}

func (h *HookServer) Run(args *HookRunArgs, reply *interface{}) error {
	client, err := newClientWithMux(h.mux, args.StreamId)
	if err != nil {
		return NewBasicError(err)
	}
	defer client.Close()

	h.lock.Lock()
	if h.context == nil {
		h.context, h.contextCancel = context.WithCancel(context.Background())
	}
	h.lock.Unlock()
	if err := h.hook.Run(h.context, args.Name, client.Ui(), client.Communicator(), args.Data); err != nil {
		return NewBasicError(err)
	}

	*reply = nil
	return nil
}

func (h *HookServer) Cancel(args *interface{}, reply *interface{}) error {
	h.lock.Lock()
	if h.contextCancel != nil {
		h.contextCancel()
	}
	h.lock.Unlock()
	return nil
}
