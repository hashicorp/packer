package zstacktype

import "time"

type Image struct {
	Uuid          string
	Name          string
	BackupStorage string
	Type          string
	OSType        string
	Platform      string
	Status        string
}

type CreateImage struct {
	Name          string
	GusetOsType   string
	RootVolume    string
	Platform      string
	BackupStorage string
	Timeout       time.Duration
}

type CreateVolumeImage struct {
	Name          string
	DataVolume    string
	BackupStorage string
	Timeout       time.Duration
}
