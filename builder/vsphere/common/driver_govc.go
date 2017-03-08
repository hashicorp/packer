package common

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/packer/packer"

	"github.com/jmoiron/jsonq"
)

// GOVC driver talks to an ESXi hypervisor or VCenter remotely over
// API using govc (github.com/vmware/govmomi/govc) to build virtual
//machines.
//TODO: verify the following assumption
//This driver can only manage one machine at a time.
type GOVCDriver struct {
	Vcenter        string
	Host           string
	Datacenter     string
	Cluster        string
	ResourcePool   string
	Username       string
	Password       string
	Insecure       bool
	Datastore      string
	CacheDatastore string
	CacheFolder    string
	SSHConfig      *SSHConfig

	comm     packer.Communicator
	ui       packer.Ui
	govcPath string
	vmName   string
	vmPath   string
	hostPath string
}

func NewGOVCDriver() (Driver, error) {
	govcDriver := &GOVCDriver{}
	if err := govcDriver.Verify(); err != nil {
		return nil, err
	}
	return govcDriver, nil
}

func (d *GOVCDriver) govc(args []string) (string, error) {
	cmd := exec.Command(d.govcPath, args...)
	log.Printf("Calling: govc %s", strings.Join(args, " "))
	env := os.Environ()
	env = append(env, fmt.Sprintf("GOVC_URL=https://%s/sdk", d.Vcenter))
	env = append(env, fmt.Sprintf("GOVC_USERNAME=%s", d.Username))
	env = append(env, fmt.Sprintf("GOVC_PASSWORD=%s", d.Password))
	if d.Insecure {
		env = append(env, "GOVC_INSECURE=1")
	}
	cmd.Env = env

	stdout, _, err := runAndLog(cmd)
	if err != nil {
		err = fmt.Errorf("Govc return an error: %v", err)
		return "", err
	}
	return stdout, nil
}

func (d *GOVCDriver) vapiPathExists(path string) (bool, error) {
	re := regexp.MustCompile("/[^/]*/?$")
	searchpath := re.ReplaceAllString(path, "")
	gocmd := []string{
		"ls",
		searchpath,
	}
	stdout, err := d.govc(gocmd)
	if err != nil {
		err = fmt.Errorf("Could not list folders: %v", err)
		return false, err
	}
	content := strings.Split(stdout, "\n")
	for i := range content {
		if content[i] == path {
			return true, nil
		}
	}
	return false, nil
}

func (d *GOVCDriver) folderDatastoreNetworkExists(folder string, datastore string, network string) (string, error) {
	folderpath := fmt.Sprintf("/%s/vm/%s", d.Datacenter, folder)
	if folder == "" {
		folderpath = fmt.Sprintf("/%s/vm", d.Datacenter)
	}
	exists, err := d.vapiPathExists(folder)
	if err != nil || !exists {
		err = fmt.Errorf("The folder %s does not exists", folder)
		return "", err
	}
	datastorepath := fmt.Sprintf("/%s/datastore/%s", d.Datacenter, datastore)
	exists, err = d.vapiPathExists(datastorepath)
	if err != nil || !exists {
		err = fmt.Errorf("The datastore %s does not exists", datastore)
		return "", err
	}

	networkpath := fmt.Sprintf("/%s/network/%s", d.Datacenter, network)
	exists, err = d.vapiPathExists(networkpath)
	if err != nil || !exists {
		err = fmt.Errorf("The network %s does not exists", network)
		return "", err
	}

	return folderpath, nil
}

//TODO: Implement this function
func (d *GOVCDriver) CloneVirtualMachine(srcVmName string, dstVmName string, folder string, datastore string, cpu uint, ram uint, diskSize uint, diskThick bool, networkName string, networkAdapter string, annotation string) error {
	//govc vm.clone -vm SrcVmName -c=nbcpu (0=idem) -datastore-cluster=[GOVC_DATASTORE_CLUSTER] -ds=[GOVC_DATASTORE] -folder=[GOVC_FOLDER] -force=false -host=[GOVC_HOST] -m=Ramsize (0=idem) -net=[GOVC_NETWORK] -net.adapter=e1000 -net.address=IP -on=false -pool=[GOVC_RESOURCE_POOL] -annotation="description"

	//srcfolderpath, err := d.folderDatastoreNetworkExists(srcfolder, srcdatastore, srcnetworkName)
	//if err != nil {
	//		return err
	//	}
	//dstfolderpath, err := d.folderDatastoreNetworkExists(dstfolder, dstdatastore, dstnetworkName)
	//if err != nil {
	//		return err
	//	}
	return errors.New("Cloning is not implemented yet.")
}

func (d *GOVCDriver) CreateVirtualMachine(vmName string, folder string, datastore string, cpu uint, ram uint, diskSize uint, diskThick bool, guestType string, networkName string, networkAdapter string, annotation string) error {
	folderpath, err := d.folderDatastoreNetworkExists(folder, datastore, networkName)
	if err != nil {
		return err
	}

	gocmd := []string{
		"vm.create", "-force=false", "-on=false",
		fmt.Sprintf("-c=%d", cpu),                   // default 1
		fmt.Sprintf("-m=%d", ram),                   // in MB
		fmt.Sprintf("-ds=%s", datastore),            // datastore1
		fmt.Sprintf("-disk=%d", diskSize*1024*1024), // in MB (converted to B)
		fmt.Sprintf("-g=%s", guestType),             // otherGuest, rhel7_64Guest, ..
		fmt.Sprintf("-dc=%s", d.Datacenter),         // ha-datacenter
		fmt.Sprintf("-folder=%s", folderpath),       // /ha-datacenter/vm/MyGroup
		fmt.Sprintf("-host=%s", d.hostPath),         // /ha-datacenter/host/MyCluster/MyHost  or /ha-datacenter/host/MyHost (if cluster empty)
		fmt.Sprintf("-net=%s", networkName),         // MyNetwork
		//TODO: add validation on the value of networkAdapter
		fmt.Sprintf("-net.adapter=%s", networkAdapter), // e1000, vmxnet3, ...
	}

	//govc does not allow to select disk type for first disk
	//	if diskThick {
	//		gocmd = append(gocmd, "-thick=true")
	//	}

	if d.ResourcePool != "" {
		gocmd = append(gocmd, fmt.Sprintf("-pool=%s", d.ResourcePool))
	}

	//TODO: Implement for datastore cluster (if needed)
	//	if datastoreCluster != "" {
	//		gocmd = append(gocmd, fmt.Sprintf("-datastore-cluster=%s", datastoreCluster))
	//	}

	if annotation != "" {
		gocmd = append(gocmd, fmt.Sprintf("-annotation=%s", annotation))
	}

	gocmd = append(gocmd, vmName)

	if _, err := d.govc(gocmd); err != nil {
		err = fmt.Errorf("Could not create VM: %v", err)
		return err
	}
	d.vmName = vmName
	d.vmPath = fmt.Sprintf("%s/%s", folderpath, vmName)
	return nil
}

func (d *GOVCDriver) Destroy() error {
	gocmd := []string{
		"vm.destroy",
		d.vmPath,
	}

	if _, err := d.govc(gocmd); err != nil {
		err = fmt.Errorf("Could not destroy VM: %v", err)
		return err
	}
	return nil
}

func (d *GOVCDriver) CreateDisk(diskSize uint, diskThick bool) error {
	gocmd := []string{
		"vm.disk.create",
		" -vm", d.vmPath,
		fmt.Sprintf("-disk=%d", diskSize*1024*1024), // in MB (converted to B)
	}

	if diskThick {
		gocmd = append(gocmd, "-thick=true")
	}

	if _, err := d.govc(gocmd); err != nil {
		err = fmt.Errorf("Could not add disk to VM: %v", err)
		return err
	}
	return nil
}

func (d *GOVCDriver) ToolsInstall() error {
	gocmd := []string{
		"vm.guest.tools", "-mount",
		d.vmPath,
	}

	if _, err := d.govc(gocmd); err != nil {
		err = fmt.Errorf("Could not mount tools to VM: %v", err)
		return err
	}
	return nil
}

func (d *GOVCDriver) IsRunning() (bool, error) {
	gocmd := []string{
		"vm.info", "-json=true",
		d.vmPath,
	}

	stdout, err := d.govc(gocmd)
	if err != nil {
		err = fmt.Errorf("Could not retrieve status of VM: %v", err)
		return false, err
	}
	var response map[string]interface{}
	decoder := json.NewDecoder(strings.NewReader(stdout))
	decoder.Decode(&response)
	jq := jsonq.NewQuery(response)
	powerState, err := jq.String("VirtualMachines", "0", "Runtime", "PowerState")
	if err != nil {
		err = fmt.Errorf("Could not retrieve status of VM: %v", err)
		return false, err
	}
	return strings.Contains(powerState, "poweredOn"), nil
}

func (d *GOVCDriver) Start() error {
	gocmd := []string{
		"vm.power",
		"-on=true",
		d.vmPath,
	}
	for i := 0; i < 20; i++ {

		_, _ = d.govc(gocmd)
		//intentionally not checking for error since poweron may fail specially after initial VM registration
		time.Sleep((time.Duration(i) * time.Second) + 1)
		running, err := d.IsRunning()
		if err != nil {
			return err
		}
		if running {
			//Powerstate is on before the vm is really started
			time.Sleep(20 * time.Second)
			return nil
		}
	}
	return errors.New("Retry limit exceeded")
}

func (d *GOVCDriver) IsStopped() (bool, error) {
	gocmd := []string{
		"vm.info", "-json=true",
		d.vmPath,
	}

	stdout, err := d.govc(gocmd)
	if err != nil {
		err = fmt.Errorf("Could not retrieve status of VM: %v", err)
		return false, err
	}
	var response map[string]interface{}
	decoder := json.NewDecoder(strings.NewReader(stdout))
	decoder.Decode(&response)
	jq := jsonq.NewQuery(response)
	powerState, err := jq.String("VirtualMachines", "0", "Runtime", "PowerState")
	if err != nil {
		err = fmt.Errorf("Could not retrieve status of VM: %v", err)
		return false, err
	}
	return strings.Contains(powerState, "poweredOff"), nil
}

func (d *GOVCDriver) Stop() error {
	gocmd := []string{
		"vm.power",
		"-off=true",
		d.vmPath,
	}
	if _, err := d.govc(gocmd); err != nil {
		err = fmt.Errorf("Could not stop the VM: %v", err)
		return err
	}
	return nil
}

func (d *GOVCDriver) IsDestroyed() (bool, error) {
	gocmd := []string{
		"ls",
		d.vmPath,
	}
	stdout, err := d.govc(gocmd)
	if err != nil {
		err = fmt.Errorf("Could not retrieve status of VM: %v", err)
		return false, err
	}
	if stdout == d.vmPath {
		return false, nil
	}
	return true, nil
}

func (d *GOVCDriver) Upload(localPath string, remoteFilename string) (string, error) {
	gocmd := []string{
		"datastore.ls",
		fmt.Sprintf("-ds=%s", d.CacheDatastore),
		d.CacheFolder,
	}
	if _, err := d.govc(gocmd); err != nil {
		gocmd = []string{
			"datastore.mkdir",
			fmt.Sprintf("-ds=%s", d.CacheDatastore),
			d.CacheFolder,
		}
		if _, err := d.govc(gocmd); err != nil {
			err = fmt.Errorf("Could not create remote directory: %v", err)
			return "", err
		}
	}
	var remotePath = filepath.ToSlash(filepath.Join(d.CacheFolder, remoteFilename))
	gocmd = []string{
		"datastore.upload",
		fmt.Sprintf("-ds=%s", d.CacheDatastore),
		localPath,
		remotePath,
	}
	if _, err := d.govc(gocmd); err != nil {
		err = fmt.Errorf("Could not upload to remote directory: %v", err)
		return "", err
	}
	return remotePath, nil
}

func (d *GOVCDriver) generateOvftoolArgs(format string, outputPath string, ovftoolOptions []string, hidePassword bool) []string {
	password := url.QueryEscape(d.Password)
	if hidePassword {
		password = "****"
	}
	vmUri := d.vmPath[len(d.Datacenter)+len("//vm/") : len(d.vmPath)]
	args := []string{
		"--noSSLVerify=true",
		"--skipManifestCheck",
		"-tt=" + format,
		"vi://" + d.Username + ":" + password + "@" + d.Vcenter + "/" + vmUri,
		outputPath,
	}
	return append(ovftoolOptions, args...)
}

func (d *GOVCDriver) ExportVirtualMachine(localpath string, format string, ovftooloptions []string) error {
	ovftool := "ovftool"
	if runtime.GOOS == "windows" {
		ovftool = "ovftool.exe"
	}

	if _, err := exec.LookPath(ovftool); err != nil {
		err := fmt.Errorf("Error %s not found: %s", ovftool, err)
		return err
	}

	log.Printf("Executing: %s %s", ovftool, strings.Join(d.generateOvftoolArgs(format, localpath, ovftooloptions, true), " "))
	var out bytes.Buffer
	cmd := exec.Command(ovftool, d.generateOvftoolArgs(format, localpath, ovftooloptions, false)...)
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		err := fmt.Errorf("Error exporting virtual machine: %s\n%s\n", err, out.String())
		return err
	}

	log.Printf("%s", out.String())
	return nil
}

func (d *GOVCDriver) AddFloppy(floppyFilename string) (string, error) {
	gocmd := []string{
		"device.floppy.add",
		"-vm", d.vmPath,
	}
	floppyDevice, err := d.govc(gocmd)
	if err != nil {
		err = fmt.Errorf("Could not add floppy drive to VM: %v", err)
		return "", err
	}
	gocmd = []string{
		"device.floppy.insert",
		fmt.Sprintf("-ds=%s", d.CacheDatastore),
		"-vm", d.vmPath,
		floppyFilename,
	}
	if _, err := d.govc(gocmd); err != nil {
		err = fmt.Errorf("Could not insert floppy in VM: %v", err)
		return "", err
	}
	return strings.TrimSpace(floppyDevice), nil
}

func (d *GOVCDriver) RemoveFloppy(floppyDevice string) error {
	gocmd := []string{
		"device.floppy.eject",
		"-vm", d.vmPath,
	}
	if _, err := d.govc(gocmd); err != nil {
		err = fmt.Errorf("Could not eject floppy from VM: %v", err)
		return err
	}
	gocmd = []string{
		"device.remove",
		"-vm", d.vmPath,
		floppyDevice,
	}
	if _, err := d.govc(gocmd); err != nil {
		err = fmt.Errorf("Could not remove floppy drive from VM: %v", err)
		return err
	}
	return nil
}

func (d *GOVCDriver) MountISO(isoFilename string) (string, error) {
	gocmd := []string{
		"device.cdrom.add",
		"-vm", d.vmPath,
	}
	cdromDevice, err := d.govc(gocmd)
	if err != nil {
		err = fmt.Errorf("Could not add cdrom drive to VM: %v", err)
		return "", err
	}
	gocmd = []string{
		"device.cdrom.insert",
		fmt.Sprintf("-ds=%s", d.CacheDatastore),
		"-vm", d.vmPath,
		isoFilename,
	}
	if _, err := d.govc(gocmd); err != nil {
		err = fmt.Errorf("Could not insert cdrom in VM: %v", err)
		return "", err
	}
	return strings.TrimSpace(cdromDevice), nil
}

func (d *GOVCDriver) UnmountISO(cdromDevice string) error {
	gocmd := []string{
		"device.cdrom.eject",
		"-vm", d.vmPath,
		"-device", cdromDevice,
	}
	if _, err := d.govc(gocmd); err != nil {
		err = fmt.Errorf("Could not eject cdrom from VM: %v", err)
		return err
	}
	gocmd = []string{
		"device.remove",
		"-vm", d.vmPath,
		cdromDevice,
	}
	if _, err := d.govc(gocmd); err != nil {
		err = fmt.Errorf("Could not remove cdrom drive from VM: %v", err)
		return err
	}
	return nil
}

func (d *GOVCDriver) VMChange(change string) error {
	gocmd := []string{
		"vm.change",
		"-vm", d.vmName,
		"-e", change,
	}
	if _, err := d.govc(gocmd); err != nil {
		err = fmt.Errorf("Could not VM configuration: %v", err)
		return err
	}
	return nil
}

func (d *GOVCDriver) VNCDisable() error {
	gocmd := []string{
		"vm.vnc",
		"-disable=true",
		d.vmName,
	}
	if _, err := d.govc(gocmd); err != nil {
		err = fmt.Errorf("Could not disable VNC configuration: %v", err)
		return err
	}
	return nil
}

func (d *GOVCDriver) VNCEnable(password string, portMin, portMax uint) (string, uint, error) {
	gocmd := []string{
		"vm.vnc",
		"-enable=true",
		fmt.Sprintf("-port-range=%d-%d", portMin, portMax),
		d.vmName,
	}
	if len(password) > 0 {
		gocmd = append(gocmd, fmt.Sprintf("-password=%s", password))
	}
	stdout, err := d.govc(gocmd)
	if err != nil {
		err = fmt.Errorf("Could not enable VNC configuration: %v", err)
		return "", 0, err
	}
	url := strings.Split(strings.TrimSpace(stdout), "@")
	host, vncPort, err := net.SplitHostPort(url[len(url)-1])
	convVncPort, err := strconv.ParseUint(vncPort, 10, 32)
	return host, uint(convVncPort), err
}

//Return the IP adress of the VM to be used by communicator
func (d *GOVCDriver) GuestIP() (string, error) {
	//TODO: This call will wait until VM has an IP govc vm.ip when the issue #678 of govc will be resolved
	gocmd := []string{
		"vm.ip",
		"-esxcli=true",
		d.vmName,
	}
	stdout, err := d.govc(gocmd)
	if err != nil {
		err = fmt.Errorf("Could not VM IP: %v", err)
		return "", err
	}
	address := strings.TrimSpace(stdout)
	return address, nil
}

func (d *GOVCDriver) Verify() error {
	checks := []func() error{
		//Verify the VM does not already exists (option force will destroy/recreate the VM ?)
		d.govcFind,
		d.checkSystemVersion,
		d.checkGuestIPHackEnabled,
		d.checkHostpathExists,
		d.checkCacheDatastoreExists,
	}

	for _, check := range checks {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

func (d *GOVCDriver) govcFind() error {
	var govcPath string
	var err error

	if d.govcPath == "" {
		govc := "govc"
		if runtime.GOOS == "windows" {
			govc = "govc.exe"
		}

		if govcPath, err = exec.LookPath(govc); err != nil {
			err := fmt.Errorf("Error %s not found: %s", govc, err)
			return err
		}
		d.govcPath = govcPath
	}
	return nil
}

func (d *GOVCDriver) checkSystemVersion() error {
	gocmd := []string{
		"about",
		"-json=true",
	}

	stdout, err := d.govc(gocmd)
	if err != nil {
		err = fmt.Errorf("Could not retrieve information from Vcenter/ESXi: %v", err)
		return err
	}
	var response map[string]interface{}
	decoder := json.NewDecoder(strings.NewReader(stdout))
	decoder.Decode(&response)
	jq := jsonq.NewQuery(response)
	product, err := jq.String("About", "Name")
	if err != nil {
		err = fmt.Errorf("Could not retrieve product name from Vcenter/ESXi: %v", err)
		return err
	}

	version, err := jq.String("About", "Version")
	if err != nil {
		err = fmt.Errorf("Could not retrieve version from Vcenter/ESXi: %v", err)
		return err
	}

	build, err := jq.String("About", "Build")
	if err != nil {
		err = fmt.Errorf("Could not retrieve build from Vcenter/ESXi: %v", err)
		return err
	}

	log.Printf("Connected to %s %s %s", product, version, build)
	return nil
}

func (d *GOVCDriver) checkGuestIPHackEnabled() error {
	gocmd := []string{
		"host.esxcli", "-json=true",
		"system", "settings", "advanced",
		"list", "-o", "/Net/GuestIPHack",
	}

	stdout, err := d.govc(gocmd)
	if err != nil {
		return err
	}
	var response map[string]interface{}
	decoder := json.NewDecoder(strings.NewReader(stdout))
	decoder.Decode(&response)
	jq := jsonq.NewQuery(response)
	netIPHack, err := jq.String("Values", "0", "IntValue", "0")
	if err != nil {
		err = fmt.Errorf("Could not retrieve advanced options from Vcenter/ESXi: %v", err)
		return err
	}
	if netIPHack != "1" {
		return errors.New(
			"GuestIPHack is required, enable by running this on the ESX machine:\n" +
				"esxcli system settings advanced set -o /Net/GuestIPHack -i 1")
	}

	return nil
}

func (d *GOVCDriver) checkHostpathExists() error {
	hostpath := fmt.Sprintf("/%s/host/%s", d.Datacenter, d.Host)
	if d.Cluster != "" {
		hostpath = fmt.Sprintf("/%s/host/%s/%s", d.Datacenter, d.Cluster, d.Host)
		//TODO: If the host is empty, its seems that we can create the VM (vcenter will choose the host in the cluster ??)
	}

	exists, err := d.vapiPathExists(hostpath)
	if err != nil || !exists {
		err = fmt.Errorf("Datacenter, cluster or host does not exists (%s): %v", hostpath, err)
		return nil
	}
	d.hostPath = hostpath
	return nil
}

func (d *GOVCDriver) checkCacheDatastoreExists() error {
	datastorepath := fmt.Sprintf("/%s/datastore/%s", d.Datacenter, d.CacheDatastore)
	exists, err := d.vapiPathExists(datastorepath)
	if err != nil || !exists {
		err = fmt.Errorf("The cache datastore %s does not exists", d.CacheDatastore)
		return err
	}
	return nil
}
