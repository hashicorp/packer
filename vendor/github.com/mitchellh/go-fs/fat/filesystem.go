package fat

import (
	"github.com/mitchellh/go-fs"
)

// FileSystem is the implementation of fs.FileSystem that can read a
// FAT filesystem.
type FileSystem struct {
	bs      *BootSectorCommon
	device  fs.BlockDevice
	fat     *FAT
	rootDir *DirectoryCluster
}

// New returns a new FileSystem for accessing a previously created
// FAT filesystem.
func New(device fs.BlockDevice) (*FileSystem, error) {
	bs, err := DecodeBootSector(device)
	if err != nil {
		return nil, err
	}

	fat, err := DecodeFAT(device, bs, 0)
	if err != nil {
		return nil, err
	}

	var rootDir *DirectoryCluster
	if bs.FATType() == FAT32 {
		panic("FAT32 not implemented yet")
	} else {
		rootDir, err = DecodeFAT16RootDirectoryCluster(device, bs)
		if err != nil {
			return nil, err
		}
	}

	result := &FileSystem{
		bs:      bs,
		device:  device,
		fat:     fat,
		rootDir: rootDir,
	}

	return result, nil
}

func (f *FileSystem) RootDir() (fs.Directory, error) {
	dir := &Directory{
		device:     f.device,
		dirCluster: f.rootDir,
		fat:        f.fat,
	}

	return dir, nil
}
