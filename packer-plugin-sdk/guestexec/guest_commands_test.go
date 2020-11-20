package guestexec

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
	if cmd != "powershell.exe -Command \"New-Item -ItemType directory -Force -ErrorAction SilentlyContinue -Path C:\\Windows\\Temp\\tempdir\"" {
		t.Fatalf("Unexpected Windows create dir cmd: %s", cmd)
	}

	// Windows OS w/ space in path
	cmd = guestCmd.CreateDir("C:\\Windows\\Temp\\temp dir")
	if cmd != "powershell.exe -Command \"New-Item -ItemType directory -Force -ErrorAction SilentlyContinue -Path C:\\Windows\\Temp\\temp` dir\"" {
		t.Fatalf("Unexpected Windows create dir cmd: %s", cmd)
	}
}

func TestChmod(t *testing.T) {
	// *nix
	guestCmd, err := NewGuestCommands(UnixOSType, false)
	if err != nil {
		t.Fatalf("Failed to create new GuestCommands for OS: %s", UnixOSType)
	}
	cmd := guestCmd.Chmod("/usr/local/bin/script.sh", "0666")
	if cmd != "chmod 0666 '/usr/local/bin/script.sh'" {
		t.Fatalf("Unexpected Unix chmod 0666 cmd: %s", cmd)
	}

	// sudo *nix
	guestCmd, err = NewGuestCommands(UnixOSType, true)
	if err != nil {
		t.Fatalf("Failed to create new sudo GuestCommands for OS: %s", UnixOSType)
	}
	cmd = guestCmd.Chmod("/usr/local/bin/script.sh", "+x")
	if cmd != "sudo chmod +x '/usr/local/bin/script.sh'" {
		t.Fatalf("Unexpected Unix chmod +x cmd: %s", cmd)
	}

	// Windows
	guestCmd, err = NewGuestCommands(WindowsOSType, false)
	if err != nil {
		t.Fatalf("Failed to create new GuestCommands for OS: %s", WindowsOSType)
	}
	cmd = guestCmd.Chmod("C:\\Program Files\\SomeApp\\someapp.exe", "+x")
	if cmd != "echo 'skipping chmod +x C:\\Program` Files\\SomeApp\\someapp.exe'" {
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
	if cmd != "powershell.exe -Command \"rm C:\\Temp\\SomeDir -recurse -force\"" {
		t.Fatalf("Unexpected Windows remove dir cmd: %s", cmd)
	}

	// Windows OS w/ space in path
	cmd = guestCmd.RemoveDir("C:\\Temp\\Some Dir")
	if cmd != "powershell.exe -Command \"rm C:\\Temp\\Some` Dir -recurse -force\"" {
		t.Fatalf("Unexpected Windows remove dir cmd: %s", cmd)
	}
}

func TestStatPath(t *testing.T) {
	// *nix
	guestCmd, err := NewGuestCommands(UnixOSType, false)
	if err != nil {
		t.Fatalf("Failed to create new GuestCommands for OS: %s", UnixOSType)
	}
	cmd := guestCmd.StatPath("/tmp/somedir")
	if cmd != "stat '/tmp/somedir'" {
		t.Fatalf("Unexpected Unix stat cmd: %s", cmd)
	}

	guestCmd, err = NewGuestCommands(UnixOSType, true)
	if err != nil {
		t.Fatalf("Failed to create new GuestCommands for OS: %s", UnixOSType)
	}
	cmd = guestCmd.StatPath("/tmp/somedir")
	if cmd != "sudo stat '/tmp/somedir'" {
		t.Fatalf("Unexpected Unix stat cmd: %s", cmd)
	}

	// Windows OS
	guestCmd, err = NewGuestCommands(WindowsOSType, false)
	if err != nil {
		t.Fatalf("Failed to create new GuestCommands for OS: %s", WindowsOSType)
	}
	cmd = guestCmd.StatPath("C:\\Temp\\SomeDir")
	if cmd != "powershell.exe -Command { if (test-path C:\\Temp\\SomeDir) { exit 0 } else { exit 1 } }" {
		t.Fatalf("Unexpected Windows stat cmd: %s", cmd)
	}

	// Windows OS w/ space in path
	cmd = guestCmd.StatPath("C:\\Temp\\Some Dir")
	if cmd != "powershell.exe -Command { if (test-path C:\\Temp\\Some` Dir) { exit 0 } else { exit 1 } }" {
		t.Fatalf("Unexpected Windows stat cmd: %s", cmd)
	}
}

func TestMovePath(t *testing.T) {
	// *nix
	guestCmd, err := NewGuestCommands(UnixOSType, false)
	if err != nil {
		t.Fatalf("Failed to create new GuestCommands for OS: %s", UnixOSType)
	}
	cmd := guestCmd.MovePath("/tmp/somedir", "/tmp/newdir")
	if cmd != "mv '/tmp/somedir' '/tmp/newdir'" {
		t.Fatalf("Unexpected Unix move cmd: %s", cmd)
	}

	// sudo *nix
	guestCmd, err = NewGuestCommands(UnixOSType, true)
	if err != nil {
		t.Fatalf("Failed to create new sudo GuestCommands for OS: %s", UnixOSType)
	}
	cmd = guestCmd.MovePath("/tmp/somedir", "/tmp/newdir")
	if cmd != "sudo mv '/tmp/somedir' '/tmp/newdir'" {
		t.Fatalf("Unexpected Unix sudo mv cmd: %s", cmd)
	}

	// Windows OS
	guestCmd, err = NewGuestCommands(WindowsOSType, false)
	if err != nil {
		t.Fatalf("Failed to create new GuestCommands for OS: %s", WindowsOSType)
	}
	cmd = guestCmd.MovePath("C:\\Temp\\SomeDir", "C:\\Temp\\NewDir")
	if cmd != "powershell.exe -Command \"mv C:\\Temp\\SomeDir C:\\Temp\\NewDir -force\"" {
		t.Fatalf("Unexpected Windows remove dir cmd: %s", cmd)
	}

	// Windows OS w/ space in path
	cmd = guestCmd.MovePath("C:\\Temp\\Some Dir", "C:\\Temp\\New Dir")
	if cmd != "powershell.exe -Command \"mv C:\\Temp\\Some` Dir C:\\Temp\\New` Dir -force\"" {
		t.Fatalf("Unexpected Windows remove dir cmd: %s", cmd)
	}
}
