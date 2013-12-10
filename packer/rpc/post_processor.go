package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
)

// An implementation of packer.PostProcessor where the PostProcessor is actually
// executed over an RPC connection.
type postProcessor struct {
	client *rpc.Client
	mux    *MuxConn
}

// PostProcessorServer wraps a packer.PostProcessor implementation and makes it
// exportable as part of a Golang RPC server.
type PostProcessorServer struct {
	client *rpc.Client
	mux    *MuxConn
	p      packer.PostProcessor
}

type PostProcessorConfigureArgs struct {
	Configs []interface{}
}

type PostProcessorProcessResponse struct {
	Err      error
	Keep     bool
	StreamId uint32
}

func PostProcessor(client *rpc.Client) *postProcessor {
	return &postProcessor{client: client}
}

func (p *postProcessor) Configure(raw ...interface{}) (err error) {
	args := &PostProcessorConfigureArgs{Configs: raw}
	if cerr := p.client.Call("PostProcessor.Configure", args, &err); cerr != nil {
		err = cerr
	}

	return
}

func (p *postProcessor) PostProcess(ui packer.Ui, a packer.Artifact) (packer.Artifact, bool, error) {
	nextId := p.mux.NextId()
	server := NewServerWithMux(p.mux, nextId)
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

	client, err := NewClientWithMux(p.mux, response.StreamId)
	if err != nil {
		return nil, false, err
	}

	return client.Artifact(), response.Keep, nil
}

func (p *PostProcessorServer) Configure(args *PostProcessorConfigureArgs, reply *error) error {
	*reply = p.p.Configure(args.Configs...)
	if *reply != nil {
		*reply = NewBasicError(*reply)
	}

	return nil
}

func (p *PostProcessorServer) PostProcess(streamId uint32, reply *PostProcessorProcessResponse) error {
	client, err := NewClientWithMux(p.mux, streamId)
	if err != nil {
		return NewBasicError(err)
	}
	defer client.Close()

	streamId = 0
	artifactResult, keep, err := p.p.PostProcess(client.Ui(), client.Artifact())
	if err == nil && artifactResult != nil {
		streamId = p.mux.NextId()
		server := NewServerWithMux(p.mux, streamId)
		server.RegisterArtifact(artifactResult)
		go server.Serve()
	}

	if err != nil {
		err = NewBasicError(err)
	}

	*reply = PostProcessorProcessResponse{
		Err:      err,
		Keep:     keep,
		StreamId: streamId,
	}

	return nil
}
