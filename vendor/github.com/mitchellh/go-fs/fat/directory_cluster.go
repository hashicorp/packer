package fat

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/mitchellh/go-fs"
	"math"
	"time"
	"unicode/utf16"
)

type DirectoryAttr uint8

const (
	AttrReadOnly  DirectoryAttr = 0x01
	AttrHidden                  = 0x02
	AttrSystem                  = 0x04
	AttrVolumeId                = 0x08
	AttrDirectory               = 0x10
	AttrArchive                 = 0x20
	AttrLongName                = AttrReadOnly | AttrHidden | AttrSystem | AttrVolumeId
)

// The size in bytes of a single directory entry.
const DirectoryEntrySize = 32

// Mask applied to the ord of the last long entry.
const LastLongEntryMask = 0x40

// DirectoryCluster represents a cluster on the disk that contains
// entries/contents.
type DirectoryCluster struct {
	entries      []*DirectoryClusterEntry
	fat16Root    bool
	startCluster uint32
}

// DirectoryClusterEntry is a single 32-byte entry that is part of the
// chain of entries in a directory cluster.
type DirectoryClusterEntry struct {
	name       string
	ext        string
	attr       DirectoryAttr
	createTime time.Time
	accessTime time.Time
	writeTime  time.Time
	cluster    uint32
	fileSize   uint32
	deleted    bool

	longOrd      uint8
	longName     string
	longChecksum uint8
}

func DecodeDirectoryCluster(startCluster uint32, device fs.BlockDevice, fat *FAT) (*DirectoryCluster, error) {
	bs := fat.bs
	chain := fat.Chain(startCluster)
	data := make([]byte, uint32(len(chain))*bs.BytesPerCluster())
	for i, clusterNumber := range chain {
		dataOffset := uint32(i) * bs.BytesPerCluster()
		devOffset := int64(bs.ClusterOffset(int(clusterNumber)))
		chainData := data[dataOffset : dataOffset+bs.BytesPerCluster()]

		if _, err := device.ReadAt(chainData, devOffset); err != nil {
			return nil, err
		}
	}

	result, err := decodeDirectoryCluster(data, bs)
	if err != nil {
		return nil, err
	}

	result.startCluster = startCluster
	return result, nil
}

// DecodeFAT16RootDirectory decodes the FAT16 root directory structure
// from the device.
func DecodeFAT16RootDirectoryCluster(device fs.BlockDevice, bs *BootSectorCommon) (*DirectoryCluster, error) {
	data := make([]byte, DirectoryEntrySize*bs.RootEntryCount)
	if _, err := device.ReadAt(data, int64(bs.RootDirOffset())); err != nil {
		return nil, err
	}

	result, err := decodeDirectoryCluster(data, bs)
	if err != nil {
		return nil, err
	}

	result.fat16Root = true
	return result, nil
}

func decodeDirectoryCluster(data []byte, bs *BootSectorCommon) (*DirectoryCluster, error) {
	entries := make([]*DirectoryClusterEntry, 0, bs.RootEntryCount)
	for i := uint16(0); i < uint16(len(data)/DirectoryEntrySize); i++ {
		offset := i * DirectoryEntrySize
		entryData := data[offset : offset+DirectoryEntrySize]
		if entryData[0] == 0 {
			break
		}

		entry, err := DecodeDirectoryClusterEntry(entryData)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}

	result := &DirectoryCluster{
		entries: entries,
	}

	return result, nil
}

func NewDirectoryCluster(start uint32, parent uint32, t time.Time) *DirectoryCluster {
	cluster := new(DirectoryCluster)
	cluster.startCluster = start

	// Create the "." and ".." entries
	cluster.entries = []*DirectoryClusterEntry{
		&DirectoryClusterEntry{
			accessTime: t,
			attr:       AttrDirectory,
			cluster:    start,
			createTime: t,
			name:       ".",
			writeTime:  t,
		},
		&DirectoryClusterEntry{
			accessTime: t,
			attr:       AttrDirectory,
			cluster:    parent,
			createTime: t,
			name:       "..",
			writeTime:  t,
		},
	}

	return cluster
}

// NewFat16RootDirectory creates a new DirectoryCluster that is meant only
// to be the root directory of a FAT12/FAT16 filesystem.
func NewFat16RootDirectoryCluster(bs *BootSectorCommon, label string) (*DirectoryCluster, error) {
	if bs.RootEntryCount == 0 {
		return nil, errors.New("root entry count is 0 in boot sector")
	}

	result := &DirectoryCluster{
		entries: make([]*DirectoryClusterEntry, 1, bs.RootEntryCount),
	}

	// Create the volume ID entry
	result.entries[0] = &DirectoryClusterEntry{
		attr: AttrVolumeId,
		name: label,
		cluster: 0,
	}

	return result, nil
}

// Bytes returns the on-disk byte data for this directory structure.
func (d *DirectoryCluster) Bytes() []byte {
	result := make([]byte, cap(d.entries)*DirectoryEntrySize)

	for i, entry := range d.entries {
		offset := i * DirectoryEntrySize
		entryBytes := entry.Bytes()
		copy(result[offset:offset+DirectoryEntrySize], entryBytes)
	}

	return result
}

// WriteToDevice writes the cluster to the device.
func (d *DirectoryCluster) WriteToDevice(device fs.BlockDevice, fat *FAT) error {
	if d.fat16Root {
		// Write the cluster to the FAT16 root directory location
		offset := int64(fat.bs.RootDirOffset())
		if _, err := device.WriteAt(d.Bytes(), offset); err != nil {
			return err
		}
	} else {
		chain := &ClusterChain{
			device:       device,
			fat:          fat,
			startCluster: d.startCluster,
		}

		if _, err := chain.Write(d.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

// Bytes returns the on-disk byte data for this directory entry.
func (d *DirectoryClusterEntry) Bytes() []byte {
	var result [DirectoryEntrySize]byte

	if d.longName != "" {
		runes := bytes.Runes([]byte(d.longName))

		// The name must be zero-terminated then padded with 0xFF
		// up to 13 characters
		if len(runes) < 13 {
			runes = append(runes, 0)
			for len(runes) < 13 {
				runes = append(runes, 0xFFFF)
			}
		}

		// LDIR_Ord
		result[0] = d.longOrd

		// LDIR_Name1
		for i := 0; i < int(math.Min(float64(len(runes)), 5)); i++ {
			offset := 1 + (i * 2)
			data := result[offset : offset+2]
			binary.LittleEndian.PutUint16(data, uint16(runes[i]))
		}

		// LDIR_Attr
		result[11] = byte(AttrLongName)

		// LDIR_Type
		result[12] = 0

		// LDIR_Chksum
		result[13] = d.longChecksum

		// LDIR_Name2
		for i := 0; i < 6; i++ {
			offset := 14 + (i * 2)
			data := result[offset : offset+2]
			binary.LittleEndian.PutUint16(data, uint16(runes[i+5]))
		}

		// LDIR_FstClusLO
		result[26] = 0
		result[27] = 0

		// LDIR_Name3
		for i := 0; i < 2; i++ {
			offset := 28 + (i * 2)
			data := result[offset : offset+2]
			binary.LittleEndian.PutUint16(data, uint16(runes[i+11]))
		}
	} else {
		// DIR_Name
		var simpleName string
		if d.name == "." || d.name == ".." {
			simpleName = d.name
		} else {
			simpleName = fmt.Sprintf("%s.%s", d.name, d.ext)
		}
		copy(result[0:11], shortNameEntryValue(simpleName))

		// DIR_Attr
		result[11] = byte(d.attr)

		// DIR_CrtTime
		crtDate, crtTime, crtTenths := encodeDOSTime(d.createTime)
		result[13] = crtTenths
		binary.LittleEndian.PutUint16(result[14:16], crtTime)
		binary.LittleEndian.PutUint16(result[16:18], crtDate)

		// DIR_LstAccDate
		accDate, _, _ := encodeDOSTime(d.accessTime)
		binary.LittleEndian.PutUint16(result[18:20], accDate)

		// DIR_FstClusHI
		binary.LittleEndian.PutUint16(result[20:22], uint16(d.cluster>>16))

		// DIR_WrtTime and DIR_WrtDate
		wrtDate, wrtTime, _ := encodeDOSTime(d.writeTime)
		binary.LittleEndian.PutUint16(result[22:24], wrtTime)
		binary.LittleEndian.PutUint16(result[24:26], wrtDate)

		// DIR_FstClusLO
		binary.LittleEndian.PutUint16(result[26:28], uint16(d.cluster&0xFFFF))

		// DIR_FileSize
		binary.LittleEndian.PutUint32(result[28:32], d.fileSize)
	}

	return result[:]
}

// IsLong returns true if this is a long entry.
func (d *DirectoryClusterEntry) IsLong() bool {
	return (d.attr & AttrLongName) == AttrLongName
}

func (d *DirectoryClusterEntry) IsVolumeId() bool {
	return (d.attr & AttrVolumeId) == AttrVolumeId
}

// DecodeDirectoryClusterEntry decodes a single directory entry in the
// Directory structure.
func DecodeDirectoryClusterEntry(data []byte) (*DirectoryClusterEntry, error) {
	var result DirectoryClusterEntry

	// Do the attributes so we can determine if we're dealing with long names
	result.attr = DirectoryAttr(data[11])
	if (result.attr & AttrLongName) == AttrLongName {
		result.longOrd = data[0]

		chars := make([]uint16, 13)
		for i := 0; i < 5; i++ {
			offset := 1 + (i * 2)
			chars[i] = binary.LittleEndian.Uint16(data[offset : offset+2])
		}

		for i := 0; i < 6; i++ {
			offset := 14 + (i * 2)
			chars[i+5] = binary.LittleEndian.Uint16(data[offset : offset+2])
		}

		for i := 0; i < 2; i++ {
			offset := 28 + (i * 2)
			chars[i+11] = binary.LittleEndian.Uint16(data[offset : offset+2])
		}

		result.longName = string(utf16.Decode(chars))
		result.longChecksum = data[13]
	} else {
		result.deleted = data[0] == 0xE5

		// Basic attributes
		if data[0] == 0x05 {
			data[0] = 0xE5
		}

		result.name = string(data[0:8])
		result.ext = string(data[8:11])

		// Creation time
		createTimeTenths := data[13]
		createTimeWord := binary.LittleEndian.Uint16(data[14:16])
		createDateWord := binary.LittleEndian.Uint16(data[16:18])
		result.createTime = decodeDOSTime(createDateWord, createTimeWord, createTimeTenths)

		// Access time
		accessDateWord := binary.LittleEndian.Uint16(data[18:20])
		result.accessTime = decodeDOSTime(accessDateWord, 0, 0)

		// Write time
		writeTimeWord := binary.LittleEndian.Uint16(data[22:24])
		writeDateWord := binary.LittleEndian.Uint16(data[24:26])
		result.writeTime = decodeDOSTime(writeDateWord, writeTimeWord, 0)

		// Cluster
		result.cluster = uint32(binary.LittleEndian.Uint16(data[20:22]))
		result.cluster <<= 4
		result.cluster |= uint32(binary.LittleEndian.Uint16(data[26:28]))

		// File size
		result.fileSize = binary.LittleEndian.Uint32(data[28:32])
	}

	return &result, nil
}

// NewLongDirectoryClusterEntry returns the series of directory cluster
// entries that need to be written for a long directory entry. This list
// of entries does NOT contain the short name entry.
func NewLongDirectoryClusterEntry(name string, shortName string) ([]*DirectoryClusterEntry, error) {
	// Split up the shortName properly
	checksum := checksumShortName(shortNameEntryValue(shortName))

	// Calcualte the number of entries we'll actually need to store
	// the long name.
	numLongEntries := len(name) / 13
	if len(name)%13 != 0 {
		numLongEntries++
	}

	entries := make([]*DirectoryClusterEntry, numLongEntries)
	for i := 0; i < numLongEntries; i++ {
		entries[i] = new(DirectoryClusterEntry)
		entry := entries[i]
		entry.attr = AttrLongName
		entry.longOrd = uint8(numLongEntries - i)

		if i == 0 {
			entry.longOrd |= LastLongEntryMask
		}

		// Calculate the offsets of the string for this entry
		j := (numLongEntries - i - 1) * 13
		k := j + 13
		if k > len(name) {
			k = len(name)
		}

		entry.longChecksum = checksum
		entry.longName = name[j:k]
	}

	return entries, nil
}

func decodeDOSTime(date, dosTime uint16, tenths uint8) time.Time {
	return time.Date(
		1980+int(date>>9),
		time.Month((date>>5)&0x0F),
		int(date&0x1F),
		int(dosTime>>11),
		int((dosTime>>5)&0x3F),
		int((dosTime&0x1F)*2),
		int(tenths)*10*int(time.Millisecond),
		time.Local)
}

func encodeDOSTime(t time.Time) (uint16, uint16, uint8) {
	var date uint16 = uint16((t.Year() - 1980) << 9)
	date |= uint16(t.Month()) << 5
	date += uint16(t.Day() & 0xFF)

	var time uint16 = uint16(t.Hour() << 11)
	time |= uint16(t.Minute() << 5)
	time += uint16(t.Second() / 2)

	var tenths uint8
	// TODO(mitchellh): Do tenths

	return date, time, tenths
}
