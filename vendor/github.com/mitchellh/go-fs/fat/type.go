package fat

import "github.com/mitchellh/go-fs"

// FATType is a simple enum of the available FAT filesystem types.
type FATType uint8

const (
	FAT12 FATType = iota
	FAT16
	FAT32
)

// TypeForDevice determines the usable FAT type based solely on
// size information about the block device.
func TypeForDevice(device fs.BlockDevice) FATType {
	sizeInMB := device.Len() / (1024 * 1024)
	switch {
	case sizeInMB < 4:
		return FAT12
	case sizeInMB < 512:
		return FAT16
	default:
		return FAT32
	}
}
