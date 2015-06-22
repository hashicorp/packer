// This package implements a provisioner for Packer that executes
// shell scripts within the remote machine.
package shelllocal

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	// "io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

const DefaultSheBang = "/bin/sh"

type config struct {
	common.PackerConfig `mapstructure:",squash"`

	// An inline script to execute. Multiple strings are all executed
	// in the context of a single shell.
	Inline []string

	// The shebang value used when running inline scripts.
	InlineShebang string `mapstructure:"inline_shebang"`

	tpl *packer.ConfigTemplate
}

type Provisioner struct {
	config config
}

type ExecuteCommandTemplate struct {
	Vars string
	Path string
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	md, err := common.DecodeConfig(&p.config, raws...)
	if err != nil {
		return err
	}

	p.config.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return err
	}
	p.config.tpl.UserVars = p.config.PackerUserVars

	// Accumulate any errors
	errs := common.CheckUnusedConfig(md)

	if p.config.Inline != nil && len(p.config.Inline) == 0 {
		p.config.Inline = nil
	}

	if p.config.InlineShebang == "" {
		p.config.InlineShebang = DefaultSheBang
	}

	templates := map[string]*string{
		"inline_shebang": &p.config.InlineShebang,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = p.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	sliceTemplates := map[string][]string{
		"inline": p.config.Inline,
	}

	for n, slice := range sliceTemplates {
		for i, elem := range slice {
			var err error
			slice[i], err = p.config.tpl.Process(elem, nil)
			if err != nil {
				errs = packer.MultiErrorAppend(
					errs, fmt.Errorf("Error processing %s[%d]: %s", n, i, err))
			}
		}
	}

	if p.config.Inline == nil {
		errs = packer.MultiErrorAppend(errs,
			errors.New("inline script must be specified."))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {

	scripts := make([]string, 0)

	if p.config.Inline != nil {
		tf, err := ioutil.TempFile("", "packer-shell-local")
		if err != nil {
			return fmt.Errorf("Error preparing shell script: %s", err)
		}
		defer os.Remove(tf.Name())

		// Set the path to the temporary file
		scripts = append(scripts, tf.Name())

		// Write our contents to it
		writer := bufio.NewWriter(tf)
		writer.WriteString(fmt.Sprintf("#!%s\n", p.config.InlineShebang))
		for _, command := range p.config.Inline {
			if _, err := writer.WriteString(command + "\n"); err != nil {
				return fmt.Errorf("Error preparing shell script: %s", err)
			}
		}

		if err := writer.Flush(); err != nil {
			return fmt.Errorf("Error preparing shell script: %s", err)
		}
		tf.Close()

	} else {
		return fmt.Errorf("Inline must be defined")
	}

	for _, path := range scripts {
		ui.Say(fmt.Sprintf("Provisioning with local shell script: %s", path))

		log.Printf("Executing local script %s", path)

		err := os.Chmod(path, 0770)
		if err != nil {
			return fmt.Errorf(
				"Error chmodding script file to 0777 "+
					"machine: %s", err)
		}

		out, err := exec.Command(path).Output()
		if err != nil {
			return fmt.Errorf("Error executeing script machine: %s", err)
		}
		ui.Say(fmt.Sprintf("%s", out))
	}

	return nil
}

func (p *Provisioner) Cancel() {
	// Just hard quit. It isn't a big deal if what we're doing keeps
	// running on the other side.
	os.Exit(0)
}
