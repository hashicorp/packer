package chroot

import "github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"

type interpolateContextProvider interface {
	GetContext() interpolate.Context
}
