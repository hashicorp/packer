package chroot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

func diskPathForLun(lun int32) string {
	return fmt.Sprintf("/dev/disk/azure/scsi1/lun%d", lun)
}

func (da diskAttacher) WaitForDevice(ctx context.Context, lun int32) (device string, err error) {
	path := diskPathForLun(lun)

	for {
		link, err := os.Readlink(path)
		if err == nil {
			return filepath.Abs("/dev/disk/azure/scsi1/" + link)
		} else if err != os.ErrNotExist {
			if pe, ok := err.(*os.PathError); ok && pe.Err != syscall.ENOENT {
				return "", err
			}
		}

		select {
		case <-time.After(100 * time.Millisecond):
			// continue
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}
