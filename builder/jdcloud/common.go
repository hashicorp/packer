package jdcloud

import (
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/template/interpolate"
	vm "github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	vpc "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
)

const (
	FINE           = 0
	CONNECT_FAILED = "Client.Timeout exceeded"
	VM_PENDING     = "pending"
	VM_RUNNING     = "running"
	VM_STARTING    = "starting"
	VM_STOPPING    = "stopping"
	VM_STOPPED     = "stopped"
	READY          = "ready"
	BUILDER_ID     = "hashicorp.jdcloud"
)

var (
	VmClient  *vm.VmClient
	VpcClient *vpc.VpcClient
	Region    string
)

type Config struct {
	JDCloudCredentialConfig   `mapstructure:",squash"`
	JDCloudInstanceSpecConfig `mapstructure:",squash"`
	common.PackerConfig       `mapstructure:",squash"`
	ctx                       interpolate.Context
}

type Builder struct {
	config Config
	runner multistep.Runner
}
