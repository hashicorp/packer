package qemu

import (
	"fmt"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/stretchr/testify/assert"
)

func runTestState(t *testing.T, config *Config) multistep.StateBag {
	state := new(multistep.BasicStateBag)

	state.Put("ui", packer.TestUi(t))
	state.Put("config", config)

	d := new(DriverMock)
	d.VersionResult = "3.0.0"
	state.Put("driver", d)

	state.Put("commHostPort", 5000)
	state.Put("floppy_path", "fake_floppy_path")
	state.Put("http_ip", "127.0.0.1")
	state.Put("http_port", 1234)
	state.Put("iso_path", "/path/to/test.iso")
	state.Put("qemu_disk_paths", []string{})
	state.Put("vnc_port", 5905)
	state.Put("vnc_password", "fake_vnc_password")

	return state
}

func Test_getCommandArgs(t *testing.T) {
	state := runTestState(t, &Config{})

	args, err := getCommandArgs("", state)
	if err != nil {
		t.Fatalf("should not have an error getting args. Error: %s", err)
	}

	expected := []string{
		"-display", "gtk",
		"-m", "0M",
		"-boot", "",
		"-fda", "fake_floppy_path",
		"-name", "",
		"-netdev", "user,id=user.0,hostfwd=tcp::5000-:0",
		"-vnc", ":5905",
		"-machine", "type=,accel=",
		"-device", ",netdev=user.0",
		"-drive", "file=/path/to/test.iso,index=0,media=cdrom",
	}

	assert.ElementsMatch(t, args, expected, "unexpected generated args")
}

func Test_CDFilesPath(t *testing.T) {
	// cd_path is set and DiskImage is false
	state := runTestState(t, &Config{})
	state.Put("cd_path", "fake_cd_path.iso")

	args, err := getCommandArgs("", state)
	if err != nil {
		t.Fatalf("should not have an error getting args. Error: %s", err)
	}

	expected := []string{
		"-display", "gtk",
		"-m", "0M",
		"-boot", "",
		"-fda", "fake_floppy_path",
		"-name", "",
		"-netdev", "user,id=user.0,hostfwd=tcp::5000-:0",
		"-vnc", ":5905",
		"-machine", "type=,accel=",
		"-device", ",netdev=user.0",
		"-drive", "file=/path/to/test.iso,index=0,media=cdrom",
		"-drive", "file=fake_cd_path.iso,index=1,media=cdrom",
	}

	assert.ElementsMatch(t, args, expected, fmt.Sprintf("unexpected generated args: %#v", args))

	// cd_path is set and DiskImage is true
	config := &Config{
		DiskImage:     true,
		DiskInterface: "virtio-scsi",
	}
	state = runTestState(t, config)
	state.Put("cd_path", "fake_cd_path.iso")

	args, err = getCommandArgs("c", state)
	if err != nil {
		t.Fatalf("should not have an error getting args. Error: %s", err)
	}

	expected = []string{
		"-display", "gtk",
		"-m", "0M",
		"-boot", "c",
		"-fda", "fake_floppy_path",
		"-name", "",
		"-netdev", "user,id=user.0,hostfwd=tcp::5000-:0",
		"-vnc", ":5905",
		"-machine", "type=,accel=",
		"-device", ",netdev=user.0",
		"-device", "virtio-scsi-pci,id=scsi0",
		"-device", "scsi-hd,bus=scsi0.0,drive=drive0",
		"-drive", "if=none,file=,id=drive0,cache=,discard=,format=,detect-zeroes=",
		"-drive", "file=fake_cd_path.iso,index=0,media=cdrom",
	}

	assert.ElementsMatch(t, args, expected, fmt.Sprintf("unexpected generated args: %#v", args))
}

func Test_OptionalConfigOptionsGetSet(t *testing.T) {
	c := &Config{
		VNCUsePassword: true,
		QMPEnable:      true,
		QMPSocketPath:  "qmp_path",
		VMName:         "MyFancyName",
		MachineType:    "pc",
		Accelerator:    "hvf",
	}

	state := runTestState(t, c)

	args, err := getCommandArgs("once=d", state)
	if err != nil {
		t.Fatalf("should not have an error getting args. Error: %s", err)
	}

	expected := []string{
		"-display", "gtk",
		"-m", "0M",
		"-boot", "once=d",
		"-fda", "fake_floppy_path",
		"-name", "MyFancyName",
		"-netdev", "user,id=user.0,hostfwd=tcp::5000-:0",
		"-vnc", ":5905,password",
		"-machine", "type=pc,accel=hvf",
		"-device", ",netdev=user.0",
		"-drive", "file=/path/to/test.iso,index=0,media=cdrom",
		"-qmp", "unix:qmp_path,server,nowait",
	}

	assert.ElementsMatch(t, args, expected, "password flag should be set, and d drive should be set.")
}
