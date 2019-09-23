package zstacktype

import "time"

type DataVolume struct {
	Uuid           string
	Name           string
	Status         string
	Type           string
	Size           uint64
	DeviceId       string
	PrimaryStorage string
}

type CreateDataVolume struct {
	Size           uint64
	Name           string
	PrimaryStorage string
	Host           string
	Timeout        time.Duration
}

type CreateDataVolumeFromImage struct {
	Uuid           string
	Name           string
	PrimaryStorage string
	Host           string
	Timeout        time.Duration
}
