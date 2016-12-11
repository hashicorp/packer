package common

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/xmlpath.v2"
)

type Parallels9Driver struct {
	// This is the path to the "prlctl" application.
	PrlctlPath string

	// This is the path to the "prlsrvctl" application.
	PrlsrvctlPath string

	// The path to the parallels_dhcp_leases file
	dhcpLeaseFile string
}

func (d *Parallels9Driver) Import(name, srcPath, dstDir string, reassignMAC bool) error {

	err := d.Prlctl("register", srcPath, "--preserve-uuid")
	if err != nil {
		return err
	}

	srcID, err := getVMID(srcPath)
	if err != nil {
		return err
	}

	srcMAC := "auto"
	if !reassignMAC {
		srcMAC, err = getFirtsMACAddress(srcPath)
		if err != nil {
			return err
		}
	}

	err = d.Prlctl("clone", srcID, "--name", name, "--dst", dstDir)
	if err != nil {
		return err
	}

	err = d.Prlctl("unregister", srcID)
	if err != nil {
		return err
	}

	err = d.Prlctl("set", name, "--device-set", "net0", "--mac", srcMAC)
	if err != nil {
		return err
	}
	return nil
}

func getVMID(path string) (string, error) {
	return getConfigValueFromXpath(path, "/ParallelsVirtualMachine/Identification/VmUuid")
}

func getFirtsMACAddress(path string) (string, error) {
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

// Finds an application bundle by identifier (for "darwin" platform only)
func getAppPath(bundleID string) (string, error) {
	var stdout bytes.Buffer

	cmd := exec.Command("mdfind", "kMDItemCFBundleIdentifier ==", bundleID)
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	pathOutput := strings.TrimSpace(stdout.String())
	if pathOutput == "" {
		if fi, err := os.Stat("/Applications/Parallels Desktop.app"); err == nil {
			if fi.IsDir() {
				return "/Applications/Parallels Desktop.app", nil
			}
		}

		return "", fmt.Errorf(
			"Could not detect Parallels Desktop! Make sure it is properly installed.")
	}

	return pathOutput, nil
}

func (d *Parallels9Driver) CompactDisk(diskPath string) error {
	prlDiskToolPath, err := exec.LookPath("prl_disk_tool")
	if err != nil {
		return err
	}

	// Analyze the disk content and remove unused blocks
	command := []string{
		"compact",
		"--hdd", diskPath,
	}
	if err := exec.Command(prlDiskToolPath, command...).Run(); err != nil {
		return err
	}

	// Remove null blocks
	command = []string{
		"compact", "--buildmap",
		"--hdd", diskPath,
	}
	if err := exec.Command(prlDiskToolPath, command...).Run(); err != nil {
		return err
	}

	return nil
}

func (d *Parallels9Driver) DeviceAddCDROM(name string, image string) (string, error) {
	command := []string{
		"set", name,
		"--device-add", "cdrom",
		"--image", image,
	}

	out, err := exec.Command(d.PrlctlPath, command...).Output()
	if err != nil {
		return "", err
	}

	deviceRe := regexp.MustCompile(`\s+(cdrom\d+)\s+`)
	matches := deviceRe.FindStringSubmatch(string(out))
	if matches == nil {
		return "", fmt.Errorf(
			"Could not determine cdrom device name in the output:\n%s", string(out))
	}

	deviceName := matches[1]
	return deviceName, nil
}

func (d *Parallels9Driver) DiskPath(name string) (string, error) {
	out, err := exec.Command(d.PrlctlPath, "list", "-i", name).Output()
	if err != nil {
		return "", err
	}

	HDDRe := regexp.MustCompile("hdd0.* image='(.*)' type=*")
	matches := HDDRe.FindStringSubmatch(string(out))
	if matches == nil {
		return "", fmt.Errorf(
			"Could not determine hdd image path in the output:\n%s", string(out))
	}

	HDDPath := matches[1]
	return HDDPath, nil
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
	return nil
}

func (d *Parallels9Driver) Version() (string, error) {
	out, err := exec.Command(d.PrlctlPath, "--version").Output()
	if err != nil {
		return "", err
	}

	versionRe := regexp.MustCompile(`prlctl version (\d+\.\d+.\d+)`)
	matches := versionRe.FindStringSubmatch(string(out))
	if matches == nil {
		return "", fmt.Errorf(
			"Could not find Parallels Desktop version in output:\n%s", string(out))
	}

	version := matches[1]
	log.Printf("Parallels Desktop version: %s", version)
	return version, nil
}

func (d *Parallels9Driver) SendKeyScanCodes(vmName string, codes ...string) error {
	var stdout, stderr bytes.Buffer

	if codes == nil || len(codes) == 0 {
		log.Printf("No scan codes to send")
		return nil
	}

	f, err := ioutil.TempFile("", "prltype")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	script := []byte(Prltype)
	_, err = f.Write(script)
	if err != nil {
		return err
	}

	args := prepend(vmName, codes)
	args = prepend(f.Name(), args)
	cmd := exec.Command("/usr/bin/python", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()

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

func (d *Parallels9Driver) SetDefaultConfiguration(vmName string) error {
	commands := make([][]string, 7)
	commands[0] = []string{"set", vmName, "--cpus", "1"}
	commands[1] = []string{"set", vmName, "--memsize", "512"}
	commands[2] = []string{"set", vmName, "--startup-view", "same"}
	commands[3] = []string{"set", vmName, "--on-shutdown", "close"}
	commands[4] = []string{"set", vmName, "--on-window-close", "keep-running"}
	commands[5] = []string{"set", vmName, "--auto-share-camera", "off"}
	commands[6] = []string{"set", vmName, "--smart-guard", "off"}

	for _, command := range commands {
		err := d.Prlctl(command...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Parallels9Driver) MAC(vmName string) (string, error) {
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

// IPAddress finds the IP address of a VM connected that uses DHCP by its MAC address
//
// Parses the file /Library/Preferences/Parallels/parallels_dhcp_leases
// file contain a list of DHCP leases given by Parallels Desktop
// Example line:
// 10.211.55.181="1418921112,1800,001c42f593fb,ff42f593fb000100011c25b9ff001c42f593fb"
// IP Address   ="Lease expiry, Lease time, MAC, MAC or DUID"
func (d *Parallels9Driver) IPAddress(mac string) (string, error) {

	if len(mac) != 12 {
		return "", fmt.Errorf("Not a valid MAC address: %s. It should be exactly 12 digits.", mac)
	}

	leases, err := ioutil.ReadFile(d.dhcpLeaseFile)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile("(.*)=\"(.*),(.*)," + strings.ToLower(mac) + ",.*\"")
	mostRecentIP := ""
	mostRecentLease := uint64(0)
	for _, l := range re.FindAllStringSubmatch(string(leases), -1) {
		ip := l[1]
		expiry, _ := strconv.ParseUint(l[2], 10, 64)
		leaseTime, _ := strconv.ParseUint(l[3], 10, 32)
		log.Printf("Found lease: %s for MAC: %s, expiring at %d, leased for %d s.\n", ip, mac, expiry, leaseTime)
		if mostRecentLease <= expiry-leaseTime {
			mostRecentIP = ip
			mostRecentLease = expiry - leaseTime
		}
	}

	if len(mostRecentIP) == 0 {
		return "", fmt.Errorf("IP lease not found for MAC address %s in: %s\n", mac, d.dhcpLeaseFile)
	}

	log.Printf("Found IP lease: %s for MAC address %s\n", mostRecentIP, mac)
	return mostRecentIP, nil
}

func (d *Parallels9Driver) ToolsISOPath(k string) (string, error) {
	appPath, err := getAppPath("com.parallels.desktop.console")
	if err != nil {
		return "", err
	}

	toolsPath := filepath.Join(appPath, "Contents", "Resources", "Tools", "prl-tools-"+k+".iso")
	log.Printf("Parallels Tools path: '%s'", toolsPath)
	return toolsPath, nil
}
