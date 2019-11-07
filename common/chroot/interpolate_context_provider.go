package chroot

import "github.com/hashicorp/packer/template/interpolate"

type interpolateContextProvider interface {
	GetContext() interpolate.Context
}
