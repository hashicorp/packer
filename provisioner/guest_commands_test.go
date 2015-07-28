package provisioner

import (
	"testing"
)

func TestNewGuestCommands(t *testing.T) {
	_, err := NewGuestCommands("Amiga", true)
	if err == nil {
		t.Fatalf("Should have returned an err for unsupported OS type")
	}
}

func TestCreateDir(t *testing.T) {
	// *nix OS
	guestCmd, err := NewGuestCommands(UnixOSType, false)
	if err != nil {
		t.Fatalf("Failed to create new GuestCommands for OS: %s", UnixOSType)
	}
	cmd := guestCmd.CreateDir("/tmp/tempdir")
	if cmd != "mkdir -p '/tmp/tempdir'" {
		t.Fatalf("Unexpected Unix create dir cmd: %s", cmd)
	}

	// *nix OS w/sudo
	guestCmd, err = NewGuestCommands(UnixOSType, true)
	if err != nil {
		t.Fatalf("Failed to create new sudo GuestCommands for OS: %s", UnixOSType)
	}
	cmd = guestCmd.CreateDir("/tmp/tempdir")
	if cmd != "sudo mkdir -p '/tmp/tempdir'" {
		t.Fatalf("Unexpected Unix sudo create dir cmd: %s", cmd)
	}

	// Windows OS
	guestCmd, err = NewGuestCommands(WindowsOSType, false)
	if err != nil {
		t.Fatalf("Failed to create new GuestCommands for OS: %s", WindowsOSType)
	}
	cmd = guestCmd.CreateDir("C:\\Windows\\Temp\\tempdir")
	if cmd != "New-Item -ItemType directory -Force -ErrorAction SilentlyContinue -Path C:\\Windows\\Temp\\tempdir" {
		t.Fatalf("Unexpected Windows create dir cmd: %s", cmd)
	}

	// Windows OS w/ space in path
	cmd = guestCmd.CreateDir("C:\\Windows\\Temp\\temp dir")
	if cmd != "New-Item -ItemType directory -Force -ErrorAction SilentlyContinue -Path C:\\Windows\\Temp\\temp` dir" {
		t.Fatalf("Unexpected Windows create dir cmd: %s", cmd)
	}
}

func TestChmodExecutable(t *testing.T) {
	// *nix
	guestCmd, err := NewGuestCommands(UnixOSType, false)
	if err != nil {
		t.Fatalf("Failed to create new GuestCommands for OS: %s", UnixOSType)
	}
	cmd := guestCmd.ChmodExecutable("/usr/local/bin/script.sh")
	if cmd != "chmod +x '/usr/local/bin/script.sh'" {
		t.Fatalf("Unexpected Unix chmod +x cmd: %s", cmd)
	}

	// sudo *nix
	guestCmd, err = NewGuestCommands(UnixOSType, true)
	if err != nil {
		t.Fatalf("Failed to create new sudo GuestCommands for OS: %s", UnixOSType)
	}
	cmd = guestCmd.ChmodExecutable("/usr/local/bin/script.sh")
	if cmd != "sudo chmod +x '/usr/local/bin/script.sh'" {
		t.Fatalf("Unexpected Unix chmod +x cmd: %s", cmd)
	}

	// Windows
	guestCmd, err = NewGuestCommands(WindowsOSType, false)
	if err != nil {
		t.Fatalf("Failed to create new GuestCommands for OS: %s", WindowsOSType)
	}
	cmd = guestCmd.ChmodExecutable("C:\\Program Files\\SomeApp\\someapp.exe")
	if cmd != "echo 'skipping chmod C:\\Program` Files\\SomeApp\\someapp.exe'" {
		t.Fatalf("Unexpected Windows chmod +x cmd: %s", cmd)
	}
}

func TestRemoveDir(t *testing.T) {
	// *nix
	guestCmd, err := NewGuestCommands(UnixOSType, false)
	if err != nil {
		t.Fatalf("Failed to create new GuestCommands for OS: %s", UnixOSType)
	}
	cmd := guestCmd.RemoveDir("/tmp/somedir")
	if cmd != "rm -rf '/tmp/somedir'" {
		t.Fatalf("Unexpected Unix remove dir cmd: %s", cmd)
	}

	// sudo *nix
	guestCmd, err = NewGuestCommands(UnixOSType, true)
	if err != nil {
		t.Fatalf("Failed to create new sudo GuestCommands for OS: %s", UnixOSType)
	}
	cmd = guestCmd.RemoveDir("/tmp/somedir")
	if cmd != "sudo rm -rf '/tmp/somedir'" {
		t.Fatalf("Unexpected Unix sudo remove dir cmd: %s", cmd)
	}

	// Windows OS
	guestCmd, err = NewGuestCommands(WindowsOSType, false)
	if err != nil {
		t.Fatalf("Failed to create new GuestCommands for OS: %s", WindowsOSType)
	}
	cmd = guestCmd.RemoveDir("C:\\Temp\\SomeDir")
	if cmd != "rm C:\\Temp\\SomeDir -recurse -force" {
		t.Fatalf("Unexpected Windows remove dir cmd: %s", cmd)
	}

	// Windows OS w/ space in path
	cmd = guestCmd.RemoveDir("C:\\Temp\\Some Dir")
	if cmd != "rm C:\\Temp\\Some` Dir -recurse -force" {
		t.Fatalf("Unexpected Windows remove dir cmd: %s", cmd)
	}
}
