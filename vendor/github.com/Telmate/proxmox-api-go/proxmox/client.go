package proxmox

// inspired by https://github.com/Telmate/vagrant-proxmox/blob/master/lib/vagrant-proxmox/proxmox/connection.rb

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TaskStatusCheckInterval - time between async checks in seconds
const TaskStatusCheckInterval = 2

const exitStatusSuccess = "OK"

// Client - URL, user and password to specifc Proxmox node
type Client struct {
	session     *Session
	ApiUrl      string
	Username    string
	Password    string
	Otp         string
	TaskTimeout int
}

// VmRef - virtual machine ref parts
// map[type:qemu node:proxmox1-xx id:qemu/132 diskread:5.57424738e+08 disk:0 netin:5.9297450593e+10 mem:3.3235968e+09 uptime:1.4567097e+07 vmid:132 template:0 maxcpu:2 netout:6.053310416e+09 maxdisk:3.4359738368e+10 maxmem:8.592031744e+09 diskwrite:1.49663619584e+12 status:running cpu:0.00386980694947209 name:appt-app1-dev.xxx.xx]
type VmRef struct {
	vmId    int
	node    string
	pool    string
	vmType  string
	haState string
}

func (vmr *VmRef) SetNode(node string) {
	vmr.node = node
	return
}

func (vmr *VmRef) SetPool(pool string) {
	vmr.pool = pool
}

func (vmr *VmRef) SetVmType(vmType string) {
	vmr.vmType = vmType
	return
}

func (vmr *VmRef) GetVmType() string {
	return vmr.vmType
}

func (vmr *VmRef) VmId() int {
	return vmr.vmId
}

func (vmr *VmRef) Node() string {
	return vmr.node
}

func (vmr *VmRef) Pool() string {
	return vmr.pool
}

func (vmr *VmRef) HaState() string {
	return vmr.haState
}

func NewVmRef(vmId int) (vmr *VmRef) {
	vmr = &VmRef{vmId: vmId, node: "", vmType: ""}
	return
}

func NewClient(apiUrl string, hclient *http.Client, tls *tls.Config, taskTimeout int) (client *Client, err error) {
	var sess *Session
	sess, err = NewSession(apiUrl, hclient, tls)
	if err == nil {
		client = &Client{session: sess, ApiUrl: apiUrl, TaskTimeout: taskTimeout}
	}
	return client, err
}

// SetAPIToken specifies a pair of user identifier and token UUID to use
// for authenticating API calls.
// If this is set, a ticket from calling `Login` will not be used.
//
// - `userID` is expected to be in the form `USER@REALM!TOKENID`
// - `token` is just the UUID you get when initially creating the token
//
// See https://pve.proxmox.com/wiki/User_Management#pveum_tokens
func (c *Client) SetAPIToken(userID, token string) {
	c.session.SetAPIToken(userID, token)
}

func (c *Client) Login(username string, password string, otp string) (err error) {
	c.Username = username
	c.Password = password
	c.Otp = otp
	return c.session.Login(username, password, otp)
}

func (c *Client) GetJsonRetryable(url string, data *map[string]interface{}, tries int) error {
	var statErr error
	for ii := 0; ii < tries; ii++ {
		_, statErr = c.session.GetJSON(url, nil, nil, data)
		if statErr == nil {
			return nil
		}
		// if statErr != io.ErrUnexpectedEOF { // don't give up on ErrUnexpectedEOF
		//   return statErr
		// }
		time.Sleep(5 * time.Second)
	}
	return statErr
}

func (c *Client) GetNodeList() (list map[string]interface{}, err error) {
	err = c.GetJsonRetryable("/nodes", &list, 3)
	return
}

func (c *Client) GetVmList() (list map[string]interface{}, err error) {
	err = c.GetJsonRetryable("/cluster/resources?type=vm", &list, 3)
	return
}

func (c *Client) CheckVmRef(vmr *VmRef) (err error) {
	if vmr.node == "" || vmr.vmType == "" {
		_, err = c.GetVmInfo(vmr)
	}
	return
}

func (c *Client) GetVmInfo(vmr *VmRef) (vmInfo map[string]interface{}, err error) {
	resp, err := c.GetVmList()
	vms := resp["data"].([]interface{})
	for vmii := range vms {
		vm := vms[vmii].(map[string]interface{})
		if int(vm["vmid"].(float64)) == vmr.vmId {
			vmInfo = vm
			vmr.node = vmInfo["node"].(string)
			vmr.vmType = vmInfo["type"].(string)
			vmr.pool = ""
			if vmInfo["pool"] != nil {
				vmr.pool = vmInfo["pool"].(string)
			}
			if vmInfo["hastate"] != nil {
				vmr.haState = vmInfo["hastate"].(string)
			}
			return
		}
	}
	return nil, errors.New(fmt.Sprintf("Vm '%d' not found", vmr.vmId))
}

func (c *Client) GetVmRefByName(vmName string) (vmr *VmRef, err error) {
	resp, err := c.GetVmList()
	vms := resp["data"].([]interface{})
	for vmii := range vms {
		vm := vms[vmii].(map[string]interface{})
		if vm["name"] != nil && vm["name"].(string) == vmName {
			vmr = NewVmRef(int(vm["vmid"].(float64)))
			vmr.node = vm["node"].(string)
			vmr.vmType = vm["type"].(string)
			vmr.pool = ""
			if vm["pool"] != nil {
				vmr.pool = vm["pool"].(string)
			}
			if vm["hastate"] != nil {
				vmr.haState = vm["hastate"].(string)
			}
			return
		}
	}
	return nil, errors.New(fmt.Sprintf("Vm '%s' not found", vmName))
}

func (c *Client) GetVmState(vmr *VmRef) (vmState map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	url := fmt.Sprintf("/nodes/%s/%s/%d/status/current", vmr.node, vmr.vmType, vmr.vmId)
	err = c.GetJsonRetryable(url, &data, 3)
	if err != nil {
		return nil, err
	}
	if data["data"] == nil {
		return nil, errors.New("Vm STATE not readable")
	}
	vmState = data["data"].(map[string]interface{})
	return
}

func (c *Client) GetVmConfig(vmr *VmRef) (vmConfig map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	url := fmt.Sprintf("/nodes/%s/%s/%d/config", vmr.node, vmr.vmType, vmr.vmId)
	err = c.GetJsonRetryable(url, &data, 3)
	if err != nil {
		return nil, err
	}
	if data["data"] == nil {
		return nil, errors.New("Vm CONFIG not readable")
	}
	vmConfig = data["data"].(map[string]interface{})
	return
}

func (c *Client) GetStorageStatus(vmr *VmRef, storageName string) (storageStatus map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	url := fmt.Sprintf("/nodes/%s/storage/%s/status", vmr.node, storageName)
	err = c.GetJsonRetryable(url, &data, 3)
	if err != nil {
		return nil, err
	}
	if data["data"] == nil {
		return nil, errors.New("Storage STATUS not readable")
	}
	storageStatus = data["data"].(map[string]interface{})
	return
}

func (c *Client) GetStorageContent(vmr *VmRef, storageName string) (data map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("/nodes/%s/storage/%s/content", vmr.node, storageName)
	err = c.GetJsonRetryable(url, &data, 3)
	if err != nil {
		return nil, err
	}
	if data["data"] == nil {
		return nil, errors.New("Storage Content not readable")
	}
	return
}

func (c *Client) GetVmSpiceProxy(vmr *VmRef) (vmSpiceProxy map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	url := fmt.Sprintf("/nodes/%s/%s/%d/spiceproxy", vmr.node, vmr.vmType, vmr.vmId)
	_, err = c.session.PostJSON(url, nil, nil, nil, &data)
	if err != nil {
		return nil, err
	}
	if data["data"] == nil {
		return nil, errors.New("Vm SpiceProxy not readable")
	}
	vmSpiceProxy = data["data"].(map[string]interface{})
	return
}

type AgentNetworkInterface struct {
	MACAddress  string
	IPAddresses []net.IP
	Name        string
	Statistics  map[string]int64
}

func (a *AgentNetworkInterface) UnmarshalJSON(b []byte) error {
	var intermediate struct {
		HardwareAddress string `json:"hardware-address"`
		IPAddresses     []struct {
			IPAddress     string `json:"ip-address"`
			IPAddressType string `json:"ip-address-type"`
			Prefix        int    `json:"prefix"`
		} `json:"ip-addresses"`
		Name       string           `json:"name"`
		Statistics map[string]int64 `json:"statistics"`
	}
	err := json.Unmarshal(b, &intermediate)
	if err != nil {
		return err
	}

	a.IPAddresses = make([]net.IP, len(intermediate.IPAddresses))
	for idx, ip := range intermediate.IPAddresses {
		a.IPAddresses[idx] = net.ParseIP(ip.IPAddress)
		if a.IPAddresses[idx] == nil {
			return fmt.Errorf("Could not parse %s as IP", ip.IPAddress)
		}
	}
	a.MACAddress = intermediate.HardwareAddress
	a.Name = intermediate.Name
	a.Statistics = intermediate.Statistics
	return nil
}

func (c *Client) GetVmAgentNetworkInterfaces(vmr *VmRef) ([]AgentNetworkInterface, error) {
	var ifs []AgentNetworkInterface
	err := c.doAgentGet(vmr, "network-get-interfaces", &ifs)
	return ifs, err
}

func (c *Client) doAgentGet(vmr *VmRef, command string, output interface{}) error {
	err := c.CheckVmRef(vmr)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("/nodes/%s/%s/%d/agent/%s", vmr.node, vmr.vmType, vmr.vmId, command)
	resp, err := c.session.Get(url, nil, nil)
	if err != nil {
		return err
	}

	return TypedResponse(resp, output)
}

func (c *Client) CreateTemplate(vmr *VmRef) error {
	err := c.CheckVmRef(vmr)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("/nodes/%s/%s/%d/template", vmr.node, vmr.vmType, vmr.vmId)
	_, err = c.session.Post(url, nil, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) MonitorCmd(vmr *VmRef, command string) (monitorRes map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(map[string]interface{}{"command": command})
	url := fmt.Sprintf("/nodes/%s/%s/%d/monitor", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err != nil {
		return nil, err
	}
	monitorRes, err = ResponseJSON(resp)
	return
}

func (c *Client) Sendkey(vmr *VmRef, qmKey string) error {
	err := c.CheckVmRef(vmr)
	if err != nil {
		return err
	}
	reqbody := ParamsToBody(map[string]interface{}{"key": qmKey})
	url := fmt.Sprintf("/nodes/%s/%s/%d/sendkey", vmr.node, vmr.vmType, vmr.vmId)
	// No return, even for errors: https://bugzilla.proxmox.com/show_bug.cgi?id=2275
	_, err = c.session.Put(url, nil, nil, &reqbody)

	return err
}

// WaitForCompletion - poll the API for task completion
func (c *Client) WaitForCompletion(taskResponse map[string]interface{}) (waitExitStatus string, err error) {
	if taskResponse["errors"] != nil {
		errJSON, _ := json.MarshalIndent(taskResponse["errors"], "", "  ")
		return string(errJSON), errors.New("Error reponse")
	}
	if taskResponse["data"] == nil {
		return "", nil
	}
	waited := 0
	taskUpid := taskResponse["data"].(string)
	for waited < c.TaskTimeout {
		exitStatus, statErr := c.GetTaskExitstatus(taskUpid)
		if statErr != nil {
			if statErr != io.ErrUnexpectedEOF { // don't give up on ErrUnexpectedEOF
				return "", statErr
			}
		}
		if exitStatus != nil {
			waitExitStatus = exitStatus.(string)
			return
		}
		time.Sleep(TaskStatusCheckInterval * time.Second)
		waited = waited + TaskStatusCheckInterval
	}
	return "", errors.New("Wait timeout for:" + taskUpid)
}

var rxTaskNode = regexp.MustCompile("UPID:(.*?):")

func (c *Client) GetTaskExitstatus(taskUpid string) (exitStatus interface{}, err error) {
	node := rxTaskNode.FindStringSubmatch(taskUpid)[1]
	url := fmt.Sprintf("/nodes/%s/tasks/%s/status", node, taskUpid)
	var data map[string]interface{}
	_, err = c.session.GetJSON(url, nil, nil, &data)
	if err == nil {
		exitStatus = data["data"].(map[string]interface{})["exitstatus"]
	}
	if exitStatus != nil && exitStatus != exitStatusSuccess {
		err = errors.New(exitStatus.(string))
	}
	return
}

func (c *Client) StatusChangeVm(vmr *VmRef, setStatus string) (exitStatus string, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("/nodes/%s/%s/%d/status/%s", vmr.node, vmr.vmType, vmr.vmId, setStatus)
	var taskResponse map[string]interface{}
	for i := 0; i < 3; i++ {
		_, err = c.session.PostJSON(url, nil, nil, nil, &taskResponse)
		exitStatus, err = c.WaitForCompletion(taskResponse)
		if exitStatus == "" {
			time.Sleep(TaskStatusCheckInterval * time.Second)
		} else {
			return
		}
	}
	return
}

func (c *Client) StartVm(vmr *VmRef) (exitStatus string, err error) {
	return c.StatusChangeVm(vmr, "start")
}

func (c *Client) StopVm(vmr *VmRef) (exitStatus string, err error) {
	return c.StatusChangeVm(vmr, "stop")
}

func (c *Client) ShutdownVm(vmr *VmRef) (exitStatus string, err error) {
	return c.StatusChangeVm(vmr, "shutdown")
}

func (c *Client) ResetVm(vmr *VmRef) (exitStatus string, err error) {
	return c.StatusChangeVm(vmr, "reset")
}

func (c *Client) SuspendVm(vmr *VmRef) (exitStatus string, err error) {
	return c.StatusChangeVm(vmr, "suspend")
}

func (c *Client) ResumeVm(vmr *VmRef) (exitStatus string, err error) {
	return c.StatusChangeVm(vmr, "resume")
}

func (c *Client) DeleteVm(vmr *VmRef) (exitStatus string, err error) {
	return c.DeleteVmParams(vmr, nil)
}

func (c *Client) DeleteVmParams(vmr *VmRef, params map[string]interface{}) (exitStatus string, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return "", err
	}

	//Remove HA if required
	if vmr.haState != "" {
		url := fmt.Sprintf("/cluster/ha/resources/%d", vmr.vmId)
		resp, err := c.session.Delete(url, nil, nil)
		if err == nil {
			taskResponse, err := ResponseJSON(resp)
			if err != nil {
				return "", err
			}
			exitStatus, err = c.WaitForCompletion(taskResponse)
			if err != nil {
				return "", err
			}
		}
	}

	reqbody := ParamsToBody(params)
	url := fmt.Sprintf("/nodes/%s/%s/%d", vmr.node, vmr.vmType, vmr.vmId)
	var taskResponse map[string]interface{}
	_, err = c.session.RequestJSON("DELETE", url, nil, nil, &reqbody, &taskResponse)
	exitStatus, err = c.WaitForCompletion(taskResponse)
	return
}

func (c *Client) CreateQemuVm(node string, vmParams map[string]interface{}) (exitStatus string, err error) {
	// Create VM disks first to ensure disks names.
	createdDisks, createdDisksErr := c.createVMDisks(node, vmParams)
	if createdDisksErr != nil {
		return "", createdDisksErr
	}

	// Then create the VM itself.
	reqbody := ParamsToBody(vmParams)
	url := fmt.Sprintf("/nodes/%s/qemu", node)
	var resp *http.Response
	resp, err = c.session.Post(url, nil, nil, &reqbody)
	defer resp.Body.Close()
	if err != nil {
		// This might not work if we never got a body. We'll ignore errors in trying to read,
		// but extract the body if possible to give any error information back in the exitStatus
		b, _ := ioutil.ReadAll(resp.Body)
		exitStatus = string(b)
		return exitStatus, err
	}

	taskResponse, err := ResponseJSON(resp)
	if err != nil {
		return "", err
	}
	exitStatus, err = c.WaitForCompletion(taskResponse)
	// Delete VM disks if the VM didn't create.
	if exitStatus != "OK" {
		deleteDisksErr := c.DeleteVMDisks(node, createdDisks)
		if deleteDisksErr != nil {
			return "", deleteDisksErr
		}
	}

	return
}

func (c *Client) CreateLxcContainer(node string, vmParams map[string]interface{}) (exitStatus string, err error) {
	reqbody := ParamsToBody(vmParams)
	url := fmt.Sprintf("/nodes/%s/lxc", node)
	var resp *http.Response
	resp, err = c.session.Post(url, nil, nil, &reqbody)
	defer resp.Body.Close()
	if err != nil {
		// This might not work if we never got a body. We'll ignore errors in trying to read,
		// but extract the body if possible to give any error information back in the exitStatus
		b, _ := ioutil.ReadAll(resp.Body)
		exitStatus = string(b)
		return exitStatus, err
	}

	taskResponse, err := ResponseJSON(resp)
	if err != nil {
		return "", err
	}
	exitStatus, err = c.WaitForCompletion(taskResponse)

	return
}

func (c *Client) CloneQemuVm(vmr *VmRef, vmParams map[string]interface{}) (exitStatus string, err error) {
	reqbody := ParamsToBody(vmParams)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/clone", vmr.node, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return "", err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
	}
	return
}

func (c *Client) CreateQemuSnapshot(vmr *VmRef, snapshotName string) (exitStatus string, err error) {
	err = c.CheckVmRef(vmr)
	snapshotParams := map[string]interface{}{
		"snapname": snapshotName,
	}
	reqbody := ParamsToBody(snapshotParams)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("/nodes/%s/%s/%d/snapshot/", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return "", err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
	}
	return
}

func (c *Client) DeleteQemuSnapshot(vmr *VmRef, snapshotName string) (exitStatus string, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("/nodes/%s/%s/%d/snapshot/%s", vmr.node, vmr.vmType, vmr.vmId, snapshotName)
	resp, err := c.session.Delete(url, nil, nil)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return "", err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
	}
	return
}

func (c *Client) ListQemuSnapshot(vmr *VmRef) (taskResponse map[string]interface{}, exitStatus string, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, "", err
	}
	url := fmt.Sprintf("/nodes/%s/%s/%d/snapshot/", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Get(url, nil, nil)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, "", err
		}
		return taskResponse, "", nil
	}
	return
}

func (c *Client) RollbackQemuVm(vmr *VmRef, snapshot string) (exitStatus string, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("/nodes/%s/%s/%d/snapshot/%s/rollback", vmr.node, vmr.vmType, vmr.vmId, snapshot)
	var taskResponse map[string]interface{}
	_, err = c.session.PostJSON(url, nil, nil, nil, &taskResponse)
	exitStatus, err = c.WaitForCompletion(taskResponse)
	return
}

// SetVmConfig - send config options
func (c *Client) SetVmConfig(vmr *VmRef, vmParams map[string]interface{}) (exitStatus interface{}, err error) {
	reqbody := ParamsToBody(vmParams)
	url := fmt.Sprintf("/nodes/%s/%s/%d/config", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
	}
	return
}

// SetLxcConfig - send config options
func (c *Client) SetLxcConfig(vmr *VmRef, vmParams map[string]interface{}) (exitStatus interface{}, err error) {
	reqbody := ParamsToBody(vmParams)
	url := fmt.Sprintf("/nodes/%s/%s/%d/config", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Put(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
	}
	return
}

// MigrateNode - Migrate a VM
func (c *Client) MigrateNode(vmr *VmRef, newTargetNode string, online bool) (exitStatus interface{}, err error) {
	reqbody := ParamsToBody(map[string]interface{}{"target": newTargetNode, "online": online})
	url := fmt.Sprintf("/nodes/%s/%s/%d/migrate", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
		return exitStatus, err
	}
	return nil, err
}

// ResizeQemuDisk allows the caller to increase the size of a disk by the indicated number of gigabytes
func (c *Client) ResizeQemuDisk(vmr *VmRef, disk string, moreSizeGB int) (exitStatus interface{}, err error) {
	size := fmt.Sprintf("+%dG", moreSizeGB)
	return c.ResizeQemuDiskRaw(vmr, disk, size)
}

// ResizeQemuDiskRaw allows the caller to provide the raw resize string to be send to proxmox.
// See the proxmox API documentation for full information, but the short version is if you prefix
// your desired size with a '+' character it will ADD size to the disk.  If you just specify the size by
// itself it will do an absolute resizing to the specified size. Permitted suffixes are K, M, G, T
// to indicate order of magnitude (kilobyte, megabyte, etc). Decrease of disk size is not permitted.
func (c *Client) ResizeQemuDiskRaw(vmr *VmRef, disk string, size string) (exitStatus interface{}, err error) {
	// PUT
	//disk:virtio0
	//size:+2G
	if disk == "" {
		disk = "virtio0"
	}
	reqbody := ParamsToBody(map[string]interface{}{"disk": disk, "size": size})
	url := fmt.Sprintf("/nodes/%s/%s/%d/resize", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Put(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
	}
	return
}

func (c *Client) MoveLxcDisk(vmr *VmRef, disk string, storage string) (exitStatus interface{}, err error) {
	reqbody := ParamsToBody(map[string]interface{}{"disk": disk, "storage": storage, "delete": true})
	url := fmt.Sprintf("/nodes/%s/%s/%d/move_volume", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
	}
	return
}

func (c *Client) MoveQemuDisk(vmr *VmRef, disk string, storage string) (exitStatus interface{}, err error) {
	if disk == "" {
		disk = "virtio0"
	}
	reqbody := ParamsToBody(map[string]interface{}{"disk": disk, "storage": storage, "delete": true})
	url := fmt.Sprintf("/nodes/%s/%s/%d/move_disk", vmr.node, vmr.vmType, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
	}
	return
}

// GetNextID - Get next free VMID
func (c *Client) GetNextID(currentID int) (nextID int, err error) {
	var data map[string]interface{}
	var url string
	if currentID >= 100 {
		url = fmt.Sprintf("/cluster/nextid?vmid=%d", currentID)
	} else {
		url = "/cluster/nextid"
	}
	_, err = c.session.GetJSON(url, nil, nil, &data)
	if err == nil {
		if data["errors"] != nil {
			if currentID >= 100 {
				return c.GetNextID(currentID + 1)
			} else {
				return -1, errors.New("error using /cluster/nextid")
			}
		}
		nextID, err = strconv.Atoi(data["data"].(string))
	} else if strings.HasPrefix(err.Error(), "400 ") {
		return c.GetNextID(currentID + 1)
	}
	return
}

// CreateVMDisk - Create single disk for VM on host node.
func (c *Client) CreateVMDisk(
	nodeName string,
	storageName string,
	fullDiskName string,
	diskParams map[string]interface{},
) error {

	reqbody := ParamsToBody(diskParams)
	url := fmt.Sprintf("/nodes/%s/storage/%s/content", nodeName, storageName)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return err
		}
		if diskName, containsData := taskResponse["data"]; !containsData || diskName != fullDiskName {
			return errors.New(fmt.Sprintf("Cannot create VM disk %s - %s", fullDiskName, diskName))
		}
	} else {
		return err
	}

	return nil
}

// createVMDisks - Make disks parameters and create all VM disks on host node.
func (c *Client) createVMDisks(
	node string,
	vmParams map[string]interface{},
) (disks []string, err error) {
	var createdDisks []string
	vmID := vmParams["vmid"].(int)
	for deviceName, deviceConf := range vmParams {
		rxStorageModels := `(ide|sata|scsi|virtio)\d+`
		if matched, _ := regexp.MatchString(rxStorageModels, deviceName); matched {
			deviceConfMap := ParsePMConf(deviceConf.(string), "")
			// This if condition to differentiate between `disk` and `cdrom`.
			if media, containsFile := deviceConfMap["media"]; containsFile && media == "disk" {
				fullDiskName := deviceConfMap["file"].(string)
				storageName, volumeName := getStorageAndVolumeName(fullDiskName, ":")
				diskParams := map[string]interface{}{
					"vmid":     vmID,
					"filename": volumeName,
					"size":     deviceConfMap["size"],
				}
				err := c.CreateVMDisk(node, storageName, fullDiskName, diskParams)
				if err != nil {
					return createdDisks, err
				} else {
					createdDisks = append(createdDisks, fullDiskName)
				}
			}
		}
	}

	return createdDisks, nil
}

// DeleteVMDisks - Delete VM disks from host node.
// By default the VM disks are deteled when the VM is deleted,
// so mainly this is used to delete the disks in case VM creation didn't complete.
func (c *Client) DeleteVMDisks(
	node string,
	disks []string,
) error {
	for _, fullDiskName := range disks {
		storageName, volumeName := getStorageAndVolumeName(fullDiskName, ":")
		url := fmt.Sprintf("/nodes/%s/storage/%s/content/%s", node, storageName, volumeName)
		_, err := c.session.Post(url, nil, nil, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// VzDump - Create backup
func (c *Client) VzDump(vmr *VmRef, params map[string]interface{}) (exitStatus interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(params)
	url := fmt.Sprintf("/nodes/%s/vzdump", vmr.node)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
	}
	return
}

// CreateVNCProxy - Creates a TCP VNC proxy connections
func (c *Client) CreateVNCProxy(vmr *VmRef, params map[string]interface{}) (vncProxyRes map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(params)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/vncproxy", vmr.node, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err != nil {
		return nil, err
	}
	vncProxyRes, err = ResponseJSON(resp)
	if err != nil {
		return nil, err
	}
	if vncProxyRes["data"] == nil {
		return nil, errors.New("VNC Proxy not readable")
	}
	vncProxyRes = vncProxyRes["data"].(map[string]interface{})
	return
}

// GetExecStatus - Gets the status of the given pid started by the guest-agent
func (c *Client) GetExecStatus(vmr *VmRef, pid string) (status map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	err = c.GetJsonRetryable(fmt.Sprintf("/nodes/%s/%s/%d/agent/exec-status?pid=%s", vmr.node, vmr.vmType, vmr.vmId, pid), &status, 3)
	if err == nil {
		status = status["data"].(map[string]interface{})
	}
	return
}

// SetQemuFirewallOptions - Set Firewall options.
func (c *Client) SetQemuFirewallOptions(vmr *VmRef, fwOptions map[string]interface{}) (exitStatus interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(fwOptions)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/options", vmr.node, vmr.vmId)
	resp, err := c.session.Put(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
	}
	return
}

// GetQemuFirewallOptions - Get VM firewall options.
func (c *Client) GetQemuFirewallOptions(vmr *VmRef) (firewallOptions map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/options", vmr.node, vmr.vmId)
	resp, err := c.session.Get(url, nil, nil)
	if err == nil {
		firewallOptions, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		return firewallOptions, nil
	}
	return
}

// CreateQemuIPSet - Create new IPSet
func (c *Client) CreateQemuIPSet(vmr *VmRef, params map[string]interface{}) (exitStatus interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	reqbody := ParamsToBody(params)
	url := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/ipset", vmr.node, vmr.vmId)
	resp, err := c.session.Post(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
	}
	return
}

// GetQemuIPSet - List IPSets
func (c *Client) GetQemuIPSet(vmr *VmRef) (ipsets map[string]interface{}, err error) {
	err = c.CheckVmRef(vmr)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("/nodes/%s/qemu/%d/firewall/ipset", vmr.node, vmr.vmId)
	resp, err := c.session.Get(url, nil, nil)
	if err == nil {
		ipsets, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		return ipsets, nil
	}
	return
}

func (c *Client) Upload(node string, storage string, contentType string, filename string, file io.Reader) error {
	var doStreamingIO bool
	var fileSize int64
	var contentLength int64

	if f, ok := file.(*os.File); ok {
		doStreamingIO = true
		fileInfo, err := f.Stat()
		if err != nil {
			return err
		}
		fileSize = fileInfo.Size()
	}

	var body io.Reader
	var mimetype string
	var err error

	if doStreamingIO {
		body, mimetype, contentLength, err = createStreamedUploadBody(contentType, filename, fileSize, file)
	} else {
		body, mimetype, err = createUploadBody(contentType, filename, file)
	}
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/nodes/%s/storage/%s/upload", c.session.ApiUrl, node, storage)
	req, err := c.session.NewRequest(http.MethodPost, url, nil, body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", mimetype)
	req.Header.Add("Accept", "application/json")

	if doStreamingIO {
		req.ContentLength = contentLength
	}

	resp, err := c.session.Do(req)
	if err != nil {
		return err
	}

	taskResponse, err := ResponseJSON(resp)
	if err != nil {
		return err
	}
	exitStatus, err := c.WaitForCompletion(taskResponse)
	if err != nil {
		return err
	}
	if exitStatus != exitStatusSuccess {
		return fmt.Errorf("Moving file to destination failed: %v", exitStatus)
	}
	return nil
}

func createUploadBody(contentType string, filename string, r io.Reader) (io.Reader, string, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	err := w.WriteField("content", contentType)
	if err != nil {
		return nil, "", err
	}

	fw, err := w.CreateFormFile("filename", filename)
	if err != nil {
		return nil, "", err
	}
	_, err = io.Copy(fw, r)
	if err != nil {
		return nil, "", err
	}

	err = w.Close()
	if err != nil {
		return nil, "", err
	}

	return &buf, w.FormDataContentType(), nil
}

// createStreamedUploadBody - Use MultiReader to create the multipart body from the file reader,
// avoiding allocation of large files in memory before upload (useful e.g. for Windows ISOs).
func createStreamedUploadBody(contentType string, filename string, fileSize int64, r io.Reader) (io.Reader, string, int64, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	err := w.WriteField("content", contentType)
	if err != nil {
		return nil, "", 0, err
	}

	_, err = w.CreateFormFile("filename", filename)
	if err != nil {
		return nil, "", 0, err
	}

	headerSize := buf.Len()

	err = w.Close()
	if err != nil {
		return nil, "", 0, err
	}

	mr := io.MultiReader(bytes.NewReader(buf.Bytes()[:headerSize]),
		r,
		bytes.NewReader(buf.Bytes()[headerSize:]))

	contentLength := int64(buf.Len()) + fileSize

	return mr, w.FormDataContentType(), contentLength, nil
}

// getStorageAndVolumeName - Extract disk storage and disk volume, since disk name is saved
// in Proxmox with its storage.
func getStorageAndVolumeName(
	fullDiskName string,
	separator string,
) (storageName string, diskName string) {
	storageAndVolumeName := strings.Split(fullDiskName, separator)
	storageName, volumeName := storageAndVolumeName[0], storageAndVolumeName[1]

	// when disk type is dir, volumeName is `file=local:100/vm-100-disk-0.raw`
	re := regexp.MustCompile(`\d+/(?P<filename>\S+.\S+)`)
	match := re.FindStringSubmatch(volumeName)
	if len(match) == 2 {
		volumeName = match[1]
	}

	return storageName, volumeName
}

func (c *Client) UpdateVMPool(vmr *VmRef, pool string) (exitStatus interface{}, err error) {
	// Same pool
	if vmr.pool == pool {
		return
	}

	// Remove from old pool
	if vmr.pool != "" {
		paramMap := map[string]interface{}{
			"vms":    vmr.vmId,
			"delete": 1,
		}
		reqbody := ParamsToBody(paramMap)
		url := fmt.Sprintf("/pools/%s", vmr.pool)
		resp, err := c.session.Put(url, nil, nil, &reqbody)
		if err == nil {
			taskResponse, err := ResponseJSON(resp)
			if err != nil {
				return nil, err
			}
			exitStatus, err = c.WaitForCompletion(taskResponse)

			if err != nil {
				return nil, err
			}
		}
	}
	// Add to the new pool
	if pool != "" {
		paramMap := map[string]interface{}{
			"vms": vmr.vmId,
		}
		reqbody := ParamsToBody(paramMap)
		url := fmt.Sprintf("/pools/%s", pool)
		resp, err := c.session.Put(url, nil, nil, &reqbody)
		if err == nil {
			taskResponse, err := ResponseJSON(resp)
			if err != nil {
				return nil, err
			}
			exitStatus, err = c.WaitForCompletion(taskResponse)
		} else {
			return nil, err
		}
	}
	return
}

func (c *Client) UpdateVMHA(vmr *VmRef, haState string) (exitStatus interface{}, err error) {
	// Same hastate
	if vmr.haState == haState {
		return
	}

	//Remove HA
	if haState == "" {
		url := fmt.Sprintf("/cluster/ha/resources/%d", vmr.vmId)
		resp, err := c.session.Delete(url, nil, nil)
		if err == nil {
			taskResponse, err := ResponseJSON(resp)
			if err != nil {
				return nil, err
			}
			exitStatus, err = c.WaitForCompletion(taskResponse)
		}
		return nil, err
	}

	//Activate HA
	if vmr.haState == "" {
		paramMap := map[string]interface{}{
			"sid": vmr.vmId,
		}
		reqbody := ParamsToBody(paramMap)
		resp, err := c.session.Post("/cluster/ha/resources", nil, nil, &reqbody)
		if err == nil {
			taskResponse, err := ResponseJSON(resp)
			if err != nil {
				return nil, err
			}
			exitStatus, err = c.WaitForCompletion(taskResponse)

			if err != nil {
				return nil, err
			}
		}
	}

	//Set wanted state
	paramMap := map[string]interface{}{
		"state": haState,
	}
	reqbody := ParamsToBody(paramMap)
	url := fmt.Sprintf("/cluster/ha/resources/%d", vmr.vmId)
	resp, err := c.session.Put(url, nil, nil, &reqbody)
	if err == nil {
		taskResponse, err := ResponseJSON(resp)
		if err != nil {
			return nil, err
		}
		exitStatus, err = c.WaitForCompletion(taskResponse)
	}

	return
}
