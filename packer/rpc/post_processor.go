package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
)

// An implementation of packer.PostProcessor where the PostProcessor is actually
// executed over an RPC connection.
type postProcessor struct {
	client *rpc.Client
	mux    *muxBroker
}

// PostProcessorServer wraps a packer.PostProcessor implementation and makes it
// exportable as part of a Golang RPC server.
type PostProcessorServer struct {
	client *rpc.Client
	mux    *muxBroker
	p      packer.PostProcessor
}

type PostProcessorConfigureArgs struct {
	Configs []interface{}
}

type PostProcessorProcessResponse struct {
	Err      *BasicError
	Keep     bool
	StreamId uint32
}

func (p *postProcessor) Configure(raw ...interface{}) (err error) {
	args := &PostProcessorConfigureArgs{Configs: raw}
	if cerr := p.client.Call("PostProcessor.Configure", args, new(interface{})); cerr != nil {
		err = cerr
	}

	return
}

func (p *postProcessor) PostProcess(ui packer.Ui, a packer.Artifact) (packer.Artifact, bool, error) {
	nextId := p.mux.NextId()
	server := newServerWithMux(p.mux, nextId)
	server.RegisterArtifact(a)
	server.RegisterUi(ui)
	go server.Serve()

	var response PostProcessorProcessResponse
	if err := p.client.Call("PostProcessor.PostProcess", nextId, &response); err != nil {
		return nil, false, err
	}

	if response.Err != nil {
		return nil, false, response.Err
	}

	if response.StreamId == 0 {
		return nil, false, nil
	}

	client, err := newClientWithMux(p.mux, response.StreamId)
	if err != nil {
		return nil, false, err
	}

	return client.Artifact(), response.Keep, nil
}

func (p *PostProcessorServer) Configure(args *PostProcessorConfigureArgs, reply *interface{}) error {
	err := p.p.Configure(args.Configs...)
	return err
}

func (p *PostProcessorServer) PostProcess(streamId uint32, reply *PostProcessorProcessResponse) error {
	client, err := newClientWithMux(p.mux, streamId)
	if err != nil {
		return NewBasicError(err)
	}
	defer client.Close()

	streamId = 0
	artifactResult, keep, err := p.p.PostProcess(client.Ui(), client.Artifact())
	if err == nil && artifactResult != nil {
		streamId = p.mux.NextId()
		server := newServerWithMux(p.mux, streamId)
		server.RegisterArtifact(artifactResult)
		go server.Serve()
	}

	*reply = PostProcessorProcessResponse{
		Err:      NewBasicError(err),
		Keep:     keep,
		StreamId: streamId,
	}

	return nil
}
