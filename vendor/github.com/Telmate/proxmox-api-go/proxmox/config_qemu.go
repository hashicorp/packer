package proxmox

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type (
	QemuDevices     map[int]map[string]interface{}
	QemuDevice      map[string]interface{}
	QemuDeviceParam []string
)

// ConfigQemu - Proxmox API QEMU options
type ConfigQemu struct {
	Name         string      `json:"name"`
	Description  string      `json:"desc"`
	Onboot       bool        `json:"onboot"`
	Agent        string      `json:"agent"`
	Memory       int         `json:"memory"`
	QemuOs       string      `json:"os"`
	QemuCores    int         `json:"cores"`
	QemuSockets  int         `json:"sockets"`
	QemuIso      string      `json:"iso"`
	FullClone    *int        `json:"fullclone"`
	QemuDisks    QemuDevices `json:"disk"`
	QemuNetworks QemuDevices `json:"network"`

	// Deprecated single disk.
	DiskSize    float64 `json:"diskGB"`
	Storage     string  `json:"storage"`
	StorageType string  `json:"storageType"` // virtio|scsi (cloud-init defaults to scsi)

	// Deprecated single nic.
	QemuNicModel string `json:"nic"`
	QemuBrige    string `json:"bridge"`
	QemuVlanTag  int    `json:"vlan"`
	QemuMacAddr  string `json:"mac"`

	// cloud-init options
	CIuser     string `json:"ciuser"`
	CIpassword string `json:"cipassword"`

	Searchdomain string `json:"searchdomain"`
	Nameserver   string `json:"nameserver"`
	Sshkeys      string `json:"sshkeys"`

	// arrays are hard, support 2 interfaces for now
	Ipconfig0 string `json:"ipconfig0"`
	Ipconfig1 string `json:"ipconfig1"`
}

// CreateVm - Tell Proxmox API to make the VM
func (config ConfigQemu) CreateVm(vmr *VmRef, client *Client) (err error) {
	if config.HasCloudInit() {
		return errors.New("Cloud-init parameters only supported on clones or updates")
	}
	vmr.SetVmType("qemu")

	params := map[string]interface{}{
		"vmid":        vmr.vmId,
		"name":        config.Name,
		"onboot":      config.Onboot,
		"agent":       config.Agent,
		"ide2":        config.QemuIso + ",media=cdrom",
		"ostype":      config.QemuOs,
		"sockets":     config.QemuSockets,
		"cores":       config.QemuCores,
		"cpu":         "host",
		"memory":      config.Memory,
		"description": config.Description,
	}

	// Create disks config.
	config.CreateQemuDisksParams(vmr.vmId, params, false)

	// Create networks config.
	config.CreateQemuNetworksParams(vmr.vmId, params)

	exitStatus, err := client.CreateQemuVm(vmr.node, params)
	if err != nil {
		return fmt.Errorf("Error creating VM: %v, error status: %s (params: %v)", err, exitStatus, params)
	}
	return
}

// HasCloudInit - are there cloud-init options?
func (config ConfigQemu) HasCloudInit() bool {
	return config.CIuser != "" ||
		config.CIpassword != "" ||
		config.Searchdomain != "" ||
		config.Nameserver != "" ||
		config.Sshkeys != "" ||
		config.Ipconfig0 != "" ||
		config.Ipconfig1 != ""
}

/*

CloneVm
Example: Request

nodes/proxmox1-xx/qemu/1012/clone

newid:145
name:tf-clone1
target:proxmox1-xx
full:1
storage:xxx

*/
func (config ConfigQemu) CloneVm(sourceVmr *VmRef, vmr *VmRef, client *Client) (err error) {
	vmr.SetVmType("qemu")
	fullclone := "1"
	if config.FullClone != nil {
		fullclone = strconv.Itoa(*config.FullClone)
	}
	storage := config.Storage
	if disk0Storage, ok := config.QemuDisks[0]["storage"].(string); ok && len(disk0Storage) > 0 {
		storage = disk0Storage
	}
	params := map[string]interface{}{
		"newid":   vmr.vmId,
		"target":  vmr.node,
		"name":    config.Name,
		"storage": storage,
		"full":    fullclone,
	}
	_, err = client.CloneQemuVm(sourceVmr, params)
	if err != nil {
		return
	}
	return config.UpdateConfig(vmr, client)
}

func (config ConfigQemu) UpdateConfig(vmr *VmRef, client *Client) (err error) {
	configParams := map[string]interface{}{
		"name":        config.Name,
		"description": config.Description,
		"onboot":      config.Onboot,
		"agent":       config.Agent,
		"sockets":     config.QemuSockets,
		"cores":       config.QemuCores,
		"memory":      config.Memory,
	}

	// Create disks config.
	config.CreateQemuDisksParams(vmr.vmId, configParams, true)

	// Create networks config.
	config.CreateQemuNetworksParams(vmr.vmId, configParams)

	// cloud-init options
	if config.CIuser != "" {
		configParams["ciuser"] = config.CIuser
	}
	if config.CIpassword != "" {
		configParams["cipassword"] = config.CIpassword
	}
	if config.Searchdomain != "" {
		configParams["searchdomain"] = config.Searchdomain
	}
	if config.Nameserver != "" {
		configParams["nameserver"] = config.Nameserver
	}
	if config.Sshkeys != "" {
		sshkeyEnc := url.PathEscape(config.Sshkeys + "\n")
		sshkeyEnc = strings.Replace(sshkeyEnc, "+", "%2B", -1)
		sshkeyEnc = strings.Replace(sshkeyEnc, "@", "%40", -1)
		sshkeyEnc = strings.Replace(sshkeyEnc, "=", "%3D", -1)
		configParams["sshkeys"] = sshkeyEnc
	}
	if config.Ipconfig0 != "" {
		configParams["ipconfig0"] = config.Ipconfig0
	}
	if config.Ipconfig1 != "" {
		configParams["ipconfig1"] = config.Ipconfig1
	}
	_, err = client.SetVmConfig(vmr, configParams)
	return err
}

func NewConfigQemuFromJson(io io.Reader) (config *ConfigQemu, err error) {
	config = &ConfigQemu{QemuVlanTag: -1}
	err = json.NewDecoder(io).Decode(config)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	log.Println(config)
	return
}

var (
	rxIso      = regexp.MustCompile(`(.*?),media`)
	rxDeviceID = regexp.MustCompile(`\d+`)
	rxDiskName = regexp.MustCompile(`(virtio|scsi)\d+`)
	rxDiskType = regexp.MustCompile(`\D+`)
	rxNicName  = regexp.MustCompile(`net\d+`)
)

func NewConfigQemuFromApi(vmr *VmRef, client *Client) (config *ConfigQemu, err error) {
	var vmConfig map[string]interface{}
	for ii := 0; ii < 3; ii++ {
		vmConfig, err = client.GetVmConfig(vmr)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		// this can happen:
		// {"data":{"lock":"clone","digest":"eb54fb9d9f120ba0c3bdf694f73b10002c375c38","description":" qmclone temporary file\n"}})
		if vmConfig["lock"] == nil {
			break
		} else {
			time.Sleep(8 * time.Second)
		}
	}

	if vmConfig["lock"] != nil {
		return nil, errors.New("vm locked, could not obtain config")
	}

	// vmConfig Sample: map[ cpu:host
	// net0:virtio=62:DF:XX:XX:XX:XX,bridge=vmbr0
	// ide2:local:iso/xxx-xx.iso,media=cdrom memory:2048
	// smbios1:uuid=8b3bf833-aad8-4545-xxx-xxxxxxx digest:aa6ce5xxxxx1b9ce33e4aaeff564d4 sockets:1
	// name:terraform-ubuntu1404-template bootdisk:virtio0
	// virtio0:ProxmoxxxxISCSI:vm-1014-disk-2,size=4G
	// description:Base image
	// cores:2 ostype:l26

	name := ""
	if _, isSet := vmConfig["name"]; isSet {
		name = vmConfig["name"].(string)
	}
	description := ""
	if _, isSet := vmConfig["description"]; isSet {
		description = vmConfig["description"].(string)
	}
	onboot := true
	if _, isSet := vmConfig["onboot"]; isSet {
		onboot = Itob(int(vmConfig["onboot"].(float64)))
	}
	agent := "1"
	if _, isSet := vmConfig["agent"]; isSet {
		agent = vmConfig["agent"].(string)
	}
	ostype := "other"
	if _, isSet := vmConfig["ostype"]; isSet {
		ostype = vmConfig["ostype"].(string)
	}
	memory := 0.0
	if _, isSet := vmConfig["memory"]; isSet {
		memory = vmConfig["memory"].(float64)
	}
	cores := 1.0
	if _, isSet := vmConfig["cores"]; isSet {
		cores = vmConfig["cores"].(float64)
	}
	sockets := 1.0
	if _, isSet := vmConfig["sockets"]; isSet {
		sockets = vmConfig["sockets"].(float64)
	}
	config = &ConfigQemu{
		Name:         name,
		Description:  strings.TrimSpace(description),
		Onboot:       onboot,
		Agent:        agent,
		QemuOs:       ostype,
		Memory:       int(memory),
		QemuCores:    int(cores),
		QemuSockets:  int(sockets),
		QemuVlanTag:  -1,
		QemuDisks:    QemuDevices{},
		QemuNetworks: QemuDevices{},
	}

	if vmConfig["ide2"] != nil {
		isoMatch := rxIso.FindStringSubmatch(vmConfig["ide2"].(string))
		config.QemuIso = isoMatch[1]
	}

	if _, isSet := vmConfig["ciuser"]; isSet {
		config.CIuser = vmConfig["ciuser"].(string)
	}
	if _, isSet := vmConfig["cipassword"]; isSet {
		config.CIpassword = vmConfig["cipassword"].(string)
	}
	if _, isSet := vmConfig["searchdomain"]; isSet {
		config.Searchdomain = vmConfig["searchdomain"].(string)
	}
	if _, isSet := vmConfig["sshkeys"]; isSet {
		config.Sshkeys, _ = url.PathUnescape(vmConfig["sshkeys"].(string))
	}
	if _, isSet := vmConfig["ipconfig0"]; isSet {
		config.Ipconfig0 = vmConfig["ipconfig0"].(string)
	}
	if _, isSet := vmConfig["ipconfig1"]; isSet {
		config.Ipconfig1 = vmConfig["ipconfig1"].(string)
	}

	// Add disks.
	diskNames := []string{}

	for k, _ := range vmConfig {
		if diskName := rxDiskName.FindStringSubmatch(k); len(diskName) > 0 {
			diskNames = append(diskNames, diskName[0])
		}
	}

	for _, diskName := range diskNames {
		diskConfStr := vmConfig[diskName]
		diskConfList := strings.Split(diskConfStr.(string), ",")

		//
		id := rxDeviceID.FindStringSubmatch(diskName)
		diskID, _ := strconv.Atoi(id[0])
		diskType := rxDiskType.FindStringSubmatch(diskName)[0]
		storageName, fileName := ParseSubConf(diskConfList[0], ":")

		//
		diskConfMap := QemuDevice{
			"type":    diskType,
			"storage": storageName,
			"file":    fileName,
		}

		// Add rest of device config.
		diskConfMap.readDeviceConfig(diskConfList[1:])

		// And device config to disks map.
		if len(diskConfMap) > 0 {
			config.QemuDisks[diskID] = diskConfMap
		}
	}

	// Add networks.
	nicNameRe := regexp.MustCompile(`net\d+`)
	nicNames := []string{}

	for k, _ := range vmConfig {
		if nicName := nicNameRe.FindStringSubmatch(k); len(nicName) > 0 {
			nicNames = append(nicNames, nicName[0])
		}
	}

	for _, nicName := range nicNames {
		nicConfStr := vmConfig[nicName]
		nicConfList := strings.Split(nicConfStr.(string), ",")

		//
		id := rxDeviceID.FindStringSubmatch(nicName)
		nicID, _ := strconv.Atoi(id[0])
		model, macaddr := ParseSubConf(nicConfList[0], "=")

		// Add model and MAC address.
		nicConfMap := QemuDevice{
			"model":   model,
			"macaddr": macaddr,
		}

		// Add rest of device config.
		nicConfMap.readDeviceConfig(nicConfList[1:])

		// And device config to networks.
		if len(nicConfMap) > 0 {
			config.QemuNetworks[nicID] = nicConfMap
		}
	}

	return
}

// Useful waiting for ISO install to complete
func WaitForShutdown(vmr *VmRef, client *Client) (err error) {
	for ii := 0; ii < 100; ii++ {
		vmState, err := client.GetVmState(vmr)
		if err != nil {
			log.Print("Wait error:")
			log.Println(err)
		} else if vmState["status"] == "stopped" {
			return nil
		}
		time.Sleep(5 * time.Second)
	}
	return errors.New("Not shutdown within wait time")
}

// This is because proxmox create/config API won't let us make usernet devices
func SshForwardUsernet(vmr *VmRef, client *Client) (sshPort string, err error) {
	vmState, err := client.GetVmState(vmr)
	if err != nil {
		return "", err
	}
	if vmState["status"] == "stopped" {
		return "", errors.New("VM must be running first")
	}
	sshPort = strconv.Itoa(vmr.VmId() + 22000)
	_, err = client.MonitorCmd(vmr, "netdev_add user,id=net1,hostfwd=tcp::"+sshPort+"-:22")
	if err != nil {
		return "", err
	}
	_, err = client.MonitorCmd(vmr, "device_add virtio-net-pci,id=net1,netdev=net1,addr=0x13")
	if err != nil {
		return "", err
	}
	return
}

// device_del net1
// netdev_del net1
func RemoveSshForwardUsernet(vmr *VmRef, client *Client) (err error) {
	vmState, err := client.GetVmState(vmr)
	if err != nil {
		return err
	}
	if vmState["status"] == "stopped" {
		return errors.New("VM must be running first")
	}
	_, err = client.MonitorCmd(vmr, "device_del net1")
	if err != nil {
		return err
	}
	_, err = client.MonitorCmd(vmr, "netdev_del net1")
	if err != nil {
		return err
	}
	return nil
}

func MaxVmId(client *Client) (max int, err error) {
	resp, err := client.GetVmList()
	vms := resp["data"].([]interface{})
	max = 0
	for vmii := range vms {
		vm := vms[vmii].(map[string]interface{})
		vmid := int(vm["vmid"].(float64))
		if vmid > max {
			max = vmid
		}
	}
	return
}

func SendKeysString(vmr *VmRef, client *Client, keys string) (err error) {
	vmState, err := client.GetVmState(vmr)
	if err != nil {
		return err
	}
	if vmState["status"] == "stopped" {
		return errors.New("VM must be running first")
	}
	for _, r := range keys {
		c := string(r)
		lower := strings.ToLower(c)
		if c != lower {
			c = "shift-" + lower
		} else {
			switch c {
			case "!":
				c = "shift-1"
			case "@":
				c = "shift-2"
			case "#":
				c = "shift-3"
			case "$":
				c = "shift-4"
			case "%%":
				c = "shift-5"
			case "^":
				c = "shift-6"
			case "&":
				c = "shift-7"
			case "*":
				c = "shift-8"
			case "(":
				c = "shift-9"
			case ")":
				c = "shift-0"
			case "_":
				c = "shift-minus"
			case "+":
				c = "shift-equal"
			case " ":
				c = "spc"
			case "/":
				c = "slash"
			case "\\":
				c = "backslash"
			case ",":
				c = "comma"
			case "-":
				c = "minus"
			case "=":
				c = "equal"
			case ".":
				c = "dot"
			case "?":
				c = "shift-slash"
			}
		}
		_, err = client.MonitorCmd(vmr, "sendkey "+c)
		if err != nil {
			return err
		}
		time.Sleep(100)
	}
	return nil
}

// Create parameters for each Nic device.
func (c ConfigQemu) CreateQemuNetworksParams(vmID int, params map[string]interface{}) error {

	// For backward compatibility.
	if len(c.QemuNetworks) == 0 && len(c.QemuNicModel) > 0 {
		deprecatedStyleMap := QemuDevice{
			"model":   c.QemuNicModel,
			"bridge":  c.QemuBrige,
			"macaddr": c.QemuMacAddr,
		}

		if c.QemuVlanTag > 0 {
			deprecatedStyleMap["tag"] = strconv.Itoa(c.QemuVlanTag)
		}

		c.QemuNetworks[0] = deprecatedStyleMap
	}

	// For new style with multi net device.
	for nicID, nicConfMap := range c.QemuNetworks {

		nicConfParam := QemuDeviceParam{}

		// Set Nic name.
		qemuNicName := "net" + strconv.Itoa(nicID)

		// Set Mac address.
		if nicConfMap["macaddr"] == nil || nicConfMap["macaddr"].(string) == "" {
			// Generate Mac based on VmID and NicID so it will be the same always.
			macaddr := make(net.HardwareAddr, 6)
			rand.Seed(time.Now().UnixNano())
			rand.Read(macaddr)
			macaddr[0] = (macaddr[0] | 2) & 0xfe // fix from github issue #18
			macAddrUppr := strings.ToUpper(fmt.Sprintf("%v", macaddr))
			// use model=mac format for older proxmox compatability
			macAddr := fmt.Sprintf("%v=%v", nicConfMap["model"], macAddrUppr)

			// Add Mac to source map so it will be returned. (useful for some use case like Terraform)
			nicConfMap["macaddr"] = macAddrUppr
			// and also add it to the parameters which will be sent to Proxmox API.
			nicConfParam = append(nicConfParam, macAddr)
		} else {
			macAddr := fmt.Sprintf("%v=%v", nicConfMap["model"], nicConfMap["macaddr"].(string))
			nicConfParam = append(nicConfParam, macAddr)
		}

		// Set bridge if not nat.
		if nicConfMap["bridge"].(string) != "nat" {
			bridge := fmt.Sprintf("bridge=%v", nicConfMap["bridge"])
			nicConfParam = append(nicConfParam, bridge)
		}

		// Keys that are not used as real/direct conf.
		ignoredKeys := []string{"id", "bridge", "macaddr", "model"}

		// Rest of config.
		nicConfParam = nicConfParam.createDeviceParam(nicConfMap, ignoredKeys)

		// Add nic to Qemu prams.
		params[qemuNicName] = strings.Join(nicConfParam, ",")
	}

	return nil
}

// Create parameters for each disk.
func (c ConfigQemu) CreateQemuDisksParams(
	vmID int,
	params map[string]interface{},
	cloned bool,
) error {

	// For backward compatibility.
	if len(c.QemuDisks) == 0 && len(c.Storage) > 0 {

		dType := c.StorageType
		if dType == "" {
			if c.HasCloudInit() {
				dType = "scsi"
			} else {
				dType = "virtio"
			}
		}
		deprecatedStyleMap := QemuDevice{
			"type":         dType,
			"storage":      c.Storage,
			"size":         c.DiskSize,
			"storage_type": "lvm",  // default old style
			"cache":        "none", // default old value
		}

		c.QemuDisks[0] = deprecatedStyleMap
	}

	// For new style with multi disk device.
	for diskID, diskConfMap := range c.QemuDisks {

		// skip the first disk for clones (may not always be right, but a template probably has at least 1 disk)
		if diskID == 0 && cloned {
			continue
		}
		diskConfParam := QemuDeviceParam{
			"media=disk",
		}

		// Device name.
		deviceType := diskConfMap["type"].(string)
		qemuDiskName := deviceType + strconv.Itoa(diskID)

		// Set disk storage.
		// Disk size.
		diskSizeGB := fmt.Sprintf("size=%v", diskConfMap["size"])
		diskConfParam = append(diskConfParam, diskSizeGB)

		// Disk name.
		var diskFile string
		// Currently ZFS local, LVM, and Directory are considered.
		// Other formats are not verified, but could be added if they're needed.
		rxStorageTypes := `(zfspool|lvm)`
		storageType := diskConfMap["storage_type"].(string)
		if matched, _ := regexp.MatchString(rxStorageTypes, storageType); matched {
			diskFile = fmt.Sprintf("file=%v:vm-%v-disk-%v", diskConfMap["storage"], vmID, diskID)
		} else {
			diskFile = fmt.Sprintf("file=%v:%v/vm-%v-disk-%v.%v", diskConfMap["storage"], vmID, vmID, diskID, diskConfMap["format"])
		}
		diskConfParam = append(diskConfParam, diskFile)

		// Set cache if not none (default).
		if diskConfMap["cache"].(string) != "none" {
			diskCache := fmt.Sprintf("cache=%v", diskConfMap["cache"])
			diskConfParam = append(diskConfParam, diskCache)
		}

		// Keys that are not used as real/direct conf.
		ignoredKeys := []string{"id", "type", "storage", "storage_type", "size", "cache"}

		// Rest of config.
		diskConfParam = diskConfParam.createDeviceParam(diskConfMap, ignoredKeys)

		// Add back to Qemu prams.
		params[qemuDiskName] = strings.Join(diskConfParam, ",")
	}

	return nil
}

// Create the parameters for each device that will be sent to Proxmox API.
func (p QemuDeviceParam) createDeviceParam(
	deviceConfMap QemuDevice,
	ignoredKeys []string,
) QemuDeviceParam {

	for key, value := range deviceConfMap {
		if ignored := inArray(ignoredKeys, key); !ignored {
			var confValue interface{}
			if bValue, ok := value.(bool); ok && bValue {
				confValue = "1"
			} else if sValue, ok := value.(string); ok && len(sValue) > 0 {
				confValue = sValue
			} else if iValue, ok := value.(int); ok && iValue > 0 {
				confValue = iValue
			}
			if confValue != nil {
				deviceConf := fmt.Sprintf("%v=%v", key, confValue)
				p = append(p, deviceConf)
			}
		}
	}

	return p
}

// readDeviceConfig - get standard sub-conf strings where `key=value` and update conf map.
func (confMap QemuDevice) readDeviceConfig(confList []string) error {
	// Add device config.
	for _, conf := range confList {
		key, value := ParseSubConf(conf, "=")
		confMap[key] = value
	}
	return nil
}

func (c ConfigQemu) String() string {
	jsConf, _ := json.Marshal(c)
	return string(jsConf)
}
