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
	"path"
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

	// Collect all paths (expanding wildcards) into pathqueue
	var pathqueue []string
	for _,filename := range s.Files {
		if strings.IndexAny(filename, "*?[") >= 0 {
			matches,err := filepath.Glob(filename)
			if err != nil {
				state.Put("error", fmt.Errorf("Error adding path %s to floppy: %s", filename, err))
				return multistep.ActionHalt
			}

			for _,filename := range matches {
				pathqueue = append(pathqueue, filename)
			}
			continue
		}
		pathqueue = append(pathqueue, filename)
	}

	// Go over each path in pathqueue and copy it.
	getDirectory := fsDirectoryCache(rootDir)
	for _,src := range pathqueue {
		ui.Message(fmt.Sprintf("Copying: %s", src))
		err = s.Add(getDirectory, src)
		if err != nil {
			state.Put("error", fmt.Errorf("Error adding path %s to floppy: %s", src, err))
			return multistep.ActionHalt
		}

		// FIXME: setting this map according to each pathqueue entry breaks
		// our testcases, because it only keeps track of the number of files
		// that are set here instead of actually verifying against the
		// filesystem...heh
//		s.FilesAdded[src] = true
	}

	// Set the path to the floppy so it can be used later
	state.Put("floppy_path", s.floppyPath)

	return multistep.ActionContinue
}

func (s *StepCreateFloppy) Add(dir getFsDirectory, src string) error {
	finfo,err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("Error adding path to floppy: %s", err)
	}

	// add a file
	if !finfo.IsDir() {
		inputF, err := os.Open(src)
		if err != nil { return err }
		defer inputF.Close()

		d,err := dir("")
		if err != nil { return err }

		entry,err := d.AddFile(path.Base(src))
		if err != nil { return err }

		fatFile,err := entry.File()
		if err != nil { return err }

		_,err = io.Copy(fatFile,inputF)
		s.FilesAdded[src] = true
		return err
	}

	// add a directory and it's subdirectories
	basedirectory := filepath.Join(src, "..")
	visit := func(pathname string, fi os.FileInfo, err error) error {
		if err != nil { return err }
		if fi.Mode().IsDir() {
			base,err := removeBase(basedirectory, pathname)
			if err != nil { return err }
			_,err = dir(filepath.ToSlash(base))
			return err
		}
		directory,filename := filepath.Split(pathname)

		base,err := removeBase(basedirectory, directory)
		if err != nil { return err }

		inputF, err := os.Open(pathname)
		if err != nil { return err }
		defer inputF.Close()

		wd,err := dir(filepath.ToSlash(base))
		if err != nil { return err }

		entry,err := wd.AddFile(filename)
		if err != nil { return err }

		fatFile,err := entry.File()
		if err != nil { return err }

		_,err = io.Copy(fatFile,inputF)
		s.FilesAdded[filename] = true
		return err
	}

	return filepath.Walk(src, visit)
}

func (s *StepCreateFloppy) Cleanup(multistep.StateBag) {
	if s.floppyPath != "" {
		log.Printf("Deleting floppy disk: %s", s.floppyPath)
		os.Remove(s.floppyPath)
	}
}

// removeBase will take a regular os.PathSeparator-separated path and remove the
// prefix directory base from it. Both paths are converted to their absolute
// formats before the stripping takes place.
func removeBase(base string, path string) (string,error) {
	var idx int
	var err error

	if res,err := filepath.Abs(path); err == nil {
		path = res
	}
	path = filepath.Clean(path)

	if base,err = filepath.Abs(base); err != nil {
		return path,err
	}

	c1,c2 := strings.Split(base, string(os.PathSeparator)), strings.Split(path, string(os.PathSeparator))
	for idx = 0; idx < len(c1); idx++ {
		if len(c1[idx]) == 0 && len(c2[idx]) != 0 { break }
		if c1[idx] != c2[idx] {
			return "", fmt.Errorf("Path %s is not prefixed by Base %s", path, base)
		}
	}
	return strings.Join(c2[idx:], string(os.PathSeparator)),nil
}

// fsDirectoryCache returns a function that can be used to grab the fs.Directory
// entry associated with a given path. If an fs.Directory entry is not found
// then it will be created relative to the rootDirectory argument that is
// passed.
type getFsDirectory func(string) (fs.Directory,error)
func fsDirectoryCache(rootDirectory fs.Directory) getFsDirectory {
	var cache map[string]fs.Directory

	cache = make(map[string]fs.Directory)
	cache[""] = rootDirectory

	Input,Output,Error := make(chan string),make(chan fs.Directory),make(chan error)
	go func(Error chan error) {
		for {
			input := path.Clean(<-Input)

			// found a directory, so yield it
			res,ok := cache[input]
			if ok {
				Output <- res
				continue
			}
			component := strings.Split(input, "/")

			// directory not cached, so start at the root and walk each component
			// creating them if they're not in cache
			var entry fs.Directory
			for i,_ := range component {

				// join all of our components into a key
				path := strings.Join(component[:i], "/")

				// check if parent directory is cached
				res,ok = cache[path]
				if !ok {
					// add directory into cache
					directory,err := entry.AddDirectory(component[i-1])
					if err != nil { Error <- err; continue }
					res,err = directory.Dir()
					if err != nil { Error <- err; continue }
					cache[path] = res
				}
				// cool, found a directory
				entry = res
			}

			// finally create our directory
			directory,err := entry.AddDirectory(component[len(component)-1])
			if err != nil { Error <- err; continue }
			res,err = directory.Dir()
			if err != nil { Error <- err; continue }
			cache[input] = res

			// ..and yield it
			Output <- entry
		}
	}(Error)

	getFilesystemDirectory := func(input string) (fs.Directory,error) {
		Input <- input
		select {
			case res := <-Output:
				return res,nil
			case err := <-Error:
				return *new(fs.Directory),err
		}
	}
	return getFilesystemDirectory
}
