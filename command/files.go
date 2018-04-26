package command

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mitchellh/packer/template"
)

type FilesCommand struct {
	Meta
}

func (c *FilesCommand) Run(args []string) int {
	flags := c.Meta.FlagSet("files", FlagSetNone)
	flags.Usage = func() { c.Ui.Say(c.Help()) }
	if err := flags.Parse(args); err != nil {
		return 1
	}

	args = flags.Args()
	if len(args) != 1 {
		flags.Usage()
		return 1
	}

	// Parse the template
	tpl, err := template.ParseFile(args[0])
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to parse template: %s", err))
		return 1
	}

	cwd := filepath.Dir(args[0])

	// Convenience...
	ui := c.Ui
	files := []string{tpl.Path}

	// Builders
	for _, v := range tpl.Builders {
		switch v.Type {
		case "parallels-iso", "parallels-pvm", "vmware-iso", "virtualbox-ovf", "virtualbox-iso", "vmware-vmx", "qemu":
			if hd, ok := v.Config["http_directory"]; ok {
				if globs, err := filepath.Glob(filepath.Join(cwd, hd.(string), "/*")); err == nil {
					files = append(files, globs...)
				}
			}
			if fp, ok := v.Config["floppy_files"]; ok {
				switch fp.(type) {
				case []string:
					for _, s := range fp.([]string) {
						if globs, err := filepath.Glob(filepath.Join(cwd, s)); err == nil {
							files = append(files, globs...)
						}
					}
				}
			}
			if fd, ok := v.Config["floppy_dirs"]; ok {
				switch fd.(type) {
				case []string:
					for _, s := range fd.([]string) {
						if globs, err := filepath.Glob(filepath.Join(cwd, s)); err == nil {
							files = append(files, globs...)
						}
					}
				}
			}
		}
	}

	// Provisioners
	for _, v := range tpl.Provisioners {
		switch v.Type {
		case "file":
			if source, ok := v.Config["source"]; ok {
				if direction, ok := v.Config["direction"]; ok {
					if direction.(string) == "upload" {
						files = append(files, filepath.Join(cwd, source.(string)))
					}
				} else {
					files = append(files, filepath.Join(cwd, source.(string)))
				}
			}
		case "shell", "windows-shell", "shell-local", "powershell":
			if script, ok := v.Config["script"]; ok {
				files = append(files, filepath.Join(cwd, script.(string)))
			}
			if scripts, ok := v.Config["scripts"]; ok {
				for _, s := range scripts.([]interface{}) {
					files = append(files, filepath.Join(cwd, s.(string)))
				}
			}
		}
	}

	// Post-processors
	for _, tv := range tpl.PostProcessors {
		for _, v := range tv {
			switch v.Type {
			case "shell-local":
				if script, ok := v.Config["script"]; ok {
					files = append(files, filepath.Join(cwd, script.(string)))
				}
				if scripts, ok := v.Config["scripts"]; ok {
					for _, s := range scripts.([]interface{}) {
						files = append(files, filepath.Join(cwd, s.(string)))
					}
				}
			}
		}
	}

	for _, s := range files {
		if rel, err := filepath.Rel(cwd, s); err == nil {
			ui.Say(fmt.Sprintf("%s", rel))
		}
	}

	return 0
}

func (*FilesCommand) Help() string {
	helpText := `
Usage: packer files TEMPLATE

  Files inspects a template, parsing and outputting the needed files to run
  template.

Options:

  -machine-readable  Machine-readable output
`

	return strings.TrimSpace(helpText)
}

func (c *FilesCommand) Synopsis() string {
	return "see components of a template"
}
