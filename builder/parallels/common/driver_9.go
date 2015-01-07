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

	"github.com/going/toolkit/xmlpath"
)

type Parallels9Driver struct {
	// This is the path to the "prlctl" application.
	PrlctlPath string
	// The path to the parallels_dhcp_leases file
	dhcp_lease_file string
}

func (d *Parallels9Driver) Import(name, srcPath, dstDir string, reassignMac bool) error {

	err := d.Prlctl("register", srcPath, "--preserve-uuid")
	if err != nil {
		return err
	}

	srcId, err := getVmId(srcPath)
	if err != nil {
		return err
	}

	srcMac := "auto"
	if !reassignMac {
		srcMac, err = getFirtsMacAddress(srcPath)
		if err != nil {
			return err
		}
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

// Finds an application bundle by identifier (for "darwin" platform only)
func getAppPath(bundleId string) (string, error) {
	var stdout bytes.Buffer

	cmd := exec.Command("mdfind", "kMDItemCFBundleIdentifier ==", bundleId)
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", err
	}

	pathOutput := strings.TrimSpace(stdout.String())
	if pathOutput == "" {
		return "", fmt.Errorf(
			"Could not detect Parallels Desktop! Make sure it is properly installed.")
	}

	return pathOutput, nil
}

func (d *Parallels9Driver) DeviceAddCdRom(name string, image string) (string, error) {
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

	device_name := matches[1]
	return device_name, nil
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
//
// Parses the file /Library/Preferences/Parallels/parallels_dhcp_leases
// file contain a list of DHCP leases given by Parallels Desktop
// Example line:
// 10.211.55.181="1418921112,1800,001c42f593fb,ff42f593fb000100011c25b9ff001c42f593fb"
// IP Address   ="Lease expiry, Lease time, MAC, MAC or DUID"
func (d *Parallels9Driver) IpAddress(mac string) (string, error) {

	if len(mac) != 12 {
		return "", fmt.Errorf("Not a valid MAC address: %s. It should be exactly 12 digits.", mac)
	}

	leases, err := ioutil.ReadFile(d.dhcp_lease_file)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile("(.*)=\"(.*),(.*)," + strings.ToLower(mac) + ",.*\"")
	mostRecentIp := ""
	mostRecentLease := uint64(0)
	for _, l := range re.FindAllStringSubmatch(string(leases), -1) {
		ip := l[1]
		expiry, _ := strconv.ParseUint(l[2], 10, 64)
		leaseTime, _ := strconv.ParseUint(l[3], 10, 32)
		log.Printf("Found lease: %s for MAC: %s, expiring at %d, leased for %d s.\n", ip, mac, expiry, leaseTime)
		if mostRecentLease <= expiry-leaseTime {
			mostRecentIp = ip
			mostRecentLease = expiry - leaseTime
		}
	}

	if len(mostRecentIp) == 0 {
		return "", fmt.Errorf("IP lease not found for MAC address %s in: %s\n", mac, d.dhcp_lease_file)
	}

	log.Printf("Found IP lease: %s for MAC address %s\n", mostRecentIp, mac)
	return mostRecentIp, nil
}

func (d *Parallels9Driver) ToolsIsoPath(k string) (string, error) {
	appPath, err := getAppPath("com.parallels.desktop.console")
	if err != nil {
		return "", err
	}

	toolsPath := filepath.Join(appPath, "Contents", "Resources", "Tools", "prl-tools-"+k+".iso")
	log.Printf("Parallels Tools path: '%s'", toolsPath)
	return toolsPath, nil
}
