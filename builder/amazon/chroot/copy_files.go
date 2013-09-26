package chroot

import (
	"fmt"
	"os/exec"
)

func copySingle(dest string, src string, copyCommand string) error {
	cpCommand := fmt.Sprintf("%s %s %s", copyCommand, src, dest)
	localCmd := exec.Command("/bin/sh", "-c", cpCommand)
	if err := localCmd.Run(); err != nil {
		return err
	}
	return nil
}
