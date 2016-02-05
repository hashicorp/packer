package fat

import (
	"errors"
	"fmt"
	"github.com/mitchellh/go-fs"
	"math"
)

// The first cluster that can really hold user data is always 2
const FirstCluster = 2

// FAT is the actual file allocation table data structure that is
// stored on disk to describe the various clusters on the disk.
type FAT struct {
	bs      *BootSectorCommon
	entries []uint32
}

func DecodeFAT(device fs.BlockDevice, bs *BootSectorCommon, n int) (*FAT, error) {
	if n > int(bs.NumFATs) {
		return nil, fmt.Errorf("FAT #%d greater than total FATs: %d", n, bs.NumFATs)
	}

	data := make([]byte, bs.SectorsPerFat*uint32(bs.BytesPerSector))
	if _, err := device.ReadAt(data, int64(bs.FATOffset(n))); err != nil {
		return nil, err
	}

	result := &FAT{
		bs:      bs,
		entries: make([]uint32, FATEntryCount(bs)),
	}

	fatType := bs.FATType()
	for i := 0; i < int(FATEntryCount(bs)); i++ {
		var entryData uint32
		switch fatType {
		case FAT12:
			entryData = fatReadEntry12(data, i)
		case FAT16:
			entryData = fatReadEntry16(data, i)
		default:
			entryData = fatReadEntry32(data, i)
		}

		result.entries[i] = entryData
	}

	return result, nil
}

// NewFAT creates a new FAT data structure, properly initialized.
func NewFAT(bs *BootSectorCommon) (*FAT, error) {
	result := &FAT{
		bs:      bs,
		entries: make([]uint32, FATEntryCount(bs)),
	}

	// Set the initial two entries according to spec
	result.entries[0] = (uint32(bs.Media) & 0xFF) |
		(0xFFFFFF00 & result.entryMask())
	result.entries[1] = 0xFFFFFFFF & result.entryMask()

	return result, nil
}

// Bytes returns the raw bytes for the FAT that should be written to
// the block device.
func (f *FAT) Bytes() []byte {
	result := make([]byte, f.bs.SectorsPerFat*uint32(f.bs.BytesPerSector))

	for i, entry := range f.entries {
		switch f.bs.FATType() {
		case FAT12:
			f.writeEntry12(result, i, entry)
		case FAT16:
			f.writeEntry16(result, i, entry)
		default:
			f.writeEntry32(result, i, entry)
		}
	}

	return result
}

func (f *FAT) AllocChain() (uint32, error) {
	return f.allocNew()
}

func (f *FAT) allocNew() (uint32, error) {
	dataSize := (f.bs.TotalSectors * uint32(f.bs.BytesPerSector))
	dataSize -= f.bs.DataOffset()
	clusterCount := dataSize / f.bs.BytesPerCluster()
	lastClusterIndex := clusterCount + FirstCluster

	var availIdx uint32
	found := false
	for i := uint32(FirstCluster); i < lastClusterIndex; i++ {
		if f.entries[i] == 0 {
			availIdx = i
			found = true
			break
		}
	}

	if !found {
		return 0, errors.New("FAT FULL")
	}

	// Mark that this is now in use
	f.entries[availIdx] = 0xFFFFFFFF & f.entryMask()

	return availIdx, nil
}

// Chain returns the chain of clusters starting at a certain cluster.
func (f *FAT) Chain(start uint32) []uint32 {
	chain := make([]uint32, 0, 2)

	cluster := start
	for {
		chain = append(chain, cluster)
		cluster = f.entries[cluster]

		if f.isEofCluster(cluster) || cluster == 0 {
			break
		}
	}

	return chain
}

// ResizeChain takes a given cluster number and resizes the chain
// to the given length. It returns the new chain of clusters.
func (f *FAT) ResizeChain(start uint32, length int) ([]uint32, error) {
	chain := f.Chain(start)
	if len(chain) == length {
		return chain, nil
	}

	change := int(math.Abs(float64(length - len(chain))))
	if length > len(chain) {
		var lastCluster uint32

		lastCluster = chain[0]
		for i := 1; i < len(chain); i++ {
			if f.isEofCluster(f.entries[lastCluster]) {
				break
			}

			lastCluster = chain[i]
		}

		for i := 0; i < change; i++ {
			newCluster, err := f.allocNew()
			if err != nil {
				return nil, err
			}

			f.entries[lastCluster] = newCluster
			lastCluster = newCluster
		}
	} else {
		panic("making chains smaller not implemented yet")
	}

	return f.Chain(start), nil
}

func (f *FAT) WriteToDevice(device fs.BlockDevice) error {
	fatBytes := f.Bytes()
	for i := 0; i < int(f.bs.NumFATs); i++ {
		offset := int64(f.bs.FATOffset(i))
		if _, err := device.WriteAt(fatBytes, offset); err != nil {
			return err
		}
	}

	return nil
}

func (f *FAT) entryMask() uint32 {
	switch f.bs.FATType() {
	case FAT12:
		return 0x0FFF
	case FAT16:
		return 0xFFFF
	default:
		return 0x0FFFFFFF
	}
}

func (f *FAT) isEofCluster(cluster uint32) bool {
	return cluster >= (0xFFFFFF8 & f.entryMask())
}

func (f *FAT) writeEntry12(data []byte, idx int, entry uint32) {
	dataIdx := idx + (idx / 2)
	data = data[dataIdx : dataIdx+2]

	if idx%2 == 1 {
		// ODD
		data[0] |= byte((entry & 0x0F) << 4)
		data[1] = byte((entry >> 4) & 0xFF)
	} else {
		// Even
		data[0] = byte(entry & 0xFF)
		data[1] = byte((entry >> 8) & 0x0F)
	}
}

func (f *FAT) writeEntry16(data []byte, idx int, entry uint32) {
	idx <<= 1
	data[idx] = byte(entry & 0xFF)
	data[idx+1] = byte((entry >> 8) & 0xFF)
}

func (f *FAT) writeEntry32(data []byte, idx int, entry uint32) {
	idx <<= 2
	data[idx] = byte(entry & 0xFF)
	data[idx+1] = byte((entry >> 8) & 0xFF)
	data[idx+2] = byte((entry >> 16) & 0xFF)
	data[idx+3] = byte((entry >> 24) & 0xFF)
}

// FATEntryCount returns the number of entries per fat for the given
// boot sector.
func FATEntryCount(bs *BootSectorCommon) uint32 {
	// Determine the number of entries that'll go in the FAT.
	var entryCount uint32 = bs.SectorsPerFat * uint32(bs.BytesPerSector)
	switch bs.FATType() {
	case FAT12:
		entryCount = uint32((uint64(entryCount) * 8) / 12)
	case FAT16:
		entryCount /= 2
	case FAT32:
		entryCount /= 4
	default:
		panic("impossible fat type")
	}

	return entryCount
}

func fatReadEntry12(data []byte, idx int) uint32 {
	idx += idx / 2

	var result uint32 = (uint32(data[idx+1]) << 8) | uint32(data[idx])
	if idx%2 == 0 {
		return result & 0xFFF
	} else {
		return result >> 4
	}
}

func fatReadEntry16(data []byte, idx int) uint32 {
	idx <<= 1
	return (uint32(data[idx+1]) << 8) | uint32(data[idx])
}

func fatReadEntry32(data []byte, idx int) uint32 {
	idx <<= 2
	return (uint32(data[idx+3]) << 24) |
		(uint32(data[idx+2]) << 16) |
		(uint32(data[idx+1]) << 8) |
		uint32(data[idx+0])
}
