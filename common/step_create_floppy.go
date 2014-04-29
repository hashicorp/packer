package common

import (
	"fmt"
	"github.com/mitchellh/go-fs"
	"github.com/mitchellh/go-fs/fat"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// StepCreateFloppy will create a floppy disk with the given files.
// The floppy disk doesn't support sub-directories. Only files at the
// root level are supported.
type StepCreateFloppy struct {
	Files []string

	floppyPath string

	FilesAdded map[string]bool
}

func (s *StepCreateFloppy) Run(state multistep.StateBag) multistep.StepAction {
	if len(s.Files) == 0 {
		log.Println("No floppy files specified. Floppy disk will not be made.")
		return multistep.ActionContinue
	}

	s.FilesAdded = make(map[string]bool)

	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating floppy disk...")

	// Create a temporary file to be our floppy drive
	floppyF, err := ioutil.TempFile("", "packer")
	if err != nil {
		state.Put("error",
			fmt.Errorf("Error creating temporary file for floppy: %s", err))
		return multistep.ActionHalt
	}
	defer floppyF.Close()

	// Set the path so we can remove it later
	s.floppyPath = floppyF.Name()

	log.Printf("Floppy path: %s", s.floppyPath)

	// Set the size of the file to be a floppy sized
	if err := floppyF.Truncate(1440 * 1024); err != nil {
		state.Put("error", fmt.Errorf("Error creating floppy: %s", err))
		return multistep.ActionHalt
	}

	// BlockDevice backed by the file for our filesystem
	log.Println("Initializing block device backed by temporary file")
	device, err := fs.NewFileDisk(floppyF)
	if err != nil {
		state.Put("error", fmt.Errorf("Error creating floppy: %s", err))
		return multistep.ActionHalt
	}

	// Format the block device so it contains a valid FAT filesystem
	log.Println("Formatting the block device with a FAT filesystem...")
	formatConfig := &fat.SuperFloppyConfig{
		FATType: fat.FAT12,
		Label:   "packer",
		OEMName: "packer",
	}
	if err := fat.FormatSuperFloppy(device, formatConfig); err != nil {
		state.Put("error", fmt.Errorf("Error creating floppy: %s", err))
		return multistep.ActionHalt
	}

	// The actual FAT filesystem
	log.Println("Initializing FAT filesystem on block device")
	fatFs, err := fat.New(device)
	if err != nil {
		state.Put("error", fmt.Errorf("Error creating floppy: %s", err))
		return multistep.ActionHalt
	}

	// Get the root directory to the filesystem
	log.Println("Reading the root directory from the filesystem")
	rootDir, err := fatFs.RootDir()
	if err != nil {
		state.Put("error", fmt.Errorf("Error creating floppy: %s", err))
		return multistep.ActionHalt
	}

	// Go over each file and copy it.
	for _, filename := range s.Files {
		ui.Message(fmt.Sprintf("Copying: %s", filename))
		if err := s.addFilespec(rootDir, filename); err != nil {
			state.Put("error", fmt.Errorf("Error adding file to floppy: %s", err))
			return multistep.ActionHalt
		}
	}

	// Set the path to the floppy so it can be used later
	state.Put("floppy_path", s.floppyPath)

	return multistep.ActionContinue
}

func (s *StepCreateFloppy) Cleanup(multistep.StateBag) {
	if s.floppyPath != "" {
		log.Printf("Deleting floppy disk: %s", s.floppyPath)
		os.Remove(s.floppyPath)
	}
}

func (s *StepCreateFloppy) addFilespec(dir fs.Directory, src string) error {
	// same as http://golang.org/src/pkg/path/filepath/match.go#L308
	if strings.IndexAny(src, "*?[") >= 0 {
		matches, err := filepath.Glob(src)
		if err != nil {
			return err
		}
		return s.addFiles(dir, matches)
	}

	finfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if finfo.IsDir() {
		return s.addDirectory(dir, src)
	}

	return s.addSingleFile(dir, src)
}

func (s *StepCreateFloppy) addFiles(dir fs.Directory, files []string) error {
	for _, file := range files {
		err := s.addFilespec(dir, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *StepCreateFloppy) addDirectory(dir fs.Directory, src string) error {
	log.Printf("Adding directory to floppy: %s", src)

	walkFn := func(path string, finfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == src {
			return nil
		}

		if finfo.IsDir() {
			return s.addDirectory(dir, path)
		}

		return s.addSingleFile(dir, path)
	}

	return filepath.Walk(src, walkFn)
}

func (s *StepCreateFloppy) addSingleFile(dir fs.Directory, src string) error {
	log.Printf("Adding file to floppy: %s", src)

	inputF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer inputF.Close()

	entry, err := dir.AddFile(filepath.Base(src))
	if err != nil {
		return err
	}

	fatFile, err := entry.File()
	if err != nil {
		return err
	}

	if _, err := io.Copy(fatFile, inputF); err != nil {
		return err
	}

	s.FilesAdded[src] = true

	return nil
}
