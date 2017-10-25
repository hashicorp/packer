package testing

// Defines whether acceptance tests should be run
const TestEnvVar = "VSPHERE_DRIVER_ACC"

// Describe the environment to run tests in
const DefaultDatastore = "datastore1"
const DefaultTemplate = "alpine"
const DefaultHost = "esxi-1.vsphere55.test"
const DefaultVCenterServer = "vcenter.vsphere55.test"
const DefaultVCenterUsername = "root"
const DefaultVCenterPassword = "jetbrains"

// Default hardware settings for
const DefaultCPUs = 2
const DefaultCPUReservation = 1000
const DefaultCPULimit = 1500
const DefaultRAM = 2048
const DefaultRAMReservation = 1024
