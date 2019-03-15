// Package shell defines code that is common in shells
package shell

import "github.com/hashicorp/packer/common"

// Provisioner contains common fields to all shell provisioners
type Provisioner struct {
	common.PackerConfig `mapstructure:",squash"`

	// If true, the script contains binary and line endings will not be
	// converted from Windows to Unix-style.
	Binary bool

	// The command used to execute the script. The '{{ .Path }}' variable
	// should be used to specify where the script goes, {{ .Vars }}
	// can be used to inject the environment_vars into the environment.
	ExecuteCommand string `mapstructure:"execute_command"`

	// An inline script to execute. Multiple strings are all executed
	// in the context of a single shell.
	Inline []string

	// The remote path where the local shell script will be uploaded to.
	// This should be set to a writable file that is in a pre-existing directory.
	// This defaults to remote_folder/remote_file
	RemotePath string `mapstructure:"remote_path"`

	// The local path of the shell script to upload and execute.
	Script string

	// An array of multiple scripts to run.
	Scripts []string

	// Valid Exit Codes - 0 is not always the only valid error code!  See
	// http://www.symantec.com/connect/articles/windows-system-error-codes-exit-codes-description
	// for examples such as 3010 - "The requested operation is successful.
	ValidExitCodes []int `mapstructure:"valid_exit_codes"`

	// An array of environment variables that will be injected before
	// your command(s) are executed.
	Vars []string `mapstructure:"environment_vars"`
}
