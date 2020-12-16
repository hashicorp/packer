package rpc

import (
	"context"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/zclconf/go-cty/cty"
)

// An implementation of packersdk.DataSource where the data source is actually
// executed over an RPC connection.
type datasource struct {
	commonClient
}

type DataSourceConfigureArgs struct {
	Configs []interface{}
}

func (d *datasource) Configure(configs ...interface{}) error {
	configs, err := encodeCTYValues(configs)
	if err != nil {
		return err
	}
	args := &DataSourceConfigureArgs{Configs: configs}
	return d.client.Call(d.endpoint+".Configure", args, new(interface{}))
}

func (d *datasource) Execute() (cty.Value, error) {
	nextId := d.mux.NextId()
	server := newServerWithMux(d.mux, nextId)
	go server.Serve()

	done := make(chan interface{})
	defer close(done)

	var result cty.Value
	if err := d.client.Call(d.endpoint+".Execute", nextId, &result); err != nil {
		return cty.NilVal, err
	}

	return result, nil
}

// DataSourceServer wraps a packersdk.DataSource implementation and makes it
// exportable as part of a Golang RPC server.
type DataSourceServer struct {
	context       context.Context
	contextCancel func()

	commonServer
	d packersdk.DataSource
}

func (d *DataSourceServer) Configure(args *DataSourceConfigureArgs, reply *interface{}) error {
	config, err := decodeCTYValues(args.Configs)
	if err != nil {
		return err
	}
	return d.d.Configure(config...)
}

func (d *DataSourceServer) Execute(streamId uint32, reply *interface{}) error {
	client, err := newClientWithMux(d.mux, streamId)
	if err != nil {
		return NewBasicError(err)
	}
	defer client.Close()

	if d.context == nil {
		d.context, d.contextCancel = context.WithCancel(context.Background())
	}
	if _, err := d.d.Execute(); err != nil {
		return NewBasicError(err)
	}

	return nil
}

func (d *DataSourceServer) Cancel(args *interface{}, reply *interface{}) error {
	if d.contextCancel != nil {
		d.contextCancel()
	}
	return nil
}
