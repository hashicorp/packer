package fat

import (
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/go-fs"
)

// SuperFloppyConfig is the configuration for various properties of
// a new super floppy formatted block device. Once this configuration is used
// to format a device, it must not be modified.
type SuperFloppyConfig struct {
	// The type of FAT filesystem to use.
	FATType FATType

	// The label of the drive. Defaults to "NONAME"
	Label string

	// The OEM name for the FAT filesystem. Defaults to "gofs" if not set.
	OEMName string
}

// Formats an fs.BlockDevice with the "super floppy" format according
// to the given configuration. The "super floppy" standard means that the
// device will be formatted so that it does not contain a partition table.
// Instead, the entire device holds a single FAT file system.
func FormatSuperFloppy(device fs.BlockDevice, config *SuperFloppyConfig) error {
	formatter := &superFloppyFormatter{
		config: config,
		device: device,
	}

	return formatter.format()
}

// An internal struct that helps maintain state and perform calculations
// during a single formatting pass.
type superFloppyFormatter struct {
	config *SuperFloppyConfig
	device fs.BlockDevice
}

func (f *superFloppyFormatter) format() error {
	// First, create the boot sector on the device. Start by configuring
	// the common elements of the boot sector.
	sectorsPerCluster, err := f.SectorsPerCluster()
	if err != nil {
		return err
	}

	bsCommon := BootSectorCommon{
		BytesPerSector:      uint16(f.device.SectorSize()),
		Media:               MediaFixed,
		NumFATs:             2,
		NumHeads:            16,
		OEMName:             f.config.OEMName,
		ReservedSectorCount: f.ReservedSectorCount(),
		SectorsPerCluster:   sectorsPerCluster,
		SectorsPerTrack:     32,
		TotalSectors:        uint32(f.device.Len() / int64(f.device.SectorSize())),
	}

	// Next, fill in the FAT-type specific boot sector information
	switch f.config.FATType {
	case FAT12, FAT16:
		// Determine the filesystem type label, standard from the spec sheet
		var label string
		if f.config.FATType == FAT12 {
			label = "FAT12   "
		} else {
			label = "FAT16   "
		}

		// For 1.44MB Floppy, for other floppy formats see https://support.microsoft.com/en-us/kb/75131.
		// We make an exception for this most common usecase as the calculations don't create a working image for older operating systems
		if f.config.FATType == FAT12 && f.device.Len() == 1474560 {
			bsCommon.RootEntryCount = 224
			bsCommon.SectorsPerFat = 9
			bsCommon.SectorsPerTrack = 18
			bsCommon.Media = 240
			bsCommon.NumHeads = 2
		} else {
			// Determine the number of root directory entries
			if f.device.Len() > 512*5*32 {
				bsCommon.RootEntryCount = 512
			} else {
				bsCommon.RootEntryCount = uint16(f.device.Len() / (5 * 32))
			}

			bsCommon.SectorsPerFat = f.sectorsPerFat(bsCommon.RootEntryCount, sectorsPerCluster)
		}

		bs := &BootSectorFat16{
			BootSectorCommon:    bsCommon,
			FileSystemTypeLabel: label,
			VolumeLabel:         f.config.Label,
		}

		// Write the boot sector
		bsBytes, err := bs.Bytes()
		if err != nil {
			return err
		}

		if _, err := f.device.WriteAt(bsBytes, 0); err != nil {
			return err
		}
	case FAT32:
		bsCommon.SectorsPerFat = f.sectorsPerFat(0, sectorsPerCluster)

		bs := &BootSectorFat32{
			BootSectorCommon:    bsCommon,
			FileSystemTypeLabel: "FAT32   ",
			FSInfoSector:        1,
			VolumeID:            uint32(time.Now().Unix()),
			VolumeLabel:         f.config.Label,
		}

		// Write the boot sector
		bsBytes, err := bs.Bytes()
		if err != nil {
			return err
		}

		if _, err := f.device.WriteAt(bsBytes, 0); err != nil {
			return err
		}

		// TODO(mitchellh): Create the fsinfo structure
		// TODO(mitchellh): write the boot sector copy
	default:
		return fmt.Errorf("Unknown FAT type: %d", f.config.FATType)
	}

	// Create the FATs
	fat, err := NewFAT(&bsCommon)
	if err != nil {
		return err
	}

	// Write the FAT
	if err := fat.WriteToDevice(f.device); err != nil {
		return err
	}

	var rootDir *DirectoryCluster
	if f.config.FATType == FAT32 {
		panic("TODO")
	} else {
		rootDir, err = NewFat16RootDirectoryCluster(&bsCommon, f.config.Label)
		if err != nil {
			return err
		}

		offset := int64(bsCommon.RootDirOffset())
		if _, err := f.device.WriteAt(rootDir.Bytes(), offset); err != nil {
			return err
		}
	}

	return nil
}

func (f *superFloppyFormatter) ReservedSectorCount() uint16 {
	if f.config.FATType == FAT32 {
		return 32
	} else {
		return 1
	}
}

func (f *superFloppyFormatter) SectorsPerCluster() (uint8, error) {
	if f.config.FATType == FAT12 {
		return f.defaultSectorsPerCluster12()
	} else if f.config.FATType == FAT16 {
		return f.defaultSectorsPerCluster16()
	} else {
		return f.defaultSectorsPerCluster32()
	}
}

func (f *superFloppyFormatter) defaultSectorsPerCluster12() (uint8, error) {
	var result uint8 = 1
	sectors := f.device.Len() / int64(f.device.SectorSize())

	for (sectors / int64(result)) > 4084 {
		result *= 2
		if int(result)*f.device.SectorSize() > 4096 {
			return 0, errors.New("disk too large for FAT12")
		}
	}

	return result, nil
}

func (f *superFloppyFormatter) defaultSectorsPerCluster16() (uint8, error) {
	sectors := f.device.Len() / int64(f.device.SectorSize())

	if sectors <= 8400 {
		return 0, errors.New("disk too small for FAT16")
	} else if sectors > 4194304 {
		return 0, errors.New("disk too large for FAT16")
	}

	switch {
	case sectors > 2097152:
		return 64, nil
	case sectors > 1048576:
		return 32, nil
	case sectors > 524288:
		return 16, nil
	case sectors > 262144:
		return 8, nil
	case sectors > 32680:
		return 4, nil
	default:
		return 2, nil
	}
}

func (f *superFloppyFormatter) defaultSectorsPerCluster32() (uint8, error) {
	sectors := f.device.Len() / int64(f.device.SectorSize())

	if sectors <= 66600 {
		return 0, errors.New("disk too small for FAT32")
	}

	switch {
	case sectors > 67108864:
		return 64, nil
	case sectors > 33554432:
		return 32, nil
	case sectors > 16777216:
		return 16, nil
	case sectors > 532480:
		return 8, nil
	default:
		return 1, nil
	}
}

func (f *superFloppyFormatter) fatCount() uint8 {
	return 2
}

func (f *superFloppyFormatter) sectorsPerFat(rootEntCount uint16, sectorsPerCluster uint8) uint32 {
	bytesPerSec := f.device.SectorSize()
	totalSectors := int(f.device.Len()) / bytesPerSec
	rootDirSectors := ((int(rootEntCount) * 32) + (bytesPerSec - 1)) / bytesPerSec

	tmp1 := totalSectors - (int(f.ReservedSectorCount()) + rootDirSectors)
	tmp2 := (256 * int(sectorsPerCluster)) + int(f.fatCount())

	if f.config.FATType == FAT32 {
		tmp2 /= 2
	}

	return uint32((tmp1 + (tmp2 - 1)) / tmp2)
}
