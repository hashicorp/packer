package jdcloud

import (
	"fmt"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/template/interpolate"
)

type JDCloudInstanceSpecConfig struct {
	ImageId         string              `mapstructure:"image_id"`
	InstanceName    string              `mapstructure:"instance_name"`
	InstanceType    string              `mapstructure:"instance_type"`
	ImageName       string              `mapstructure:"image_name"`
	SubnetId        string              `mapstructure:"subnet_id"`
	Comm            communicator.Config `mapstructure:",squash"`
	InstanceId      string
	ArtifactId      string
	PublicIpAddress string
}

func (jd *JDCloudInstanceSpecConfig) Prepare(ctx *interpolate.Context) []error {

	errs := jd.Comm.Prepare(ctx)

	if jd == nil {
		return append(errs, fmt.Errorf("[PRE-FLIGHT] Configuration appears to be empty"))
	}

	if len(jd.ImageId) == 0 {
		errs = append(errs, fmt.Errorf("[PRE-FLIGHT] 'image_id' empty"))
	}

	if len(jd.InstanceName) == 0 {
		errs = append(errs, fmt.Errorf("[PRE-FLIGHT] 'instance_name' empty"))
	}

	if len(jd.InstanceType) == 0 {
		errs = append(errs, fmt.Errorf("[PRE-FLIGHT] 'instance-type' empty"))
	}

	noPassword := len(jd.Comm.SSHPassword) == 0
	noKeys := len(jd.Comm.SSHKeyPairName) == 0 && len(jd.Comm.SSHPrivateKeyFile) == 0
	noTempKey := len(jd.Comm.SSHTemporaryKeyPairName) == 0
	if noPassword && noKeys && noTempKey {
		errs = append(errs, fmt.Errorf("[PRE-FLIGHT] Didn't detect any credentials, you have to specify either "+
			"{password} or "+
			"{key_name+local_private_key_path} or "+
			"{temporary_key_pair_name} cheers :)"))
	}

	return errs
}
