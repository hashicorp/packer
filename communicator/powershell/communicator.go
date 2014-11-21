// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package powershell

import (
	"fmt"
	"bytes"
	"github.com/mitchellh/packer/packer"
	"io"
	"log"
	"path/filepath"
	"os/exec"
	"os"
	"strings"
	"container/list"
)


type comm struct {
	config *Config
}

type Config struct {
	Username string
	Password string
	RemoteHostIP string
	VmName string
	Ui packer.Ui
}

// Creates a new packer.Communicator implementation over SSH. This takes
// an already existing TCP connection and SSH configuration.
func New(config *Config) (result *comm, err error) {
	// Establish an initial connection and connect
	result = &comm{
		config: config,
	}

	return
}

func (c *comm) Start(cmd *packer.RemoteCmd) (err error) {
	username := c.config.Username
	password := c.config.Password
	remoteHost := c.config.RemoteHostIP

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock { ")
	blockBuffer.WriteString("$ip4 = '" + remoteHost + "';")
	blockBuffer.WriteString("$username = '" + username + "';")
	blockBuffer.WriteString("$password = '" + password + "';")
	blockBuffer.WriteString("$secstr = New-Object -TypeName System.Security.SecureString;")
	blockBuffer.WriteString("$password.ToCharArray() | ForEach-Object {$secstr.AppendChar($_)};")
	blockBuffer.WriteString("$cred = new-object -typename System.Management.Automation.PSCredential -argumentlist $username, $secstr;")
	blockBuffer.WriteString("Invoke-Command -ComputerName $ip4 ")
	blockBuffer.WriteString(cmd.Command)
	blockBuffer.WriteString(" -credential $cred")
	blockBuffer.WriteString("}")

	log.Printf("Start blockBuffer: %s",  blockBuffer.String())

	script := exec.Command("powershell", blockBuffer.String())

	script.Stdin = cmd.Stdin
	script.Stdout = cmd.Stdout
	script.Stderr = cmd.Stderr

	log.Printf(fmt.Sprintf("Executing remote script..."))
	err = script.Run()

	return
}

func (c *comm) Upload(string, io.Reader, *os.FileInfo) error {
	panic("not implemented for powershell")
}

func (c *comm) UploadDir(dst string, src string, excl []string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	ui := c.config.Ui

	if info.IsDir() {
		ui.Say(fmt.Sprintf("Uploading folder to the VM '%s' => '%s'...",  src, dst))
		err := c.uploadFolder(dst, src)
		if err != nil {
			return err
		}
	} else {
		target_file := filepath.Join(dst,filepath.Base(src))
		ui.Say(fmt.Sprintf("Uploading file to the VM '%s' => '%s'...", src, target_file))
		err := c.uploadFile(target_file, src)
		if err != nil {
			return err
		}
	}

	return err
}

func (c *comm) Download(string, io.Writer) error {
	panic("not implemented yet")
}

// region private helpers

func (c *comm) uploadFile(dscPath string, srcPath string) error {

	dscPath = filepath.FromSlash(dscPath)
	srcPath = filepath.FromSlash(srcPath)

	vmName := c.config.VmName

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("Invoke-Command -scriptblock { ")
	blockBuffer.WriteString("Copy-VMFile ")
	blockBuffer.WriteString("'" + vmName + "' ")
	blockBuffer.WriteString("-SourcePath ")
	blockBuffer.WriteString("'" + srcPath + "' ")
	blockBuffer.WriteString("-DestinationPath ")
	blockBuffer.WriteString("'" + dscPath + "' ")
	blockBuffer.WriteString("-CreateFullPath -FileSource Host -Force ")
	blockBuffer.WriteString("}")

	log.Printf("uploadFile blockBuffer: %s",  blockBuffer.String())

	script := exec.Command("powershell", blockBuffer.String())

	var stdout, stderr bytes.Buffer

	script.Stdout = &stdout
	script.Stderr = &stderr

	err := script.Run()

	stderrString := strings.TrimSpace(stdout.String())
	stdoutString := strings.TrimSpace(stdout.String())

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("uploadFile error: %s", stderrString)
	}

	if len(stderrString) > 0 {
		err = fmt.Errorf("uploadFile error: %s", stderrString)
	}

	return err
}

func (c *comm) uploadFolder(dscPath string, srcPath string ) error {
	l := list.New()

	type dstSrc struct {
		src string
		dst string
	}

	treeWalk := func(path string, info os.FileInfo, prevErr error) error {
		// If there was a prior error, return it
		if prevErr != nil {
			return prevErr
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(srcPath, path)
		if err != nil {
			return err
		}

		l.PushBack(
			dstSrc{
				src: path,
				dst: filepath.Join(dscPath,rel),
			})

		return nil
	}

	filepath.Walk(srcPath, treeWalk)

	var err error
	for e := l.Front(); e != nil; e = e.Next() {
		pair := e.Value.(dstSrc)
		log.Printf("'%s' ==> '%s'\n", pair.src, pair.dst)
		err = c.uploadFile(pair.dst, pair.src)
		if err != nil {
			return err
		}
	}

	return err
}

