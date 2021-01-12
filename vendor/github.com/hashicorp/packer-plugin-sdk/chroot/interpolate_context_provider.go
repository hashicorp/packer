package chroot

import "github.com/hashicorp/packer-plugin-sdk/template/interpolate"

type interpolateContextProvider interface {
	GetContext() interpolate.Context
}
