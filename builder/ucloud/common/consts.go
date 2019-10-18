package common

const (
	// DefaultPasswordStr, DefaultPasswordNum and DefaultPasswordSpe are used to general default value of root password of UHost instance
	DefaultPasswordNum = "012346789"
	DefaultPasswordStr = "abcdefghijklmnopqrstuvwxyz"
	DefaultPasswordSpe = "-_"
)

const (
	OsTypeWindows         = "Windows"
	SecurityGroupNonWeb   = "recommend non web"
	InstanceStateRunning  = "Running"
	InstanceStateStopped  = "Stopped"
	BootDiskStateNormal   = "Normal"
	ImageStateAvailable   = "Available"
	ImageStateUnavailable = "Unavailable"
	IpTypePrivate         = "Private"
)

const (
	DefaultCreateImageTimeOut = 3600
)

var BootDiskTypeMap = NewStringConverter(map[string]string{
	"cloud_ssd":    "CLOUD_SSD",
	"local_normal": "LOCAL_NORMAL",
	"local_ssd":    "LOCAL_SSD",
})
