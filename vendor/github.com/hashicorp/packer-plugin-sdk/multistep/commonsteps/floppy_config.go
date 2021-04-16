//go:generate struct-markdown

package commonsteps

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

// A floppy can be made available for your build. This is most useful for
// unattended Windows installs, which look for an Autounattend.xml file on
// removable media. By default, no floppy will be attached. All files listed in
// this setting get placed into the root directory of the floppy and the floppy
// is attached as the first floppy device. The summary size of the listed files
// must not exceed 1.44 MB. The supported ways to move large files into the OS
// are using `http_directory` or [the file
// provisioner](/docs/provisioners/file).
type FloppyConfig struct {
	// A list of files to place onto a floppy disk that is attached when the VM
	// is booted. Currently, no support exists for creating sub-directories on
	// the floppy. Wildcard characters (\\*, ?, and \[\]) are allowed. Directory
	// names are also allowed, which will add all the files found in the
	// directory to the floppy.
	FloppyFiles []string `mapstructure:"floppy_files"`
	// A list of directories to place onto the floppy disk recursively. This is
	// similar to the `floppy_files` option except that the directory structure
	// is preserved. This is useful for when your floppy disk includes drivers
	// or if you just want to organize it's contents as a hierarchy. Wildcard
	// characters (\\*, ?, and \[\]) are allowed. The maximum summary size of
	// all files in the listed directories are the same as in `floppy_files`.
	FloppyDirectories []string `mapstructure:"floppy_dirs"`
	FloppyLabel       string   `mapstructure:"floppy_label"`
}

func (c *FloppyConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	var err error

	if c.FloppyFiles == nil {
		c.FloppyFiles = make([]string, 0)
	}

	for _, path := range c.FloppyFiles {
		if strings.ContainsAny(path, "*?[") {
			_, err = filepath.Glob(path)
		} else {
			_, err = os.Stat(path)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("Bad Floppy disk file '%s': %s", path, err))
		}
	}

	if c.FloppyDirectories == nil {
		c.FloppyDirectories = make([]string, 0)
	}

	for _, path := range c.FloppyDirectories {
		if strings.ContainsAny(path, "*?[") {
			_, err = filepath.Glob(path)
		} else {
			_, err = os.Stat(path)
		}
		if err != nil {
			errs = append(errs, fmt.Errorf("Bad Floppy disk directory '%s': %s", path, err))
		}
	}

	return errs
}
