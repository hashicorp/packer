package provisioner

import (
	"fmt"
	"strings"
)

const UnixOSType = "unix"
const WindowsOSType = "windows"
const DefaultOSType = UnixOSType

type guestOSTypeCommand struct {
	chmodExecutable string
	mkdir           string
	removeDir       string
}

var guestOSTypeCommands = map[string]guestOSTypeCommand{
	UnixOSType: guestOSTypeCommand{
		chmodExecutable: "chmod +x '%s'",
		mkdir:           "mkdir -p '%s'",
		removeDir:       "rm -rf '%s'",
	},
	WindowsOSType: guestOSTypeCommand{
		chmodExecutable: "echo 'skipping chmod %s'", // no-op
		mkdir:           "New-Item -ItemType directory -Force -ErrorAction SilentlyContinue -Path %s",
		removeDir:       "rm %s -recurse -force",
	},
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

func (g *GuestCommands) ChmodExecutable(path string) string {
	return g.sudo(fmt.Sprintf(g.commands().chmodExecutable, g.escapePath(path)))
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
