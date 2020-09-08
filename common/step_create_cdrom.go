package common

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/packer/helper/builder/localexec"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer/tmp"
)

// StepCreateCD will create a CD disk with the given files.
type StepCreateCD struct {
	// Files can be either files or directories. Any files provided here will
	// be written to the root of the CD. Directories will be written to the
	// root of the CD as well, but will retain their subdirectory structure.
	Files []string
	Label string

	CDPath string

	filesAdded map[string]bool
}

func (s *StepCreateCD) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if len(s.Files) == 0 {
		log.Println("No CD files specified. CD disk will not be made.")
		return multistep.ActionContinue
	}

	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating CD disk...")

	if s.Label == "" {
		s.Label = "packer"
	} else {
		log.Printf("CD label is set to %s", s.Label)
	}

	// Track what files are added. Used for testing step.
	s.filesAdded = make(map[string]bool)

	// Create a temporary file to be our CD drive
	CDF, err := tmp.File("packer*.iso")
	// Set the path so we can remove it later
	CDPath := CDF.Name()
	CDF.Close()
	os.Remove(CDPath)
	if err != nil {
		state.Put("error",
			fmt.Errorf("Error creating temporary file for CD: %s", err))
		return multistep.ActionHalt
	}

	log.Printf("CD path: %s", CDPath)
	s.CDPath = CDPath

	// Consolidate all files provided into a single directory to become our
	// "root" directory.
	rootFolder, err := tmp.Dir("packer_to_cdrom")
	if err != nil {
		state.Put("error",
			fmt.Errorf("Error creating temporary file for CD: %s", err))
		return multistep.ActionHalt
	}

	for _, toAdd := range s.Files {
		err = s.AddFile(rootFolder, toAdd)
		if err != nil {
			state.Put("error",
				fmt.Errorf("Error creating temporary file for CD: %s", err))
			return multistep.ActionHalt
		}
	}

	cmd := retrieveCommandForOS(s.Label, rootFolder, CDPath)
	err = localexec.RunAndStream(cmd, ui, []string{})
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message("Done copying paths from CD_dirs")

	// Set the path to the CD so it can be used later
	state.Put("cd_path", CDPath)

	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepCreateCD) Cleanup(multistep.StateBag) {
	if s.CDPath != "" {
		log.Printf("Deleting CD disk: %s", s.CDPath)
		os.Remove(s.CDPath)
	}
}

func retrieveCommandForOS(label string, source string, dest string) *exec.Cmd {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("oscdimg")
		args := []string{"-j1", "-o", "-m", "-l" + label, source, dest}
		cmd.Args = append(cmd.Args, args...)
		return cmd
	case "darwin":
		cmd := exec.Command("hdiutil")
		args := []string{"makehybrid", "-o", dest, "-hfs",
			"-joliet", "-iso", "-default-volume-name", label, source}
		cmd.Args = append(cmd.Args, args...)
		return cmd
	default:
		cmd := exec.Command("mkisofs")
		args := []string{"-joliet", "-volid", label, "-o", dest, source}
		cmd.Args = append(cmd.Args, args...)
		return cmd
	}
}

func (s *StepCreateCD) AddFile(dst, src string) error {
	finfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("Error adding path to CD: %s", err)
	}

	// add a file
	if !finfo.IsDir() {
		inputF, err := os.Open(src)
		if err != nil {
			return err
		}
		defer inputF.Close()

		// Create a new file in the root directory
		dest, err := os.Create(filepath.Join(dst, finfo.Name()))
		if err != nil {
			return fmt.Errorf("Error opening file for copy %s to CD root", src)
		}
		defer dest.Close()
		nBytes, err := io.Copy(dest, inputF)
		if err != nil {
			return fmt.Errorf("Error copying %s to CD root", src)
		}
		s.filesAdded[src] = true
		log.Printf("Wrote %d bytes to %s", nBytes, finfo.Name())
		return err
	}

	// Add a directory and its subdirectories
	visit := func(pathname string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// add a file
		if !fi.IsDir() {
			inputF, err := os.Open(pathname)
			if err != nil {
				return err
			}
			defer inputF.Close()

			fileDst, err := os.Create(filepath.Join(dst, pathname))
			if err != nil {
				return fmt.Errorf("Error opening file %s on CD", src)
			}
			defer fileDst.Close()
			nBytes, err := io.Copy(fileDst, inputF)
			if err != nil {
				return fmt.Errorf("Error copying %s to CD", src)
			}
			s.filesAdded[pathname] = true
			log.Printf("Wrote %d bytes to %s", nBytes, pathname)
			return err
		}

		if fi.Mode().IsDir() {
			// create the directory on the CD, continue walk.
			err := os.Mkdir(filepath.Join(dst, pathname), fi.Mode())
			if err != nil {
				err = fmt.Errorf("error creating new directory %s: %s",
					filepath.Join(dst, pathname), err)
			}
			return err
		}
		return err
	}

	return filepath.Walk(src, visit)
}
