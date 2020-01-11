package proxmox
import (
	"io"
	"github.com/Telmate/proxmox-api-go/proxmox"
)

type vmIdMock struct {
	getVmList func() (map[string]interface{})
	getNextID func(int) (int)
}
func (m vmIdMock) StartVm(vmref *proxmox.VmRef) (string, error) {
	return "", nil
}
func (m vmIdMock) GetVmList() (map[string]interface{}, error) {
	return m.getVmList(), nil
}
func (m vmIdMock) GetNextID(currentId int) (int, error) {
	return m.getNextID(currentId), nil
}
func (m vmIdMock) createVMDisks(string, map[string]interface{}) (int, error) {
	return 0, nil
}
func (m vmIdMock) CheckVmRef(vmref *proxmox.VmRef) (error) {
	return nil
}
func (m vmIdMock) CloneQemuVm(*proxmox.VmRef, map[string]interface{}) (string, error) {
	return "", nil
}
func (m vmIdMock) CreateLxcContainer(string, map[string]interface{}) (string, error) {
	return "", nil
}
func (m vmIdMock) CreateQemuVm(string, map[string]interface{}) (string, error) {
	return "", nil
}
func (m vmIdMock) CreateTemplate(*proxmox.VmRef) error {
	return nil
}
func (m vmIdMock) CreateVMDisk(string, string, string, map[string]interface{}) (error) {
	return nil
}
func (m vmIdMock) DeleteVMDisks(string, []string) (error) {
	return nil
}
func (m vmIdMock) DeleteVm(*proxmox.VmRef) (string, error) {
	return "", nil
}
func (m vmIdMock) GetJsonRetryable(string, *map[string]interface{}, int) error {
	return nil
}
func (m vmIdMock) GetNodeList() (map[string]interface{}, error) {
	return nil, nil
}
func (m vmIdMock) GetTaskExitstatus(string) (interface{}, error) {
	return nil, nil
}
func (m vmIdMock) GetVmAgentNetworkInterfaces(*proxmox.VmRef) ([]proxmox.AgentNetworkInterface, error) {
	return nil, nil
}
func (m vmIdMock) GetVmConfig(*proxmox.VmRef) (map[string]interface{}, error) {
	return nil, nil
}
func (m vmIdMock) GetVmInfo(*proxmox.VmRef) (map[string]interface{}, error) {
	return nil, nil
}
func (m vmIdMock) GetVmRefByName(string) (*proxmox.VmRef, error) {
	return nil, nil
}
func (m vmIdMock) GetVmSpiceProxy(*proxmox.VmRef) (map[string]interface{}, error) {
	return nil, nil
}
func (m vmIdMock) GetVmState(*proxmox.VmRef) (map[string]interface{}, error) {
	return nil, nil
}
func (m vmIdMock) Login(string, string, string) (error) {
	return nil
}
func (m vmIdMock) MonitorCmd(*proxmox.VmRef, string) (map[string]interface{}, error) {
	return nil, nil
}
func (m vmIdMock) MoveQemuDisk(*proxmox.VmRef, string, string) (interface{}, error) {
	return nil, nil
}
func (m vmIdMock) ResetVm(*proxmox.VmRef) (string, error) {
	return "", nil
}
func (m vmIdMock) ResizeQemuDisk(*proxmox.VmRef, string, int) (interface{}, error) {
	return nil, nil
}
func (m vmIdMock) ResumeVm(*proxmox.VmRef) (string, error) {
	return "", nil
}
func (m vmIdMock) RollbackQemuVm(*proxmox.VmRef, string) (string, error) {
	return "", nil
}
func (m vmIdMock) Sendkey(*proxmox.VmRef, string) error {
	return nil
}
func (m vmIdMock) SetLxcConfig(*proxmox.VmRef, map[string]interface{}) (interface{}, error) {
	return nil, nil
}
func (m vmIdMock) SetVmConfig(*proxmox.VmRef, map[string]interface{}) (interface{}, error) {
	return nil, nil
}
func (m vmIdMock) ShutdownVm(*proxmox.VmRef) (string, error) {
	return "", nil
}
func (m vmIdMock) StatusChangeVm(*proxmox.VmRef, string) (string, error) {
	return "", nil
}
func (m vmIdMock) StopVm(*proxmox.VmRef) (string, error) {
	return "", nil
}
func (m vmIdMock) SuspendVm(*proxmox.VmRef) (string, error) {
	return "", nil
}
func (m vmIdMock) UpdateVMHA(*proxmox.VmRef, string) (interface{}, error) {
	return nil, nil
}
func (m vmIdMock) UpdateVMPool(*proxmox.VmRef, string) (interface{}, error) {
	return nil, nil
}
func (m vmIdMock) Upload(string, string, string, string, io.Reader) error {
	return nil
}
func (m vmIdMock) WaitForCompletion(map[string]interface{}) (string, error) {
	return "", nil
}

var _ proxmox.PveClient = &vmIdMock{}
