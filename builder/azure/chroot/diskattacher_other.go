// +build !linux,!freebsd

package chroot

import (
	"context"
)

func (da diskAttacher) WaitForDevice(ctx context.Context, lun int32) (device string, err error) {
	panic("The azure-chroot builder does not work on this platform.")
}
