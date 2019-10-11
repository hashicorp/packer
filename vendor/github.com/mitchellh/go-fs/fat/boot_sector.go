package fat

import (
	"encoding/binary"
	"errors"
	"fmt"
	"unicode"

	"github.com/mitchellh/go-fs"
)

type MediaType uint8

// The standard value for "fixed", non-removable media, directly
// from the FAT specification.
const MediaFixed MediaType = 0xF8

type BootSectorCommon struct {
	OEMName             string
	BytesPerSector      uint16
	SectorsPerCluster   uint8
	ReservedSectorCount uint16
	NumFATs             uint8
	RootEntryCount      uint16
	TotalSectors        uint32
	Media               MediaType
	SectorsPerFat       uint32
	SectorsPerTrack     uint16
	NumHeads            uint16
}

// DecodeBootSector takes a BlockDevice and decodes the FAT boot sector
// from it.
func DecodeBootSector(device fs.BlockDevice) (*BootSectorCommon, error) {
	var sector [512]byte
	if _, err := device.ReadAt(sector[:], 0); err != nil {
		return nil, err
	}

	if sector[510] != 0x55 || sector[511] != 0xAA {
		return nil, errors.New("corrupt boot sector signature")
	}

	result := new(BootSectorCommon)

	// BS_OEMName
	result.OEMName = string(sector[3:11])

	// BPB_BytsPerSec
	result.BytesPerSector = binary.LittleEndian.Uint16(sector[11:13])

	// BPB_SecPerClus
	result.SectorsPerCluster = sector[13]

	// BPB_RsvdSecCnt
	result.ReservedSectorCount = binary.LittleEndian.Uint16(sector[14:16])

	// BPB_NumFATs
	result.NumFATs = sector[16]

	// BPB_RootEntCnt
	result.RootEntryCount = binary.LittleEndian.Uint16(sector[17:19])

	// BPB_Media
	result.Media = MediaType(sector[21])

	// BPB_SecPerTrk
	result.SectorsPerTrack = binary.LittleEndian.Uint16(sector[24:26])

	// BPB_NumHeads
	result.NumHeads = binary.LittleEndian.Uint16(sector[26:28])

	// BPB_TotSec16 / BPB_TotSec32
	result.TotalSectors = uint32(binary.LittleEndian.Uint16(sector[19:21]))
	if result.TotalSectors == 0 {
		result.TotalSectors = binary.LittleEndian.Uint32(sector[32:36])
	}

	// BPB_FATSz16 / BPB_FATSz32
	result.SectorsPerFat = uint32(binary.LittleEndian.Uint16(sector[22:24]))
	if result.SectorsPerFat == 0 {
		result.SectorsPerFat = binary.LittleEndian.Uint32(sector[36:40])
	}

	return result, nil
}

func (b *BootSectorCommon) Bytes() ([]byte, error) {
	var sector [512]byte

	// BS_jmpBoot
	sector[0] = 0xEB
	sector[1] = 0x3C
	sector[2] = 0x90

	// BS_OEMName
	if len(b.OEMName) > 8 {
		return nil, errors.New("OEMName must be 8 bytes or less")
	}

	for i, r := range b.OEMName {
		if r > unicode.MaxASCII {
			return nil, fmt.Errorf("%#U in OEM name not a valid ASCII char. Must be ASCII.", r)
		}

		sector[0x3+i] = byte(r)
	}

	// BPB_BytsPerSec
	binary.LittleEndian.PutUint16(sector[11:13], b.BytesPerSector)

	// BPB_SecPerClus
	sector[13] = uint8(b.SectorsPerCluster)

	// BPB_RsvdSecCnt
	binary.LittleEndian.PutUint16(sector[14:16], b.ReservedSectorCount)

	// BPB_NumFATs
	sector[16] = b.NumFATs

	// BPB_RootEntCnt
	binary.LittleEndian.PutUint16(sector[17:19], b.RootEntryCount)

	// BPB_Media
	sector[21] = byte(b.Media)

	// BPB_SecPerTrk
	binary.LittleEndian.PutUint16(sector[24:26], b.SectorsPerTrack)

	// BPB_Numheads
	binary.LittleEndian.PutUint16(sector[26:28], b.NumHeads)

	// BPB_Hiddsec
	// sector[28:32] - it is always set to 0 because we don't partition drives yet.

	// Important signature of every FAT boot sector
	sector[510] = 0x55
	sector[511] = 0xAA

	return sector[:], nil
}

// BytesPerCluster returns the number of bytes per cluster.
func (b *BootSectorCommon) BytesPerCluster() uint32 {
	return uint32(b.SectorsPerCluster) * uint32(b.BytesPerSector)
}

// ClusterOffset returns the offset of the data section of a particular
// cluster.
func (b *BootSectorCommon) ClusterOffset(n int) uint32 {
	offset := b.DataOffset()
	offset += (uint32(n) - FirstCluster) * b.BytesPerCluster()
	return offset
}

// DataOffset returns the offset of the data section of the disk.
func (b *BootSectorCommon) DataOffset() uint32 {
	offset := uint32(b.RootDirOffset())
	offset += uint32(b.RootEntryCount * DirectoryEntrySize)
	return offset
}

// FATOffset returns the offset in bytes for the given index of the FAT
func (b *BootSectorCommon) FATOffset(n int) int {
	offset := uint32(b.ReservedSectorCount * b.BytesPerSector)
	offset += b.SectorsPerFat * uint32(b.BytesPerSector) * uint32(n)
	return int(offset)
}

// Calculates the FAT type that this boot sector represents.
func (b *BootSectorCommon) FATType() FATType {
	var rootDirSectors uint32
	rootDirSectors = (uint32(b.RootEntryCount) * 32) + (uint32(b.BytesPerSector) - 1)
	rootDirSectors /= uint32(b.BytesPerSector)
	dataSectors := b.SectorsPerFat * uint32(b.NumFATs)
	dataSectors += uint32(b.ReservedSectorCount)
	dataSectors += rootDirSectors
	dataSectors = b.TotalSectors - dataSectors
	countClusters := dataSectors / uint32(b.SectorsPerCluster)

	switch {
	case countClusters < 4085:
		return FAT12
	case countClusters < 65525:
		return FAT16
	default:
		return FAT32
	}
}

// RootDirOffset returns the byte offset when the root directory
// entries for FAT12/16 filesystems start. NOTE: This is absolutely useless
// for FAT32 because the root directory is just the beginning of the data
// region.
func (b *BootSectorCommon) RootDirOffset() int {
	offset := b.FATOffset(0)
	offset += int(uint32(b.NumFATs) * b.SectorsPerFat * uint32(b.BytesPerSector))
	return offset
}

// BootSectorFat16 is the BootSector for FAT12 and FAT16 filesystems.
// It contains the common fields to all FAT filesystems and also some
// unique.
type BootSectorFat16 struct {
	BootSectorCommon

	DriveNumber         uint8
	VolumeID            uint32
	VolumeLabel         string
	FileSystemTypeLabel string
}

func (b *BootSectorFat16) Bytes() ([]byte, error) {
	sector, err := b.BootSectorCommon.Bytes()
	if err != nil {
		return nil, err
	}

	// BPB_TotSec16 AND BPB_TotSec32
	if b.TotalSectors < 0x10000 {
		binary.LittleEndian.PutUint16(sector[19:21], uint16(b.TotalSectors))
	} else {
		binary.LittleEndian.PutUint32(sector[32:36], b.TotalSectors)
	}

	// BPB_FATSz16
	if b.SectorsPerFat > 0x10000 {
		return nil, fmt.Errorf("SectorsPerFat value too big for non-FAT32: %d", b.SectorsPerFat)
	}

	binary.LittleEndian.PutUint16(sector[22:24], uint16(b.SectorsPerFat))

	// BS_DrvNum
	sector[36] = b.DriveNumber

	// BS_BootSig
	sector[38] = 0x29

	// BS_VolID
	binary.LittleEndian.PutUint32(sector[39:43], b.VolumeID)

	// BS_VolLab
	if len(b.VolumeLabel) > 11 {
		return nil, errors.New("VolumeLabel must be 11 bytes or less")
	}

	for i, r := range b.VolumeLabel {
		if r > unicode.MaxASCII {
			return nil, fmt.Errorf("%#U in VolumeLabel not a valid ASCII char. Must be ASCII.", r)
		}

		sector[43+i] = byte(r)
	}

	// BS_FilSysType
	if len(b.FileSystemTypeLabel) > 8 {
		return nil, errors.New("FileSystemTypeLabel must be 8 bytes or less")
	}

	for i, r := range b.FileSystemTypeLabel {
		if r > unicode.MaxASCII {
			return nil, fmt.Errorf("%#U in FileSystemTypeLabel not a valid ASCII char. Must be ASCII.", r)
		}

		sector[54+i] = byte(r)
	}

	return sector, nil
}

type BootSectorFat32 struct {
	BootSectorCommon

	RootCluster         uint32
	FSInfoSector        uint16
	BackupBootSector    uint16
	DriveNumber         uint8
	VolumeID            uint32
	VolumeLabel         string
	FileSystemTypeLabel string
}

func (b *BootSectorFat32) Bytes() ([]byte, error) {
	sector, err := b.BootSectorCommon.Bytes()
	if err != nil {
		return nil, err
	}

	// BPB_RootEntCount - must be 0
	sector[17] = 0
	sector[18] = 0

	// BPB_FATSz32
	binary.LittleEndian.PutUint32(sector[36:40], b.SectorsPerFat)

	// BPB_ExtFlags - Unused?

	// BPB_FSVer. Explicitly set to 0 because that is really important
	// to get correct.
	sector[42] = 0
	sector[43] = 0

	// BPB_RootClus
	binary.LittleEndian.PutUint32(sector[44:48], b.RootCluster)

	// BPB_FSInfo
	binary.LittleEndian.PutUint16(sector[48:50], b.FSInfoSector)

	// BPB_BkBootSec
	binary.LittleEndian.PutUint16(sector[50:52], b.BackupBootSector)

	// BS_DrvNum
	sector[64] = b.DriveNumber

	// BS_BootSig
	sector[66] = 0x29

	// BS_VolID
	binary.LittleEndian.PutUint32(sector[67:71], b.VolumeID)

	// BS_VolLab
	if len(b.VolumeLabel) > 11 {
		return nil, errors.New("VolumeLabel must be 11 bytes or less")
	}

	for i, r := range b.VolumeLabel {
		if r > unicode.MaxASCII {
			return nil, fmt.Errorf("%#U in VolumeLabel not a valid ASCII char. Must be ASCII.", r)
		}

		sector[71+i] = byte(r)
	}

	// BS_FilSysType
	if len(b.FileSystemTypeLabel) > 8 {
		return nil, errors.New("FileSystemTypeLabel must be 8 bytes or less")
	}

	for i, r := range b.FileSystemTypeLabel {
		if r > unicode.MaxASCII {
			return nil, fmt.Errorf("%#U in FileSystemTypeLabel not a valid ASCII char. Must be ASCII.", r)
		}

		sector[82+i] = byte(r)
	}

	return sector, nil
}
