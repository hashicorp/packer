package testing

// Defines whether acceptance tests should be run
const TestEnvVar = "VSPHERE_DRIVER_ACC"

// Describe the environment to run tests in
const TestDatastore = "datastore1"
const TestTemplate = "alpine"
const TestHost = "esxi-1.vsphere55.test"
const TestVCenterServer = "vcenter.vsphere55.test"
const TestVCenterUsername = "root"
const TestVCenterPassword = "jetbrains"

// For test of hardware settings
const TestCPUs = 2
const TestCPUReservation = 1000
const TestCPULimit = 1500
const TestRAM = 2048
const TestRAMReservation = 1024

const TestFolder = "folder1/folder2"
const TestResourcePool = "pool1/pool2"
