package common

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

// Fusion5Driver is a driver that can run VMware Fusion 5.
type Fusion5Driver struct {
	VmwareDriver

	// This is the path to the "VMware Fusion.app"
	AppPath string

	// SSHConfig are the SSH settings for the Fusion VM
	SSHConfig *SSHConfig
}

func NewFusion5Driver(dconfig *DriverConfig, config *SSHConfig) Driver {
	return &Fusion5Driver{
		AppPath:   dconfig.FusionAppPath,
		SSHConfig: config,
	}
}

func (d *Fusion5Driver) Clone(dst, src string, linked bool, snapshot string) error {
	return errors.New("Cloning is not supported with Fusion 5. Please use Fusion 6+.")
}

func (d *Fusion5Driver) CompactDisk(diskPath string) error {
	defragCmd := exec.Command(d.vdiskManagerPath(), "-d", diskPath)
	if _, _, err := runAndLog(defragCmd); err != nil {
		return err
	}

	shrinkCmd := exec.Command(d.vdiskManagerPath(), "-k", diskPath)
	if _, _, err := runAndLog(shrinkCmd); err != nil {
		return err
	}

	return nil
}

func (d *Fusion5Driver) CreateDisk(output string, size string, adapter_type string, type_id string) error {
	cmd := exec.Command(d.vdiskManagerPath(), "-c", "-s", size, "-a", adapter_type, "-t", type_id, output)
	if _, _, err := runAndLog(cmd); err != nil {
		return err
	}

	return nil
}

func (d *Fusion5Driver) IsRunning(vmxPath string) (bool, error) {
	vmxPath, err := filepath.Abs(vmxPath)
	if err != nil {
		return false, err
	}

	cmd := exec.Command(d.vmrunPath(), "-T", "fusion", "list")
	stdout, _, err := runAndLog(cmd)
	if err != nil {
		return false, err
	}

	for _, line := range strings.Split(stdout, "\n") {
		if line == vmxPath {
			return true, nil
		}
	}

	return false, nil
}

func (d *Fusion5Driver) CommHost(state multistep.StateBag) (string, error) {
	return CommHost(d.SSHConfig)(state)
}

func (d *Fusion5Driver) Start(vmxPath string, headless bool) error {
	guiArgument := "gui"
	if headless == true {
		guiArgument = "nogui"
	}

	cmd := exec.Command(d.vmrunPath(), "-T", "fusion", "start", vmxPath, guiArgument)
	if _, _, err := runAndLog(cmd); err != nil {
		return err
	}

	return nil
}

func (d *Fusion5Driver) Stop(vmxPath string) error {
	cmd := exec.Command(d.vmrunPath(), "-T", "fusion", "stop", vmxPath, "hard")
	if _, _, err := runAndLog(cmd); err != nil {
		// Check if the VM is running. If its not, it was already stopped
		running, rerr := d.IsRunning(vmxPath)
		if rerr == nil && !running {
			return nil
		}

		return err
	}

	return nil
}

func (d *Fusion5Driver) SuppressMessages(vmxPath string) error {
	dir := filepath.Dir(vmxPath)
	base := filepath.Base(vmxPath)
	base = strings.Replace(base, ".vmx", "", -1)

	plistPath := filepath.Join(dir, base+".plist")
	return ioutil.WriteFile(plistPath, []byte(fusionSuppressPlist), 0644)
}

func (d *Fusion5Driver) Verify() error {
	if _, err := os.Stat(d.AppPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("Fusion application not found at path: %s", d.AppPath)
		}

		return err
	}

	if _, err := os.Stat(d.vmrunPath()); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(
				"Critical application 'vmrun' not found at path: %s", d.vmrunPath())
		}

		return err
	}

	if _, err := os.Stat(d.vdiskManagerPath()); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(
				"Critical application vdisk manager not found at path: %s",
				d.vdiskManagerPath())
		}

		return err
	}

	libpath := filepath.Join("/", "Library", "Preferences", "VMware Fusion")

	d.VmwareDriver.DhcpLeasesPath = func(device string) string {
		return "/var/db/vmware/vmnet-dhcpd-" + device + ".leases"
	}
	d.VmwareDriver.DhcpConfPath = func(device string) string {
		return filepath.Join(libpath, device, "dhcpd.conf")
	}
	d.VmwareDriver.VmnetnatConfPath = func(device string) string {
		return filepath.Join(libpath, device, "nat.conf")
	}
	d.VmwareDriver.NetworkMapper = func() (NetworkNameMapper, error) {
		pathNetworking := filepath.Join(libpath, "networking")
		if _, err := os.Stat(pathNetworking); err != nil {
			return nil, fmt.Errorf("Could not find networking conf file: %s", pathNetworking)
		}
		log.Printf("Located networkmapper configuration file using Fusion5: %s", pathNetworking)

		fd, err := os.Open(pathNetworking)
		if err != nil {
			return nil, err
		}
		defer fd.Close()

		return ReadNetworkingConfig(fd)
	}

	return nil
}

func (d *Fusion5Driver) vdiskManagerPath() string {
	return filepath.Join(d.AppPath, "Contents", "Library", "vmware-vdiskmanager")
}

func (d *Fusion5Driver) vmrunPath() string {
	return filepath.Join(d.AppPath, "Contents", "Library", "vmrun")
}

func (d *Fusion5Driver) ToolsIsoPath(k string) string {
	return filepath.Join(d.AppPath, "Contents", "Library", "isoimages", k+".iso")
}

func (d *Fusion5Driver) ToolsInstall() error {
	return nil
}

const fusionSuppressPlist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>disallowUpgrade</key>
	<true/>
</dict>
</plist>`

func (d *Fusion5Driver) GetVmwareDriver() VmwareDriver {
	return d.VmwareDriver
}
