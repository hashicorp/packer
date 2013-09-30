package chroot

import (
	"fmt"
	"log"
	"os/exec"
)

func ChrootCommand(chroot string, command string) *exec.Cmd {
	cmd := fmt.Sprintf("sudo chroot %s", chroot)
	return ShellCommand(cmd, command)
}

func ShellCommand(commands ...string) *exec.Cmd {
	cmds := append([]string{"-c"}, commands...)
	cmd := exec.Command("/bin/sh", cmds...)
	log.Printf("ShellCommand: %s %v", cmd.Path, cmd.Args[1:])
	return cmd
}
