package provisioner

import (
	"fmt"
	"strings"

	"github.com/mitchellh/packer/communicator/ssh"
	"github.com/mitchellh/packer/communicator/winrm"
	"github.com/mitchellh/packer/packer"
)

const UnixOSType = "unix"
const WindowsOSType = "windows"
const DefaultOSType = UnixOSType

type guestOSTypeCommand struct {
	chmod     string
	mkdir     string
	removeDir string
}

var guestOSTypeCommands = map[string]guestOSTypeCommand{
	UnixOSType: {
		chmod:     "chmod %s '%s'",
		mkdir:     "mkdir -p '%s'",
		removeDir: "rm -rf '%s'",
	},
	WindowsOSType: {
		chmod:     "echo 'skipping chmod %s %s'", // no-op
		mkdir:     "powershell.exe -Command \"New-Item -ItemType directory -Force -ErrorAction SilentlyContinue -Path %s\"",
		removeDir: "powershell.exe -Command \"rm %s -recurse -force\"",
	},
}

func GuestOSTypeFromComm(comm packer.Communicator) (string, error) {
	switch comm.(type) {
	case *winrm.Communicator:
		return WindowsOSType, nil
	case *ssh.Communicator:
		return UnixOSType, nil
	default:
		return "", fmt.Errorf("Unable to guess guest os type from connection type. Please specify `guest_os_type` in the provisioner.")
	}
}

type GuestCommands struct {
	GuestOSType string
	Sudo        bool
}

func NewGuestCommands(osType string, sudo bool) (*GuestCommands, error) {
	_, ok := guestOSTypeCommands[osType]
	if !ok {
		return nil, fmt.Errorf("Invalid osType: \"%s\"", osType)
	}
	return &GuestCommands{GuestOSType: osType, Sudo: sudo}, nil
}

func (g *GuestCommands) Chmod(path string, mode string) string {
	return g.sudo(fmt.Sprintf(g.commands().chmod, mode, g.escapePath(path)))
}

func (g *GuestCommands) CreateDir(path string) string {
	return g.sudo(fmt.Sprintf(g.commands().mkdir, g.escapePath(path)))
}

func (g *GuestCommands) RemoveDir(path string) string {
	return g.sudo(fmt.Sprintf(g.commands().removeDir, g.escapePath(path)))
}

func (g *GuestCommands) commands() guestOSTypeCommand {
	return guestOSTypeCommands[g.GuestOSType]
}

func (g *GuestCommands) escapePath(path string) string {
	if g.GuestOSType == WindowsOSType {
		return strings.Replace(path, " ", "` ", -1)
	}
	return path
}

func (g *GuestCommands) sudo(cmd string) string {
	if g.GuestOSType == UnixOSType && g.Sudo {
		return "sudo " + cmd
	}
	return cmd
}
