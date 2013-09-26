package chroot

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func copySingle(dst string, src string, copyCommand string) error {
	cpCommand := fmt.Sprintf("sudo cp -fn %s %s", src, dest)
	localcmd := exec.Command("/bin/sh", "-c", cpCommand)
	log.Println(localcmd.Args)
	out, err := localcmd.CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	log.Printf("output: %s", out)
	return nil
}
