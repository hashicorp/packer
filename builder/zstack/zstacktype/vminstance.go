package zstacktype

import "time"

type VmInstance struct {
	Uuid       string
	Name       string
	PublicIp   string
	State      string
	RootVolume string
	Host       string
}

type CreateVm struct {
	Name             string
	L3               string
	InstanceOffering string
	Image            string
	Sshkey           string
	UserData         string
	Timeout          time.Duration
}
