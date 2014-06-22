package common

import (
	"bytes"
	"fmt"
	"github.com/going/toolkit/xmlpath"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type Parallels9Driver struct {
	// This is the path to the "prlctl" application.
	PrlctlPath string
}

func (d *Parallels9Driver) Import(name, srcPath, dstDir string) error {

	err := d.Prlctl("register", srcPath, "--preserve-uuid")
	if err != nil {
		return err
	}

	srcId, err := getVmId(srcPath)
	if err != nil {
		return err
	}

	srcMac, err := getFirtsMacAddress(srcPath)
	if err != nil {
		return err
	}

	err = d.Prlctl("clone", srcId, "--name", name, "--dst", dstDir)
	if err != nil {
		return err
	}

	err = d.Prlctl("unregister", srcId)
	if err != nil {
		return err
	}

	err = d.Prlctl("set", name, "--device-set", "net0", "--mac", srcMac)
	return nil
}

func getVmId(path string) (string, error) {
	return getConfigValueFromXpath(path, "/ParallelsVirtualMachine/Identification/VmUuid")
}

func getFirtsMacAddress(path string) (string, error) {
	return getConfigValueFromXpath(path, "/ParallelsVirtualMachine/Hardware/NetworkAdapter[@id='0']/MAC")
}

func getConfigValueFromXpath(path, xpath string) (string, error) {
	file, err := os.Open(path + "/config.pvs")
	if err != nil {
		return "", err
	}
	xpathComp := xmlpath.MustCompile(xpath)
	root, err := xmlpath.Parse(file)
	if err != nil {
		return "", err
	}
	value, _ := xpathComp.String(root)
	return value, nil
}

func (d *Parallels9Driver) IsRunning(name string) (bool, error) {
	var stdout bytes.Buffer

	cmd := exec.Command(d.PrlctlPath, "list", name, "--no-header", "--output", "status")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return false, err
	}

	log.Printf("Checking VM state: %s\n", strings.TrimSpace(stdout.String()))

	for _, line := range strings.Split(stdout.String(), "\n") {
		if line == "running" {
			return true, nil
		}

		if line == "suspended" {
			return true, nil
		}
		if line == "paused" {
			return true, nil
		}
		if line == "stopping" {
			return true, nil
		}
	}

	return false, nil
}

func (d *Parallels9Driver) Stop(name string) error {
	if err := d.Prlctl("stop", name); err != nil {
		return err
	}

	// We sleep here for a little bit to let the session "unlock"
	time.Sleep(2 * time.Second)

	return nil
}

func (d *Parallels9Driver) Prlctl(args ...string) error {
	var stdout, stderr bytes.Buffer

	log.Printf("Executing prlctl: %#v", args)
	cmd := exec.Command(d.PrlctlPath, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	stdoutString := strings.TrimSpace(stdout.String())
	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("prlctl error: %s", stderrString)
	}

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	return err
}

func (d *Parallels9Driver) Verify() error {
	version, _ := d.Version()
	if !strings.HasPrefix(version, "9.") {
		return fmt.Errorf("The packer-parallels builder plugin only supports Parallels Desktop v. 9. You have: %s!\n", version)
	}
	return nil
}

func (d *Parallels9Driver) Version() (string, error) {
	var stdout bytes.Buffer

	cmd := exec.Command(d.PrlctlPath, "--version")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	versionOutput := strings.TrimSpace(stdout.String())
	re := regexp.MustCompile("prlctl version ([0-9\\.]+)")
	verMatch := re.FindAllStringSubmatch(versionOutput, 1)

	if len(verMatch) != 1 {
		return "", fmt.Errorf("prlctl version not found!\n")
	}

	version := verMatch[0][1]
	log.Printf("prlctl version: %s\n", version)
	return version, nil
}

func (d *Parallels9Driver) SendKeyScanCodes(vmName string, codes ...string) error {
	var stdout, stderr bytes.Buffer

	args := prepend(vmName, codes)
	cmd := exec.Command("prltype", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	stdoutString := strings.TrimSpace(stdout.String())
	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("prltype error: %s", stderrString)
	}

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	return err
}

func prepend(head string, tail []string) []string {
	tmp := make([]string, len(tail)+1)
	for i := 0; i < len(tail); i++ {
		tmp[i+1] = tail[i]
	}
	tmp[0] = head
	return tmp
}

func (d *Parallels9Driver) Mac(vmName string) (string, error) {
	var stdout bytes.Buffer

	cmd := exec.Command(d.PrlctlPath, "list", "-i", vmName)
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		log.Printf("MAC address for NIC: nic0 on Virtual Machine: %s not found!\n", vmName)
		return "", err
	}

	stdoutString := strings.TrimSpace(stdout.String())
	re := regexp.MustCompile("net0.* mac=([0-9A-F]{12}) card=.*")
	macMatch := re.FindAllStringSubmatch(stdoutString, 1)

	if len(macMatch) != 1 {
		return "", fmt.Errorf("MAC address for NIC: nic0 on Virtual Machine: %s not found!\n", vmName)
	}

	mac := macMatch[0][1]
	log.Printf("Found MAC address for NIC: net0 - %s\n", mac)
	return mac, nil
}

// Finds the IP address of a VM connected that uses DHCP by its MAC address
func (d *Parallels9Driver) IpAddress(mac string) (string, error) {
	var stdout bytes.Buffer
	dhcp_lease_file := "/Library/Preferences/Parallels/parallels_dhcp_leases"

	if len(mac) != 12 {
		return "", fmt.Errorf("Not a valid MAC address: %s. It should be exactly 12 digits.", mac)
	}

	cmd := exec.Command("grep", "-i", mac, dhcp_lease_file)
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	stdoutString := strings.TrimSpace(stdout.String())
	re := regexp.MustCompile("(.*)=.*")
	ipMatch := re.FindAllStringSubmatch(stdoutString, 1)

	if len(ipMatch) != 1 {
		return "", fmt.Errorf("IP lease not found for MAC address %s in: %s\n", mac, dhcp_lease_file)
	}

	ip := ipMatch[0][1]
	log.Printf("Found IP lease: %s for MAC address %s\n", ip, mac)
	return ip, nil
}
