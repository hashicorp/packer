package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
)

// An implementation of packer.PostProcessor where the PostProcessor is actually
// executed over an RPC connection.
type postProcessor struct {
	client *rpc.Client
}

// PostProcessorServer wraps a packer.PostProcessor implementation and makes it
// exportable as part of a Golang RPC server.
type PostProcessorServer struct {
	p packer.PostProcessor
}

type PostProcessorConfigureArgs struct {
	Configs []interface{}
}

type PostProcessorProcessResponse struct {
	Err        error
	Keep       bool
	RPCAddress string
}

func PostProcessor(client *rpc.Client) *postProcessor {
	return &postProcessor{client}
}
func (p *postProcessor) Configure(raw ...interface{}) (err error) {
	args := &PostProcessorConfigureArgs{Configs: raw}
	if cerr := p.client.Call("PostProcessor.Configure", args, &err); cerr != nil {
		err = cerr
	}

	return
}

func (p *postProcessor) PostProcess(ui packer.Ui, a packer.Artifact) (packer.Artifact, bool, error) {
	server := rpc.NewServer()
	RegisterArtifact(server, a)
	RegisterUi(server, ui)

	var response PostProcessorProcessResponse
	if err := p.client.Call("PostProcessor.PostProcess", serveSingleConn(server), &response); err != nil {
		return nil, false, err
	}

	if response.Err != nil {
		return nil, false, response.Err
	}

	if response.RPCAddress == "" {
		return nil, false, nil
	}

	client, err := rpcDial(response.RPCAddress)
	if err != nil {
		return nil, false, err
	}

	return Artifact(client), response.Keep, nil
}

func (p *PostProcessorServer) Configure(args *PostProcessorConfigureArgs, reply *error) error {
	*reply = p.p.Configure(args.Configs...)
	if *reply != nil {
		*reply = NewBasicError(*reply)
	}

	return nil
}

func (p *PostProcessorServer) PostProcess(address string, reply *PostProcessorProcessResponse) error {
	client, err := rpcDial(address)
	if err != nil {
		return err
	}

	responseAddress := ""

	artifact, keep, err := p.p.PostProcess(&Ui{client: client}, Artifact(client))
	if err == nil && artifact != nil {
		server := rpc.NewServer()
		RegisterArtifact(server, artifact)
		responseAddress = serveSingleConn(server)
	}

	if err != nil {
		err = NewBasicError(err)
	}

	*reply = PostProcessorProcessResponse{
		Err:        err,
		Keep:       keep,
		RPCAddress: responseAddress,
	}

	return nil
}
