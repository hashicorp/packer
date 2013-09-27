package chroot

import (
	"fmt"
	"log"
	"os/exec"
)

func ChrootCommand(chroot string, command string) *exec.Cmd {
	chrootCommand := fmt.Sprintf("chroot %s %s", chroot, command)
	return ShellCommand(chrootCommand)
}

func ShellCommand(command string) *exec.Cmd {
	cmd := exec.Command("/bin/sh", "-c", command)
	log.Printf("WrappedCommand(%s) -> #%v", command, cmd.Args)
	return cmd
}
