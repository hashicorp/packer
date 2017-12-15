package common

type DriverMock struct {
	IsRunning_Called bool
	IsRunning_VmName string
	IsRunning_Return bool
	IsRunning_Err    error

	IsOff_Called bool
	IsOff_VmName string
	IsOff_Return bool
	IsOff_Err    error

	Uptime_Called bool
	Uptime_VmName string
	Uptime_Return uint64
	Uptime_Err    error

	Start_Called bool
	Start_VmName string
	Start_Err    error

	Stop_Called bool
	Stop_VmName string
	Stop_Err    error

	Verify_Called bool
	Verify_Err    error

	Mac_Called bool
	Mac_VmName string
	Mac_Return string
	Mac_Err    error

	IpAddress_Called bool
	IpAddress_Mac    string
	IpAddress_Return string
	IpAddress_Err    error

	GetHostName_Called bool
	GetHostName_Ip     string
	GetHostName_Return string
	GetHostName_Err    error

	GetVirtualMachineGeneration_Called bool
	GetVirtualMachineGeneration_VmName string
	GetVirtualMachineGeneration_Return uint
	GetVirtualMachineGeneration_Err    error

	GetHostAdapterIpAddressForSwitch_Called     bool
	GetHostAdapterIpAddressForSwitch_SwitchName string
	GetHostAdapterIpAddressForSwitch_Return     string
	GetHostAdapterIpAddressForSwitch_Err        error

	TypeScanCodes_Called    bool
	TypeScanCodes_VmName    string
	TypeScanCodes_ScanCodes string
	TypeScanCodes_Err       error

	GetVirtualMachineNetworkAdapterAddress_Called bool
	GetVirtualMachineNetworkAdapterAddress_VmName string
	GetVirtualMachineNetworkAdapterAddress_Return string
	GetVirtualMachineNetworkAdapterAddress_Err    error

	SetNetworkAdapterVlanId_Called     bool
	SetNetworkAdapterVlanId_SwitchName string
	SetNetworkAdapterVlanId_VlanId     string
	SetNetworkAdapterVlanId_Err        error

	SetVmNetworkAdapterMacAddress_Called bool
	SetVmNetworkAdapterMacAddress_VmName string
	SetVmNetworkAdapterMacAddress_Mac    string
	SetVmNetworkAdapterMacAddress_Err    error

	SetVirtualMachineVlanId_Called bool
	SetVirtualMachineVlanId_VmName string
	SetVirtualMachineVlanId_VlanId string
	SetVirtualMachineVlanId_Err    error

	UntagVirtualMachineNetworkAdapterVlan_Called     bool
	UntagVirtualMachineNetworkAdapterVlan_VmName     string
	UntagVirtualMachineNetworkAdapterVlan_SwitchName string
	UntagVirtualMachineNetworkAdapterVlan_Err        error

	CreateExternalVirtualSwitch_Called     bool
	CreateExternalVirtualSwitch_VmName     string
	CreateExternalVirtualSwitch_SwitchName string
	CreateExternalVirtualSwitch_Err        error

	GetVirtualMachineSwitchName_Called bool
	GetVirtualMachineSwitchName_VmName string
	GetVirtualMachineSwitchName_Return string
	GetVirtualMachineSwitchName_Err    error

	ConnectVirtualMachineNetworkAdapterToSwitch_Called     bool
	ConnectVirtualMachineNetworkAdapterToSwitch_VmName     string
	ConnectVirtualMachineNetworkAdapterToSwitch_SwitchName string
	ConnectVirtualMachineNetworkAdapterToSwitch_Err        error

	DeleteVirtualSwitch_Called     bool
	DeleteVirtualSwitch_SwitchName string
	DeleteVirtualSwitch_Err        error

	CreateVirtualSwitch_Called     bool
	CreateVirtualSwitch_SwitchName string
	CreateVirtualSwitch_SwitchType string
	CreateVirtualSwitch_Return     bool
	CreateVirtualSwitch_Err        error

	AddVirtualMachineHardDrive_Called         bool
	AddVirtualMachineHardDrive_VmName         string
	AddVirtualMachineHardDrive_VhdFile        string
	AddVirtualMachineHardDrive_VhdName        string
	AddVirtualMachineHardDrive_VhdSizeBytes   int64
	AddVirtualMachineHardDrive_ControllerType string
	AddVirtualMachineHardDrive_Err            error

	CreateVirtualMachine_Called           bool
	CreateVirtualMachine_VmName           string
	CreateVirtualMachine_Path             string
	CreateVirtualMachine_HarddrivePath    string
	CreateVirtualMachine_VhdPath          string
	CreateVirtualMachine_Ram              int64
	CreateVirtualMachine_DiskSize         int64
	CreateVirtualMachine_SwitchName       string
	CreateVirtualMachine_Generation       uint
	CreateVirtualMachine_DifferentialDisk bool
	CreateVirtualMachine_Err              error

	CloneVirtualMachine_Called                bool
	CloneVirtualMachine_CloneFromVmxcPath     string
	CloneVirtualMachine_CloneFromVmName       string
	CloneVirtualMachine_CloneFromSnapshotName string
	CloneVirtualMachine_CloneAllSnapshots     bool
	CloneVirtualMachine_VmName                string
	CloneVirtualMachine_Path                  string
	CloneVirtualMachine_HarddrivePath         string
	CloneVirtualMachine_Ram                   int64
	CloneVirtualMachine_SwitchName            string
	CloneVirtualMachine_Err                   error

	DeleteVirtualMachine_Called bool
	DeleteVirtualMachine_VmName string
	DeleteVirtualMachine_Err    error

	SetVirtualMachineCpuCount_Called bool
	SetVirtualMachineCpuCount_VmName string
	SetVirtualMachineCpuCount_Cpu    uint
	SetVirtualMachineCpuCount_Err    error

	SetVirtualMachineMacSpoofing_Called bool
	SetVirtualMachineMacSpoofing_VmName string
	SetVirtualMachineMacSpoofing_Enable bool
	SetVirtualMachineMacSpoofing_Err    error

	SetVirtualMachineDynamicMemory_Called bool
	SetVirtualMachineDynamicMemory_VmName string
	SetVirtualMachineDynamicMemory_Enable bool
	SetVirtualMachineDynamicMemory_Err    error

	SetVirtualMachineSecureBoot_Called bool
	SetVirtualMachineSecureBoot_VmName string
	SetVirtualMachineSecureBoot_Enable bool
	SetVirtualMachineSecureBoot_Err    error

	SetVirtualMachineVirtualizationExtensions_Called bool
	SetVirtualMachineVirtualizationExtensions_VmName string
	SetVirtualMachineVirtualizationExtensions_Enable bool
	SetVirtualMachineVirtualizationExtensions_Err    error

	EnableVirtualMachineIntegrationService_Called                 bool
	EnableVirtualMachineIntegrationService_VmName                 string
	EnableVirtualMachineIntegrationService_IntegrationServiceName string
	EnableVirtualMachineIntegrationService_Err                    error

	ExportVirtualMachine_Called bool
	ExportVirtualMachine_VmName string
	ExportVirtualMachine_Path   string
	ExportVirtualMachine_Err    error

	CompactDisks_Called  bool
	CompactDisks_ExpPath string
	CompactDisks_VhdDir  string
	CompactDisks_Err     error

	CopyExportedVirtualMachine_Called     bool
	CopyExportedVirtualMachine_ExpPath    string
	CopyExportedVirtualMachine_OutputPath string
	CopyExportedVirtualMachine_VhdDir     string
	CopyExportedVirtualMachine_VmDir      string
	CopyExportedVirtualMachine_Err        error

	RestartVirtualMachine_Called bool
	RestartVirtualMachine_VmName string
	RestartVirtualMachine_Err    error

	CreateDvdDrive_Called             bool
	CreateDvdDrive_VmName             string
	CreateDvdDrive_IsoPath            string
	CreateDvdDrive_Generation         uint
	CreateDvdDrive_ControllerNumber   uint
	CreateDvdDrive_ControllerLocation uint
	CreateDvdDrive_Err                error

	MountDvdDrive_Called             bool
	MountDvdDrive_VmName             string
	MountDvdDrive_Path               string
	MountDvdDrive_ControllerNumber   uint
	MountDvdDrive_ControllerLocation uint
	MountDvdDrive_Err                error

	SetBootDvdDrive_Called             bool
	SetBootDvdDrive_VmName             string
	SetBootDvdDrive_ControllerNumber   uint
	SetBootDvdDrive_ControllerLocation uint
	SetBootDvdDrive_Generation         uint
	SetBootDvdDrive_Err                error

	UnmountDvdDrive_Called             bool
	UnmountDvdDrive_VmName             string
	UnmountDvdDrive_ControllerNumber   uint
	UnmountDvdDrive_ControllerLocation uint
	UnmountDvdDrive_Err                error

	DeleteDvdDrive_Called             bool
	DeleteDvdDrive_VmName             string
	DeleteDvdDrive_ControllerNumber   uint
	DeleteDvdDrive_ControllerLocation uint
	DeleteDvdDrive_Err                error

	MountFloppyDrive_Called bool
	MountFloppyDrive_VmName string
	MountFloppyDrive_Path   string
	MountFloppyDrive_Err    error

	UnmountFloppyDrive_Called bool
	UnmountFloppyDrive_VmName string
	UnmountFloppyDrive_Err    error
}

func (d *DriverMock) IsRunning(vmName string) (bool, error) {
	d.IsRunning_Called = true
	d.IsRunning_VmName = vmName
	return d.IsRunning_Return, d.IsRunning_Err
}

func (d *DriverMock) IsOff(vmName string) (bool, error) {
	d.IsOff_Called = true
	d.IsOff_VmName = vmName
	return d.IsOff_Return, d.IsOff_Err
}

func (d *DriverMock) Uptime(vmName string) (uint64, error) {
	d.Uptime_Called = true
	d.Uptime_VmName = vmName
	return d.Uptime_Return, d.Uptime_Err
}

func (d *DriverMock) Start(vmName string) error {
	d.Start_Called = true
	d.Start_VmName = vmName
	return d.Start_Err
}

func (d *DriverMock) Stop(vmName string) error {
	d.Stop_Called = true
	d.Stop_VmName = vmName
	return d.Stop_Err
}

func (d *DriverMock) Verify() error {
	d.Verify_Called = true
	return d.Verify_Err
}

func (d *DriverMock) Mac(vmName string) (string, error) {
	d.Mac_Called = true
	d.Mac_VmName = vmName
	return d.Mac_Return, d.Mac_Err
}

func (d *DriverMock) IpAddress(mac string) (string, error) {
	d.IpAddress_Called = true
	d.IpAddress_Mac = mac
	return d.IpAddress_Return, d.IpAddress_Err
}

func (d *DriverMock) GetHostName(ip string) (string, error) {
	d.GetHostName_Called = true
	d.GetHostName_Ip = ip
	return d.GetHostName_Return, d.GetHostName_Err
}

func (d *DriverMock) GetVirtualMachineGeneration(vmName string) (uint, error) {
	d.GetVirtualMachineGeneration_Called = true
	d.GetVirtualMachineGeneration_VmName = vmName
	return d.GetVirtualMachineGeneration_Return, d.GetVirtualMachineGeneration_Err
}

func (d *DriverMock) GetHostAdapterIpAddressForSwitch(switchName string) (string, error) {
	d.GetHostAdapterIpAddressForSwitch_Called = true
	d.GetHostAdapterIpAddressForSwitch_SwitchName = switchName
	return d.GetHostAdapterIpAddressForSwitch_Return, d.GetHostAdapterIpAddressForSwitch_Err
}

func (d *DriverMock) TypeScanCodes(vmName string, scanCodes string) error {
	d.TypeScanCodes_Called = true
	d.TypeScanCodes_VmName = vmName
	d.TypeScanCodes_ScanCodes = scanCodes
	return d.TypeScanCodes_Err
}

func (d *DriverMock) GetVirtualMachineNetworkAdapterAddress(vmName string) (string, error) {
	d.GetVirtualMachineNetworkAdapterAddress_Called = true
	d.GetVirtualMachineNetworkAdapterAddress_VmName = vmName
	return d.GetVirtualMachineNetworkAdapterAddress_Return, d.GetVirtualMachineNetworkAdapterAddress_Err
}

func (d *DriverMock) SetNetworkAdapterVlanId(switchName string, vlanId string) error {
	d.SetNetworkAdapterVlanId_Called = true
	d.SetNetworkAdapterVlanId_SwitchName = switchName
	d.SetNetworkAdapterVlanId_VlanId = vlanId
	return d.SetNetworkAdapterVlanId_Err
}

func (d *DriverMock) SetVmNetworkAdapterMacAddress(vmName string, mac string) error {
	d.SetVmNetworkAdapterMacAddress_Called = true
	d.SetVmNetworkAdapterMacAddress_VmName = vmName
	d.SetVmNetworkAdapterMacAddress_Mac = mac
	return d.SetVmNetworkAdapterMacAddress_Err
}

func (d *DriverMock) SetVirtualMachineVlanId(vmName string, vlanId string) error {
	d.SetVirtualMachineVlanId_Called = true
	d.SetVirtualMachineVlanId_VmName = vmName
	d.SetVirtualMachineVlanId_VlanId = vlanId
	return d.SetVirtualMachineVlanId_Err
}

func (d *DriverMock) UntagVirtualMachineNetworkAdapterVlan(vmName string, switchName string) error {
	d.UntagVirtualMachineNetworkAdapterVlan_Called = true
	d.UntagVirtualMachineNetworkAdapterVlan_VmName = vmName
	d.UntagVirtualMachineNetworkAdapterVlan_SwitchName = switchName
	return d.UntagVirtualMachineNetworkAdapterVlan_Err
}

func (d *DriverMock) CreateExternalVirtualSwitch(vmName string, switchName string) error {
	d.CreateExternalVirtualSwitch_Called = true
	d.CreateExternalVirtualSwitch_VmName = vmName
	d.CreateExternalVirtualSwitch_SwitchName = switchName
	return d.CreateExternalVirtualSwitch_Err
}

func (d *DriverMock) GetVirtualMachineSwitchName(vmName string) (string, error) {
	d.GetVirtualMachineSwitchName_Called = true
	d.GetVirtualMachineSwitchName_VmName = vmName
	return d.GetVirtualMachineSwitchName_Return, d.GetVirtualMachineSwitchName_Err
}

func (d *DriverMock) ConnectVirtualMachineNetworkAdapterToSwitch(vmName string, switchName string) error {
	d.ConnectVirtualMachineNetworkAdapterToSwitch_Called = true
	d.ConnectVirtualMachineNetworkAdapterToSwitch_VmName = vmName
	d.ConnectVirtualMachineNetworkAdapterToSwitch_SwitchName = switchName
	return d.ConnectVirtualMachineNetworkAdapterToSwitch_Err
}

func (d *DriverMock) DeleteVirtualSwitch(switchName string) error {
	d.DeleteVirtualSwitch_Called = true
	d.DeleteVirtualSwitch_SwitchName = switchName
	return d.DeleteVirtualSwitch_Err
}

func (d *DriverMock) CreateVirtualSwitch(switchName string, switchType string) (bool, error) {
	d.CreateVirtualSwitch_Called = true
	d.CreateVirtualSwitch_SwitchName = switchName
	d.CreateVirtualSwitch_SwitchType = switchType
	return d.CreateVirtualSwitch_Return, d.CreateVirtualSwitch_Err
}

func (d *DriverMock) AddVirtualMachineHardDrive(vmName string, vhdFile string, vhdName string, vhdSizeBytes int64, controllerType string) error {
	d.AddVirtualMachineHardDrive_Called = true
	d.AddVirtualMachineHardDrive_VmName = vmName
	d.AddVirtualMachineHardDrive_VhdFile = vhdFile
	d.AddVirtualMachineHardDrive_VhdName = vhdName
	d.AddVirtualMachineHardDrive_VhdSizeBytes = vhdSizeBytes
	d.AddVirtualMachineHardDrive_ControllerType = controllerType
	return d.AddVirtualMachineHardDrive_Err
}

func (d *DriverMock) CreateVirtualMachine(vmName string, path string, harddrivePath string, vhdPath string, ram int64, diskSize int64, switchName string, generation uint, diffDisks bool) error {
	d.CreateVirtualMachine_Called = true
	d.CreateVirtualMachine_VmName = vmName
	d.CreateVirtualMachine_Path = path
	d.CreateVirtualMachine_HarddrivePath = harddrivePath
	d.CreateVirtualMachine_VhdPath = vhdPath
	d.CreateVirtualMachine_Ram = ram
	d.CreateVirtualMachine_DiskSize = diskSize
	d.CreateVirtualMachine_SwitchName = switchName
	d.CreateVirtualMachine_Generation = generation
	d.CreateVirtualMachine_DifferentialDisk = diffDisks
	return d.CreateVirtualMachine_Err
}

func (d *DriverMock) CloneVirtualMachine(cloneFromVmxcPath string, cloneFromVmName string, cloneFromSnapshotName string, cloneAllSnapshots bool, vmName string, path string, harddrivePath string, ram int64, switchName string) error {
	d.CloneVirtualMachine_Called = true
	d.CloneVirtualMachine_CloneFromVmxcPath = cloneFromVmxcPath
	d.CloneVirtualMachine_CloneFromVmName = cloneFromVmName
	d.CloneVirtualMachine_CloneFromSnapshotName = cloneFromSnapshotName
	d.CloneVirtualMachine_CloneAllSnapshots = cloneAllSnapshots
	d.CloneVirtualMachine_VmName = vmName
	d.CloneVirtualMachine_Path = path
	d.CloneVirtualMachine_HarddrivePath = harddrivePath
	d.CloneVirtualMachine_Ram = ram
	d.CloneVirtualMachine_SwitchName = switchName
	return d.CloneVirtualMachine_Err
}

func (d *DriverMock) DeleteVirtualMachine(vmName string) error {
	d.DeleteVirtualMachine_Called = true
	d.DeleteVirtualMachine_VmName = vmName
	return d.DeleteVirtualMachine_Err
}

func (d *DriverMock) SetVirtualMachineCpuCount(vmName string, cpu uint) error {
	d.SetVirtualMachineCpuCount_Called = true
	d.SetVirtualMachineCpuCount_VmName = vmName
	d.SetVirtualMachineCpuCount_Cpu = cpu
	return d.SetVirtualMachineCpuCount_Err
}

func (d *DriverMock) SetVirtualMachineMacSpoofing(vmName string, enable bool) error {
	d.SetVirtualMachineMacSpoofing_Called = true
	d.SetVirtualMachineMacSpoofing_VmName = vmName
	d.SetVirtualMachineMacSpoofing_Enable = enable
	return d.SetVirtualMachineMacSpoofing_Err
}

func (d *DriverMock) SetVirtualMachineDynamicMemory(vmName string, enable bool) error {
	d.SetVirtualMachineDynamicMemory_Called = true
	d.SetVirtualMachineDynamicMemory_VmName = vmName
	d.SetVirtualMachineDynamicMemory_Enable = enable
	return d.SetVirtualMachineDynamicMemory_Err
}

func (d *DriverMock) SetVirtualMachineSecureBoot(vmName string, enable bool) error {
	d.SetVirtualMachineSecureBoot_Called = true
	d.SetVirtualMachineSecureBoot_VmName = vmName
	d.SetVirtualMachineSecureBoot_Enable = enable
	return d.SetVirtualMachineSecureBoot_Err
}

func (d *DriverMock) SetVirtualMachineVirtualizationExtensions(vmName string, enable bool) error {
	d.SetVirtualMachineVirtualizationExtensions_Called = true
	d.SetVirtualMachineVirtualizationExtensions_VmName = vmName
	d.SetVirtualMachineVirtualizationExtensions_Enable = enable
	return d.SetVirtualMachineVirtualizationExtensions_Err
}

func (d *DriverMock) EnableVirtualMachineIntegrationService(vmName string, integrationServiceName string) error {
	d.EnableVirtualMachineIntegrationService_Called = true
	d.EnableVirtualMachineIntegrationService_VmName = vmName
	d.EnableVirtualMachineIntegrationService_IntegrationServiceName = integrationServiceName
	return d.EnableVirtualMachineIntegrationService_Err
}

func (d *DriverMock) ExportVirtualMachine(vmName string, path string) error {
	d.ExportVirtualMachine_Called = true
	d.ExportVirtualMachine_VmName = vmName
	d.ExportVirtualMachine_Path = path
	return d.ExportVirtualMachine_Err
}

func (d *DriverMock) CompactDisks(expPath string, vhdDir string) error {
	d.CompactDisks_Called = true
	d.CompactDisks_ExpPath = expPath
	d.CompactDisks_VhdDir = vhdDir
	return d.CompactDisks_Err
}

func (d *DriverMock) CopyExportedVirtualMachine(expPath string, outputPath string, vhdDir string, vmDir string) error {
	d.CopyExportedVirtualMachine_Called = true
	d.CopyExportedVirtualMachine_ExpPath = expPath
	d.CopyExportedVirtualMachine_OutputPath = outputPath
	d.CopyExportedVirtualMachine_VhdDir = vhdDir
	d.CopyExportedVirtualMachine_VmDir = vmDir
	return d.CopyExportedVirtualMachine_Err
}

func (d *DriverMock) RestartVirtualMachine(vmName string) error {
	d.RestartVirtualMachine_Called = true
	d.RestartVirtualMachine_VmName = vmName
	return d.RestartVirtualMachine_Err
}

func (d *DriverMock) CreateDvdDrive(vmName string, isoPath string, generation uint) (uint, uint, error) {
	d.CreateDvdDrive_Called = true
	d.CreateDvdDrive_VmName = vmName
	d.CreateDvdDrive_IsoPath = isoPath
	d.CreateDvdDrive_Generation = generation
	return d.CreateDvdDrive_ControllerNumber, d.CreateDvdDrive_ControllerLocation, d.CreateDvdDrive_Err
}

func (d *DriverMock) MountDvdDrive(vmName string, path string, controllerNumber uint, controllerLocation uint) error {
	d.MountDvdDrive_Called = true
	d.MountDvdDrive_VmName = vmName
	d.MountDvdDrive_Path = path
	d.MountDvdDrive_ControllerNumber = controllerNumber
	d.MountDvdDrive_ControllerLocation = controllerLocation
	return d.MountDvdDrive_Err
}

func (d *DriverMock) SetBootDvdDrive(vmName string, controllerNumber uint, controllerLocation uint, generation uint) error {
	d.SetBootDvdDrive_Called = true
	d.SetBootDvdDrive_VmName = vmName
	d.SetBootDvdDrive_ControllerNumber = controllerNumber
	d.SetBootDvdDrive_ControllerLocation = controllerLocation
	d.SetBootDvdDrive_Generation = generation
	return d.SetBootDvdDrive_Err
}

func (d *DriverMock) UnmountDvdDrive(vmName string, controllerNumber uint, controllerLocation uint) error {
	d.UnmountDvdDrive_Called = true
	d.UnmountDvdDrive_VmName = vmName
	d.UnmountDvdDrive_ControllerNumber = controllerNumber
	d.UnmountDvdDrive_ControllerLocation = controllerLocation
	return d.UnmountDvdDrive_Err
}

func (d *DriverMock) DeleteDvdDrive(vmName string, controllerNumber uint, controllerLocation uint) error {
	d.DeleteDvdDrive_Called = true
	d.DeleteDvdDrive_VmName = vmName
	d.DeleteDvdDrive_ControllerNumber = controllerNumber
	d.DeleteDvdDrive_ControllerLocation = controllerLocation
	return d.DeleteDvdDrive_Err
}

func (d *DriverMock) MountFloppyDrive(vmName string, path string) error {
	d.MountFloppyDrive_Called = true
	d.MountFloppyDrive_VmName = vmName
	d.MountFloppyDrive_Path = path
	return d.MountFloppyDrive_Err
}

func (d *DriverMock) UnmountFloppyDrive(vmName string) error {
	d.UnmountFloppyDrive_Called = true
	d.UnmountFloppyDrive_VmName = vmName
	return d.UnmountFloppyDrive_Err
}
