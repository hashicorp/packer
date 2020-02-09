package rpc

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/rpc"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/zclconf/go-cty/cty"
)

// commonClient allows to rpc call funcs that can be defined on the different
// build blocks of packer
type commonClient struct {
	// endpoint is usually the type of build block we are connecting to.
	//
	// eg: Provisioner / PostProcessor / Builder / Artifact / Communicator
	endpoint string
	client   *rpc.Client
	mux      *muxBroker
}

type commonServer struct {
	mux              *muxBroker
	selfConfigurable interface {
		ConfigSpec() hcldec.ObjectSpec
	}
}

type ConfigSpecResponse struct {
	ConfigSpec []byte
}

func (p *commonClient) ConfigSpec() hcldec.ObjectSpec {
	// TODO(azr): the RPC Call can fail but the ConfigSpec signature doesn't
	// return an error; should we simply panic ? Logging this for now; will
	// decide later. The correct approach would probably be to return an error
	// in ConfigSpec but that will break a lot of things.
	resp := &ConfigSpecResponse{}
	cerr := p.client.Call(p.endpoint+".ConfigSpec", new(interface{}), resp)
	if cerr != nil {
		err := fmt.Errorf("ConfigSpec failed: %v", cerr)
		panic(err.Error())
	}

	res := hcldec.ObjectSpec{}
	err := gob.NewDecoder(bytes.NewReader(resp.ConfigSpec)).Decode(&res)
	if err != nil {
		panic("ici:" + err.Error())
	}
	return res
}

func (s *commonServer) ConfigSpec(_ interface{}, reply *ConfigSpecResponse) error {
	spec := s.selfConfigurable.ConfigSpec()
	b := bytes.NewBuffer(nil)
	err := gob.NewEncoder(b).Encode(spec)
	reply.ConfigSpec = b.Bytes()

	return err
}

func init() {
	gob.Register(new(hcldec.AttrSpec))
	gob.Register(new(hcldec.BlockSpec))
	gob.Register(new(hcldec.BlockAttrsSpec))
	gob.Register(new(hcldec.BlockListSpec))
	gob.Register(new(hcldec.BlockObjectSpec))
	gob.Register(new(cty.Value))
}
