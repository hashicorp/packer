package proxmox

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

// LXC options for the Proxmox API
type configLxc struct {
	Ostemplate         string      `json:"ostemplate"`
	Arch               string      `json:"arch"`
	BWLimit            int         `json:"bwlimit,omitempty"`
	CMode              string      `json:"cmode"`
	Console            bool        `json:"console"`
	Cores              int         `json:"cores,omitempty"`
	CPULimit           int         `json:"cpulimit"`
	CPUUnits           int         `json:"cpuunits"`
	Description        string      `json:"description,omitempty"`
	Features           QemuDevice  `json:"features,omitempty"`
	Force              bool        `json:"force,omitempty"`
	Hookscript         string      `json:"hookscript,omitempty"`
	Hostname           string      `json:"hostname,omitempty"`
	IgnoreUnpackErrors bool        `json:"ignore-unpack-errors,omitempty"`
	Lock               string      `json:"lock,omitempty"`
	Memory             int         `json:"memory"`
	Mountpoints        QemuDevices `json:"mountpoints,omitempty"`
	Nameserver         string      `json:"nameserver,omitempty"`
	Networks           QemuDevices `json:"networks,omitempty"`
	OnBoot             bool        `json:"onboot"`
	OsType             string      `json:"ostype,omitempty"`
	Password           string      `json:"password,omitempty"`
	Pool               string      `json:"pool,omitempty"`
	Protection         bool        `json:"protection"`
	Restore            bool        `json:"restore,omitempty"`
	RootFs             string      `json:"rootfs,omitempty"`
	SearchDomain       string      `json:"searchdomain,omitempty"`
	SSHPublicKeys      string      `json:"ssh-public-keys,omitempty"`
	Start              bool        `json:"start"`
	Startup            string      `json:"startup,omitempty"`
	Storage            string      `json:"storage"`
	Swap               int         `json:"swap"`
	Template           bool        `json:"template,omitempty"`
	Tty                int         `json:"tty"`
	Unique             bool        `json:"unique,omitempty"`
	Unprivileged       bool        `json:"unprivileged"`
	Unused             []string    `json:"unused,omitempty"`
}

func NewConfigLxc() configLxc {
	return configLxc{
		Arch:         "amd64",
		CMode:        "tty",
		Console:      true,
		CPULimit:     0,
		CPUUnits:     1024,
		Memory:       512,
		OnBoot:       false,
		Protection:   false,
		Start:        false,
		Storage:      "local",
		Swap:         512,
		Template:     false,
		Tty:          2,
		Unprivileged: false,
	}
}

func NewConfigLxcFromJson(io io.Reader) (config configLxc, err error) {
	config = NewConfigLxc()
	err = json.NewDecoder(io).Decode(config)
	if err != nil {
		log.Fatal(err)
		return config, err
	}
	log.Println(config)
	return
}

func NewConfigLxcFromApi(vmr *VmRef, client *Client) (config *configLxc, err error) {
	// prepare json map to receive the information from the api
	var lxcConfig map[string]interface{}
	lxcConfig, err = client.GetVmConfig(vmr)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// prepare a new lxc config to store and return\
	// the information from api
	newConfig := NewConfigLxc()
	config = &newConfig

	arch := ""
	if _, isSet := lxcConfig["arch"]; isSet {
		arch = lxcConfig["arch"].(string)
	}
	cmode := ""
	if _, isSet := lxcConfig["cmode"]; isSet {
		cmode = lxcConfig["cmode"].(string)
	}
	console := true
	if _, isSet := lxcConfig["console"]; isSet {
		console = Itob(int(lxcConfig["console"].(float64)))
	}
	cores := 1
	if _, isSet := lxcConfig["cores"]; isSet {
		cores = int(lxcConfig["cores"].(float64))
	}
	cpulimit := 0
	if _, isSet := lxcConfig["cpulimit"]; isSet {
		cpulimit, _ = strconv.Atoi(lxcConfig["cpulimit"].(string))
	}
	cpuunits := 1024
	if _, isSet := lxcConfig["cpuunits"]; isSet {
		cpuunits = int(lxcConfig["cpuunits"].(float64))
	}
	description := ""
	if _, isSet := lxcConfig["description"]; isSet {
		description = lxcConfig["description"].(string)
	}

	// add features, if any
	if features, isSet := lxcConfig["features"]; isSet {
		featureList := strings.Split(features.(string), ",")

		// create new device map to store features
		featureMap := QemuDevice{}
		// add all features to device map
		featureMap.readDeviceConfig(featureList)
		// prepare empty feature map
		if config.Features == nil {
			config.Features = QemuDevice{}
		}
		// and device config to networks
		if len(featureMap) > 0 {
			config.Features = featureMap
		}
	}

	hookscript := ""
	if _, isSet := lxcConfig["hookscript"]; isSet {
		hookscript = lxcConfig["hookscript"].(string)
	}
	hostname := ""
	if _, isSet := lxcConfig["hostname"]; isSet {
		hostname = lxcConfig["hostname"].(string)
	}
	lock := ""
	if _, isSet := lxcConfig["lock"]; isSet {
		lock = lxcConfig["lock"].(string)
	}
	memory := 512
	if _, isSet := lxcConfig["memory"]; isSet {
		memory = int(lxcConfig["memory"].(float64))
	}

	// add mountpoints
	mpNames := []string{}

	for k, _ := range lxcConfig {
		if mpName := rxMpName.FindStringSubmatch(k); len(mpName) > 0 {
			mpNames = append(mpNames, mpName[0])
		}
	}

	for _, mpName := range mpNames {
		mpConfStr := lxcConfig[mpName]
		mpConfList := strings.Split(mpConfStr.(string), ",")

		id := rxDeviceID.FindStringSubmatch(mpName)
		mpID, _ := strconv.Atoi(id[0])
		// add mp id
		mpConfMap := QemuDevice{
			"id": mpID,
		}
		// add rest of device config
		mpConfMap.readDeviceConfig(mpConfList)
		// prepare empty mountpoint map
		if config.Mountpoints == nil {
			config.Mountpoints = QemuDevices{}
		}
		// and device config to mountpoints
		if len(mpConfMap) > 0 {
			config.Mountpoints[mpID] = mpConfMap
		}
	}

	nameserver := ""
	if _, isSet := lxcConfig["nameserver"]; isSet {
		nameserver = lxcConfig["nameserver"].(string)
	}

	// add networks
	nicNames := []string{}

	for k, _ := range lxcConfig {
		if nicName := rxNicName.FindStringSubmatch(k); len(nicName) > 0 {
			nicNames = append(nicNames, nicName[0])
		}
	}

	for _, nicName := range nicNames {
		nicConfStr := lxcConfig[nicName]
		nicConfList := strings.Split(nicConfStr.(string), ",")

		id := rxDeviceID.FindStringSubmatch(nicName)
		nicID, _ := strconv.Atoi(id[0])
		// add nic id
		nicConfMap := QemuDevice{
			"id": nicID,
		}
		// add rest of device config
		nicConfMap.readDeviceConfig(nicConfList)
		// prepare empty network map
		if config.Networks == nil {
			config.Networks = QemuDevices{}
		}
		// and device config to networks
		if len(nicConfMap) > 0 {
			config.Networks[nicID] = nicConfMap
		}
	}

	onboot := false
	if _, isSet := lxcConfig["onboot"]; isSet {
		onboot = Itob(int(lxcConfig["onboot"].(float64)))
	}
	ostype := ""
	if _, isSet := lxcConfig["ostype"]; isSet {
		ostype = lxcConfig["ostype"].(string)
	}
	protection := false
	if _, isSet := lxcConfig["protection"]; isSet {
		protection = Itob(int(lxcConfig["protection"].(float64)))
	}
	rootfs := ""
	if _, isSet := lxcConfig["rootfs"]; isSet {
		rootfs = lxcConfig["rootfs"].(string)
	}
	searchdomain := ""
	if _, isSet := lxcConfig["searchdomain"]; isSet {
		searchdomain = lxcConfig["searchdomain"].(string)
	}
	startup := ""
	if _, isSet := lxcConfig["startup"]; isSet {
		startup = lxcConfig["startup"].(string)
	}
	swap := 512
	if _, isSet := lxcConfig["swap"]; isSet {
		swap = int(lxcConfig["swap"].(float64))
	}
	template := false
	if _, isSet := lxcConfig["template"]; isSet {
		template = Itob(int(lxcConfig["template"].(float64)))
	}
	tty := 2
	if _, isSet := lxcConfig["tty"]; isSet {
		tty = int(lxcConfig["tty"].(float64))
	}
	unprivileged := false
	if _, isset := lxcConfig["unprivileged"]; isset {
		unprivileged = Itob(int(lxcConfig["unprivileged"].(float64)))
	}
	var unused []string
	if _, isset := lxcConfig["unused"]; isset {
		unused = lxcConfig["unused"].([]string)
	}

	config.Arch = arch
	config.CMode = cmode
	config.Console = console
	config.Cores = cores
	config.CPULimit = cpulimit
	config.CPUUnits = cpuunits
	config.Description = description
	config.OnBoot = onboot
	config.Hookscript = hookscript
	config.Hostname = hostname
	config.Lock = lock
	config.Memory = memory
	config.Nameserver = nameserver
	config.OnBoot = onboot
	config.OsType = ostype
	config.Protection = protection
	config.RootFs = rootfs
	config.SearchDomain = searchdomain
	config.Startup = startup
	config.Swap = swap
	config.Template = template
	config.Tty = tty
	config.Unprivileged = unprivileged
	config.Unused = unused

	return
}

// create LXC container using the Proxmox API
func (config configLxc) CreateLxc(vmr *VmRef, client *Client) (err error) {
	vmr.SetVmType("lxc")

	// convert config to map
	params, _ := json.Marshal(&config)
	var paramMap map[string]interface{}
	json.Unmarshal(params, &paramMap)

	// build list of features
	// add features as parameter list to lxc parameters
	// this overwrites the orginal formatting with a
	// comma separated list of "key=value" pairs
	featuresParam := QemuDeviceParam{}
	featuresParam = featuresParam.createDeviceParam(config.Features, nil)
	if len(featuresParam) > 0 {
		paramMap["features"] = strings.Join(featuresParam, ",")
	}

	// build list of mountpoints
	// this does the same as for the feature list
	// except that there can be multiple of these mountpoint sets
	// and each mountpoint set comes with a new id
	for mpID, mpConfMap := range config.Mountpoints {
		mpConfParam := QemuDeviceParam{}
		mpConfParam = mpConfParam.createDeviceParam(mpConfMap, nil)

		// add mp to lxc parameters
		mpName := fmt.Sprintf("mp%v", mpID)
		paramMap[mpName] = strings.Join(mpConfParam, ",")
	}

	// build list of network parameters
	for nicID, nicConfMap := range config.Networks {
		nicConfParam := QemuDeviceParam{}
		nicConfParam = nicConfParam.createDeviceParam(nicConfMap, nil)

		// add nic to lxc parameters
		nicName := fmt.Sprintf("net%v", nicID)
		paramMap[nicName] = strings.Join(nicConfParam, ",")
	}

	// build list of unused volumes for sake of completenes,
	// even if it is not recommended to change these volumes manually
	for volID, vol := range config.Unused {
		// add volume to lxc parameters
		volName := fmt.Sprintf("unused%v", volID)
		paramMap[volName] = vol
	}

	// now that we concatenated the key value parameter
	// list for the networks, mountpoints and unused volumes,
	// remove the original keys, since the Proxmox API does
	// not know how to handle this key
	delete(paramMap, "networks")
	delete(paramMap, "mountpoints")
	delete(paramMap, "unused")

	// amend vmid
	paramMap["vmid"] = vmr.vmId

	exitStatus, err := client.CreateLxcContainer(vmr.node, paramMap)
	if err != nil {
		return fmt.Errorf("Error creating LXC container: %v, error status: %s (params: %v)", err, exitStatus, params)
	}
	return
}

func (config configLxc) UpdateConfig(vmr *VmRef, client *Client) (err error) {
	// convert config to map
	params, _ := json.Marshal(&config)
	var paramMap map[string]interface{}
	json.Unmarshal(params, &paramMap)

	// build list of features
	// add features as parameter list to lxc parameters
	// this overwrites the orginal formatting with a
	// comma separated list of "key=value" pairs
	featuresParam := QemuDeviceParam{}
	featuresParam = featuresParam.createDeviceParam(config.Features, nil)
	paramMap["features"] = strings.Join(featuresParam, ",")

	// build list of mountpoints
	// this does the same as for the feature list
	// except that there can be multiple of these mountpoint sets
	// and each mountpoint set comes with a new id
	for mpID, mpConfMap := range config.Mountpoints {
		mpConfParam := QemuDeviceParam{}
		mpConfParam = mpConfParam.createDeviceParam(mpConfMap, nil)

		// add mp to lxc parameters
		mpName := fmt.Sprintf("mp%v", mpID)
		paramMap[mpName] = strings.Join(mpConfParam, ",")
	}

	// build list of network parameters
	for nicID, nicConfMap := range config.Networks {
		nicConfParam := QemuDeviceParam{}
		nicConfParam = nicConfParam.createDeviceParam(nicConfMap, nil)

		// add nic to lxc parameters
		nicName := fmt.Sprintf("net%v", nicID)
		paramMap[nicName] = strings.Join(nicConfParam, ",")
	}

	// build list of unused volumes for sake of completenes,
	// even if it is not recommended to change these volumes manually
	for volID, vol := range config.Unused {
		// add volume to lxc parameters
		volName := fmt.Sprintf("unused%v", volID)
		paramMap[volName] = vol
	}

	// now that we concatenated the key value parameter
	// list for the networks, mountpoints and unused volumes,
	// remove the original keys, since the Proxmox API does
	// not know how to handle this key
	delete(paramMap, "networks")
	delete(paramMap, "mountpoints")
	delete(paramMap, "unused")

	// delete parameters wich are not supported in updated operations
	delete(paramMap, "pool")
	delete(paramMap, "storage")
	delete(paramMap, "password")
	delete(paramMap, "ostemplate")
	delete(paramMap, "start")
	// even though it is listed as a PUT option in the API documentation
	// we remove it here because "it should not be modified manually";
	// also, error "500 unable to modify read-only option: 'unprivileged'"
	delete(paramMap, "unprivileged")

	_, err = client.SetLxcConfig(vmr, paramMap)
	return err
}
