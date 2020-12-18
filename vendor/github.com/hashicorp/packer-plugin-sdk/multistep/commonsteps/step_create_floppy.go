package commonsteps

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
	"github.com/mitchellh/go-fs"
	"github.com/mitchellh/go-fs/fat"
)

// StepCreateFloppy will create a floppy disk with the given files.
type StepCreateFloppy struct {
	Files       []string
	Directories []string
	Label       string

	floppyPath string

	FilesAdded map[string]bool
}

func (s *StepCreateFloppy) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if len(s.Files) == 0 && len(s.Directories) == 0 {
		log.Println("No floppy files specified. Floppy disk will not be made.")
		return multistep.ActionContinue
	}

	if s.Label == "" {
		s.Label = "packer"
	} else {
		log.Printf("Floppy label is set to %s", s.Label)
	}

	s.FilesAdded = make(map[string]bool)

	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Creating floppy disk...")

	// Create a temporary file to be our floppy drive
	floppyF, err := tmp.File("packer")
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
		Label:   s.Label,
		OEMName: s.Label,
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

	// Get the root directory to the filesystem and create a cache for any directories within
	log.Println("Reading the root directory from the filesystem")
	rootDir, err := fatFs.RootDir()
	if err != nil {
		state.Put("error", fmt.Errorf("Error creating floppy: %s", err))
		return multistep.ActionHalt
	}
	cache := fsDirectoryCache(rootDir)

	// Utility functions for walking through a directory grabbing all files flatly
	globFiles := func(files []string, list chan string) {
		for _, filename := range files {
			if strings.ContainsAny(filename, "*?[") {
				matches, _ := filepath.Glob(filename)
				if err != nil {
					continue
				}

				for _, match := range matches {
					list <- match
				}
				continue
			}
			list <- filename
		}
		close(list)
	}

	var crawlDirectoryFiles []string
	crawlDirectory := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			crawlDirectoryFiles = append(crawlDirectoryFiles, path)
			ui.Message(fmt.Sprintf("Adding file: %s", path))
		}
		return nil
	}
	crawlDirectoryFiles = []string{}

	// Collect files and copy them flatly...because floppy_files is broken on purpose.
	var filelist chan string
	filelist = make(chan string)
	go globFiles(s.Files, filelist)

	ui.Message("Copying files flatly from floppy_files")
	for {
		filename, ok := <-filelist
		if !ok {
			break
		}

		finfo, err := os.Stat(filename)
		if err != nil {
			state.Put("error", fmt.Errorf("Error trying to stat : %s : %s", filename, err))
			return multistep.ActionHalt
		}

		// walk through directory adding files to the root of the fs
		if finfo.IsDir() {
			ui.Message(fmt.Sprintf("Copying directory: %s", filename))

			err := filepath.Walk(filename, crawlDirectory)
			if err != nil {
				state.Put("error", fmt.Errorf("Error adding file from floppy_files : %s : %s", filename, err))
				return multistep.ActionHalt
			}

			for _, crawlfilename := range crawlDirectoryFiles {
				if err = s.Add(cache, crawlfilename); err != nil {
					state.Put("error", fmt.Errorf("Error adding file from floppy_files : %s : %s", filename, err))
					return multistep.ActionHalt
				}
				s.FilesAdded[crawlfilename] = true
			}

			crawlDirectoryFiles = []string{}
			continue
		}

		// add just a single file
		ui.Message(fmt.Sprintf("Copying file: %s", filename))
		if err = s.Add(cache, filename); err != nil {
			state.Put("error", fmt.Errorf("Error adding file from floppy_files : %s : %s", filename, err))
			return multistep.ActionHalt
		}
		s.FilesAdded[filename] = true
	}
	ui.Message("Done copying files from floppy_files")

	// Collect all paths (expanding wildcards) into pathqueue
	ui.Message("Collecting paths from floppy_dirs")
	var pathqueue []string
	for _, filename := range s.Directories {
		if strings.ContainsAny(filename, "*?[") {
			matches, err := filepath.Glob(filename)
			if err != nil {
				state.Put("error", fmt.Errorf("Error adding path %s to floppy: %s", filename, err))
				return multistep.ActionHalt
			}

			for _, filename := range matches {
				pathqueue = append(pathqueue, filename)
			}
			continue
		}
		pathqueue = append(pathqueue, filename)
	}
	ui.Message(fmt.Sprintf("Resulting paths from floppy_dirs : %v", pathqueue))

	// Go over each path in pathqueue and copy it.
	for _, src := range pathqueue {
		ui.Message(fmt.Sprintf("Recursively copying : %s", src))
		err = s.Add(cache, src)
		if err != nil {
			state.Put("error", fmt.Errorf("Error adding path %s to floppy: %s", src, err))
			return multistep.ActionHalt
		}
	}
	ui.Message("Done copying paths from floppy_dirs")

	// Set the path to the floppy so it can be used later
	state.Put("floppy_path", s.floppyPath)

	return multistep.ActionContinue
}

func (s *StepCreateFloppy) Add(dircache directoryCache, src string) error {
	finfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("Error adding path to floppy: %s", err)
	}

	// add a file
	if !finfo.IsDir() {
		inputF, err := os.Open(src)
		if err != nil {
			return err
		}
		defer inputF.Close()

		d, err := dircache("")
		if err != nil {
			return err
		}

		entry, err := d.AddFile(path.Base(filepath.ToSlash(src)))
		if err != nil {
			return err
		}

		fatFile, err := entry.File()
		if err != nil {
			return err
		}

		_, err = io.Copy(fatFile, inputF)
		s.FilesAdded[src] = true
		return err
	}

	// add a directory and it's subdirectories
	basedirectory := filepath.Join(src, "..")
	visit := func(pathname string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.Mode().IsDir() {
			base, err := removeBase(basedirectory, pathname)
			if err != nil {
				return err
			}
			_, err = dircache(filepath.ToSlash(base))
			return err
		}
		directory, filename := filepath.Split(filepath.ToSlash(pathname))

		base, err := removeBase(basedirectory, filepath.FromSlash(directory))
		if err != nil {
			return err
		}

		inputF, err := os.Open(pathname)
		if err != nil {
			return err
		}
		defer inputF.Close()

		wd, err := dircache(filepath.ToSlash(base))
		if err != nil {
			return err
		}

		entry, err := wd.AddFile(filename)
		if err != nil {
			return err
		}

		fatFile, err := entry.File()
		if err != nil {
			return err
		}

		_, err = io.Copy(fatFile, inputF)
		s.FilesAdded[pathname] = true
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
func removeBase(base string, path string) (string, error) {
	var idx int
	var err error

	if res, err := filepath.Abs(path); err == nil {
		path = res
	}
	path = filepath.Clean(path)

	if base, err = filepath.Abs(base); err != nil {
		return path, err
	}

	c1, c2 := strings.Split(base, string(os.PathSeparator)), strings.Split(path, string(os.PathSeparator))
	for idx = 0; idx < len(c1); idx++ {
		if len(c1[idx]) == 0 && len(c2[idx]) != 0 {
			break
		}
		if c1[idx] != c2[idx] {
			return "", fmt.Errorf("Path %s is not prefixed by Base %s", path, base)
		}
	}
	return strings.Join(c2[idx:], string(os.PathSeparator)), nil
}

// fsDirectoryCache returns a function that can be used to grab the fs.Directory
// entry associated with a given path. If an fs.Directory entry is not found
// then it will be created relative to the rootDirectory argument that is
// passed.
type directoryCache func(string) (fs.Directory, error)

func fsDirectoryCache(rootDirectory fs.Directory) directoryCache {
	var cache map[string]fs.Directory

	cache = make(map[string]fs.Directory)
	cache[""] = rootDirectory

	Input, Output, Error := make(chan string), make(chan fs.Directory), make(chan error)
	go func(Error chan error) {
		for {
			input := <-Input
			if len(input) > 0 {
				input = path.Clean(input)
			}

			// found a directory, so yield it
			res, ok := cache[input]
			if ok {
				Output <- res
				continue
			}
			component := strings.Split(input, "/")

			// directory not cached, so start at the root and walk each component
			// creating them if they're not in cache
			var entry fs.Directory
			for i := range component {

				// join all of our components into a key
				path := strings.Join(component[:i], "/")

				// check if parent directory is cached
				res, ok = cache[path]
				if !ok {
					// add directory into cache
					directory, err := entry.AddDirectory(component[i-1])
					if err != nil {
						Error <- err
						continue
					}
					res, err = directory.Dir()
					if err != nil {
						Error <- err
						continue
					}
					cache[path] = res
				}
				// cool, found a directory
				entry = res
			}

			// finally create our directory
			directory, err := entry.AddDirectory(component[len(component)-1])
			if err != nil {
				Error <- err
				continue
			}
			res, err = directory.Dir()
			if err != nil {
				Error <- err
				continue
			}
			cache[input] = res

			// ..and yield it
			Output <- entry
		}
	}(Error)

	getFilesystemDirectory := func(input string) (fs.Directory, error) {
		Input <- input
		select {
		case res := <-Output:
			return res, nil
		case err := <-Error:
			return *new(fs.Directory), err
		}
	}
	return getFilesystemDirectory
}
