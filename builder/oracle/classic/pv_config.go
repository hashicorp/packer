package classic

import (
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type PVConfig struct {
	// PersistentVolumeSize lets us control the volume size by using persistent boot storage
	PersistentVolumeSize      int    `mapstructure:"persistent_volume_size"`
	BuilderImageList          string `mapstructure:"builder_image_list"`
	BuilderUploadImageCommand string `mapstructure:"builder_upload_image_command"`
	/* TODO:
	default to OL image
	make sure if set then PVS is above
	some way to choose which connection to use for master
	possible ignore everything for builder and always use SSH keys
	*/
}

func (c *PVConfig) IsPV() bool {
	return c.PersistentVolumeSize > 0
}

func (c *PVConfig) Prepare(ctx *interpolate.Context) (errs *packer.MultiError) {
	if !c.IsPV() {
		return nil
	}

	if c.BuilderUploadImageCommand == "" {
		c.BuilderUploadImageCommand = `curl --connect-timeout 5 \
--max-time 3600 \
--retry 5 \
--retry-delay 0 \
-o {{ .DiskImagePath }} \
'...'`
		c.BuilderUploadImageCommand = "false"
	}
	/*
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("Persistent storage volumes are only supported on unix, and must use the ssh communicator."))
	*/
	return
}
