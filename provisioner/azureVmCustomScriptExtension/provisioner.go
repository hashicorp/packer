// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package azureVmCustomScriptExtension

import (
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
	"os"
	"bytes"
	"log"
	"path/filepath"
	"bufio"
	"io/ioutil"
	"strings"
	"code.google.com/p/go-uuid/uuid"
)

const DistrDstPathDefault = "C:/PackerDistr"

type config struct {
	common.PackerConfig `mapstructure:",squash"`

	// The local path of the script.
	ScriptPath string `mapstructure:"script_path"`
	DistrSrcPath string `mapstructure:"distr_src_path"`
	Inline []string		`mapstructure:"inline"`
	tpl *packer.ConfigTemplate
}

type Provisioner struct {
	config config
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

	sliceTemplates := map[string][]string{
		"inline":           p.config.Inline,
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

	log.Println(fmt.Sprintf("%s: %v","inline", p.config.ScriptPath))

	templates := map[string]*string{
		"script_path":      &p.config.ScriptPath,
		"distr_src_path": 	&p.config.DistrSrcPath,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = p.config.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	if len(p.config.ScriptPath) == 0 && p.config.Inline == nil {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("Either a script file or inline script must be specified."))
	}

	if len(p.config.ScriptPath) != 0 {
		if _, err := os.Stat(p.config.ScriptPath); err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("script_path: '%v' check the path is correct.", p.config.ScriptPath))
		}
	}
	log.Println(fmt.Sprintf("%s: %v","script_path", p.config.ScriptPath))

	if len(p.config.DistrSrcPath) != 0 {
		if _, err := os.Stat(p.config.DistrSrcPath); err != nil {
			errs = packer.MultiErrorAppend(errs,
				fmt.Errorf("distr_src_path: '%v' check the path is correct.", p.config.DistrSrcPath))
		}
	}
	log.Println(fmt.Sprintf("%s: %v","distr_src_path", p.config.DistrSrcPath))

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *Provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {

	var err error
	errorMsg := "Error preparing shell script, %s error: %s"

	ui.Say("Provisioning...")

	if len(p.config.DistrSrcPath) != 0 {
		err = comm.UploadDir("skiped param", p.config.DistrSrcPath, nil)
		if err != nil {
			return err
		}
	}

	tempDir := os.TempDir()
	packerTempDir, err := ioutil.TempDir(tempDir, "packer_script")
	if err != nil {
		err := fmt.Errorf("Error creating temporary directory: %s", err.Error())
		return err
	}

	// create a temporary script file and upload it to Azure storage
	ui.Message("Preparing execution script...")
	provisionFileName := fmt.Sprintf("provision-%s.ps1", uuid.New())

	scriptPath := filepath.Join(packerTempDir, provisionFileName)
	tf, err := os.Create(scriptPath)

	if err != nil {
		return fmt.Errorf(errorMsg, "os.Create",  err.Error())
	}

	defer os.RemoveAll(packerTempDir)

	writer := bufio.NewWriter(tf)

	if p.config.Inline != nil {
		log.Println("Writing inline commands to execution script...")

		// Write our contents to it
		for _, command := range p.config.Inline {
			log.Println(command)
			if _, err := writer.WriteString(command + ";\n"); err != nil {
				return fmt.Errorf(errorMsg, "writer.WriteString",  err.Error())
			}
		}
	}

	// add content to the temp script file
	if len(p.config.ScriptPath) != 0 {
		log.Println("Writing file commands to execution script...")

		f, err := os.Open(p.config.ScriptPath)
		if err != nil {
			return fmt.Errorf(errorMsg, "os.Open",  err.Error())
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			log.Println(scanner.Text())
			if _, err := writer.WriteString(scanner.Text() + "\n"); err != nil {
				return fmt.Errorf(errorMsg, "writer.WriteString",  err.Error())
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf(errorMsg, "scanner.Scan",  err.Error())
		}
	}

	log.Println("Writing SysPrep to execution script...")
	sysprepPs := []string{
		"Write-Host 'Executing Sysprep from File...'",
		"Start-Process $env:windir\\System32\\Sysprep\\sysprep.exe -NoNewWindow -Wait -Argument '/quiet /generalize /oobe /quit'",
		"Write-Host 'Sysprep is done!'",
	}

	for _, command := range sysprepPs {
		log.Println(command)
		if _, err := writer.WriteString(command + ";\n"); err != nil {
			return fmt.Errorf(errorMsg, "writer.WriteString",  err.Error())
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("Error preparing shell script: %s", err.Error())
	}

	tf.Close()

	// upload to Azure storage
	ui.Message("Uploading execution script...")
	err = comm.UploadDir("skiped param", scriptPath, nil)
	if err != nil {
		return err
	}

	// execute script with Custom script extension
	runScript := provisionFileName

	var stdoutBuff bytes.Buffer
	var stderrBuff bytes.Buffer
	var cmd packer.RemoteCmd
	cmd.Stdout = &stdoutBuff;
	cmd.Stderr = &stderrBuff;

	cmd.Command = runScript

	ui.Message("Starting provisioning. It may take some time...")
	err = comm.Start(&cmd)
	if err != nil {
		err = fmt.Errorf(errorMsg, "comm.Start", err.Error())
		return err
	}
	
	ui.Message("Provision is Completed")

	stderrString := stderrBuff.String()
	if(len(stderrString)>0) {
		err = fmt.Errorf(errorMsg, "stderrString", stderrString)
		log.Printf("Provision stderr: %s", stderrString)
	}

	ui.Say("Script output")

	stdoutString := stdoutBuff.String()
	if(len(stdoutString)>0) {
		log.Printf("Provision stdout: %s", stdoutString)
		scriptMessages := strings.Split(stdoutString, "\\n")
		for _, m := range scriptMessages{
			ui.Message(m)
		}
	}

	return err
}

func (p *Provisioner) Cancel() {
	// Just hard quit. It isn't a big deal if what we're doing keeps
	// running on the other side.
	os.Exit(0)
}
