package rpc

import (
	"net/rpc"

	"github.com/hashicorp/packer/packer"
)

// An implementation of packer.PreProcessor where the PreProcessor is actually
// executed over an RPC connection.
type preProcessor struct {
	client *rpc.Client
	mux    *muxBroker
}

// PreProcessorServer wraps a packer.PreProcessor implementation and makes it
// exportable as part of a Golang RPC server.
type PreProcessorServer struct {
	mux *muxBroker
	p   packer.PreProcessor
}

type PreProcessorConfigureArgs struct {
	Configs []interface{}
}

type PreProcessorProcessResponse struct {
	Err      *BasicError
	Keep     bool
	StreamId uint32
}

func (p *preProcessor) Configure(raw ...interface{}) (err error) {
	args := &PreProcessorConfigureArgs{Configs: raw}
	if cerr := p.client.Call("PreProcessor.Configure", args, new(interface{})); cerr != nil {
		err = cerr
	}

	return
}

func (p *preProcessor) PreProcess(ui packer.Ui) (err error) {
	nextId := p.mux.NextId()
	server := newServerWithMux(p.mux, nextId)
	server.RegisterUi(ui)
	go server.Serve()

	var response PreProcessorProcessResponse
	if err := p.client.Call("PreProcessor.PreProcess", nextId, &response); err != nil {
		return err
	}

	if response.Err != nil {
		return response.Err
	}

	if response.StreamId == 0 {
		return nil
	}

	_, err = newClientWithMux(p.mux, response.StreamId)
	if err != nil {
		return err
	}

	return nil
}

func (p *PreProcessorServer) Configure(args *PreProcessorConfigureArgs, reply *interface{}) error {
	err := p.p.Configure(args.Configs...)
	return err
}

func (p *PreProcessorServer) PreProcess(streamId uint32, reply *PreProcessorProcessResponse) error {
	client, err := newClientWithMux(p.mux, streamId)
	if err != nil {
		return NewBasicError(err)
	}
	defer client.Close()

	streamId = 0
	err = p.p.PreProcess(client.Ui())
	if err == nil {
		streamId = p.mux.NextId()
		server := newServerWithMux(p.mux, streamId)
		go server.Serve()
	}

	*reply = PreProcessorProcessResponse{
		Err:      NewBasicError(err),
		StreamId: streamId,
	}

	return nil
}
