// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package azureVmCustomScriptExtension

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"io"
	"path/filepath"
	"os"
	"io/ioutil"
	"encoding/base64"
	"log"

	"github.com/mitchellh/packer/builder/azure/driver_restapi/response"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/response/model"
	azureservice "github.com/mitchellh/packer/builder/azure/driver_restapi/request"
	storageservice "github.com/mitchellh/packer/builder/azure/driver_restapi/storage_service/request"
	"time"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/utils"
	"code.google.com/p/go-uuid/uuid"
)

const extPublisher = "Microsoft.Compute"
const extName = "CustomScriptExtension"

type comm struct {
	config *Config
	uris string
}

type Config struct {
	ServiceName string
	VmName	string
	StorageServiceDriver *storageservice.StorageServiceDriver
	AzureServiceRequestManager *azureservice.Manager
	ContainerName string
	Ui packer.Ui
	IsOSImage bool
}

func New(config *Config) (result *comm, err error) {
	result = &comm{
		config: config,
	}

	return
}

func (c *comm) Start(cmd *packer.RemoteCmd) (err error) {
	ext, err := c.requestCustomScriptExtension()
	if err != nil {
		return
	}

	nameOfReference := fmt.Sprintf("PackerCustomScriptExtension-%s", uuid.New())
	nameOfPublisher := extPublisher
	nameOfExtension := extName
	versionOfExtension := ext.Version

	log.Println("Installing CustomScriptExtension...")
	state := "enable"
	params := c.buildParams(cmd.Command)

	err = c.updateRoleResourceExtension(nameOfReference, nameOfPublisher, nameOfExtension, versionOfExtension, state, params)
	if err != nil {
		return
	}

	stdOutBuff, stdErrBuff, err := c.pollCustomScriptExtensionIsReady()
	if err != nil {
		return
	}

	_, err = cmd.Stdout.Write([]byte(stdOutBuff))
	if err != nil {
		err = fmt.Errorf("cmd.Stdout error: %s", err.Error())
		return
	}

	_, err = cmd.Stderr.Write([]byte(stdErrBuff))
	if err != nil {
		err = fmt.Errorf("cmd.Stdout error: %s", err.Error())
		return
	}

	log.Println("Uninstalling CustomScriptExtension...")

	state = "uninstall"
	params = nil
	err = c.updateRoleResourceExtension(nameOfReference, nameOfPublisher, nameOfExtension, versionOfExtension, state, params)
	if err != nil {
		return
	}

	c.sleepSec(20)

	err = c.pollCustomScriptIsUninstalled()
	if err != nil {
		return
	}

	return
}

func (c *comm) sleepSec(d time.Duration){
	log.Printf("Sleep for %v sec", uint(d))
	time.Sleep(time.Second*d)
}

func (c *comm) requestCustomScriptExtension() (*model.ResourceExtension, error) {
	reqManager := c.config.AzureServiceRequestManager

	log.Println("Requesting resource extensions...")
	requestData := reqManager.ListResourceExtensions()
	resp, err := reqManager.Execute(requestData)

	if err != nil {
		return nil, err
	}

	list, err := response.ParseResourceExtensionList(resp.Body)

	if err != nil {
		return nil, err
	}

	log.Println("Searching for CustomScriptExtension...")
	ext := list.FirstOrNull("CustomScriptExtension")
	log.Printf("CustomScriptExtension: %v\n\n", ext)

	if ext == nil {
		err = fmt.Errorf("CustomScriptExtension is nil")
		return nil, err
	}
	
	return ext, nil
}

func (c *comm) updateRoleResourceExtension(
	nameOfReference, nameOfPublisher, nameOfExtension, versionOfExtension, state string,
	params []azureservice.ResourceExtensionParameterValue) error {

	reqManager := c.config.AzureServiceRequestManager

	serviceName := c.config.ServiceName
	vmName := c.config.VmName


	log.Println("Updating Role Resource Extension...")

	requestData := reqManager.UpdateRoleResourceExtensionReference(serviceName, vmName, nameOfReference, nameOfPublisher, nameOfExtension, versionOfExtension, state, params)
	err := reqManager.ExecuteSync(requestData)

	if err != nil {
		return err
	}

	return nil
}

func (c *comm)buildParams(runScript string) (params []azureservice.ResourceExtensionParameterValue) {
	storageAccountName, storageAccountKey := c.config.StorageServiceDriver.GetProps()

	account := "{\"storageAccountName\":\"" + storageAccountName + "\",\"storageAccountKey\": \"" + storageAccountKey + "\"}";

	scriptfile := "{\"fileUris\": [" + c.uris + "], \"commandToExecute\":\"powershell -ExecutionPolicy Unrestricted -file " + runScript + "\"}"

	params = []azureservice.ResourceExtensionParameterValue {
		azureservice.ResourceExtensionParameterValue{
			Key: "CustomScriptExtensionPublicConfigParameter",
			Value: base64.StdEncoding.EncodeToString([]byte(scriptfile)),
			Type: "Public",
		},
		azureservice.ResourceExtensionParameterValue{
			Key: "CustomScriptExtensionPrivateConfigParameter",
			Value: base64.StdEncoding.EncodeToString([]byte(account)),
			Type: "Private",
		},
	}

	return
}

func (c *comm) pollCustomScriptExtensionIsReady() (stdOutBuff, stdErrBuff string, err error) {
	reqManager := c.config.AzureServiceRequestManager
	log.Println("Polling CustomScriptExtension is ready. It may take some time...")

	var deployment *model.Deployment
	var res *model.ResourceExtensionStatus
	const statusSuccess = "Success"
	const statusError = "Error"

//	needUpdateStatus := true

	serviceName := c.config.ServiceName
	vmName := c.config.VmName

	const attemptLimit uint = 30;

	requestData := reqManager.GetDeployment(serviceName, vmName)
	updateCount := attemptLimit

	for ; updateCount > 0; updateCount--{

		repeatCount := attemptLimit
		for ; repeatCount > 0; repeatCount-- {
			resp, errEx := reqManager.Execute(requestData)

			if errEx != nil {
				err = errEx
				return
			}

			deployment, err = response.ParseDeployment(resp.Body)

			if err != nil {
				return
			}

			if deployment.RoleInstanceList[0].InstanceStatus == "ReadyRole" {
				if len(deployment.RoleInstanceList[0].ResourceExtensionStatusList) > 0 {
					break
				}
			}

			c.sleepSec(45)
		}

		if repeatCount == 0 {
			err = fmt.Errorf("InstanceStatus is not 'ReadyRole' or CustomScriptExtension ResourceExtensionStatusList is empty after %d attempts", attemptLimit)
			return
		}

		extHandlerName := extPublisher + "." + extName

		for _, s := range deployment.RoleInstanceList[0].ResourceExtensionStatusList {
			if s.HandlerName == extHandlerName {
				res = &s
			}
		}

		if res == nil {
			err = fmt.Errorf("CustomScriptExtension status not found")
			return
		}

		log.Printf("CustomScriptExtension status: %v", res)

		extensionSettingStatus := res.ExtensionSettingStatus

		if extensionSettingStatus.Status == statusError {
			err = fmt.Errorf("CustomScriptExtension operation '%s' status: %s", extensionSettingStatus.Operation, extensionSettingStatus.FormattedMessage.Message )
			return
		}

		log.Printf("CustomScriptExtension INFO: operation '%s' status: %s",extensionSettingStatus.Operation, extensionSettingStatus.Status)

		var stdOut, stdErr string

		for _, subStatus := range res.ExtensionSettingStatus.SubStatusList {
			if subStatus.Name == "StdOut" {
				if subStatus.Status != statusSuccess {
					stdOut = fmt.Sprintf("StdOut failed with message: '%s'", subStatus.FormattedMessage.Message)
				} else {
					stdOut = subStatus.FormattedMessage.Message
				}
				continue
			}

			if subStatus.Name == "StdErr" {
				if subStatus.Status != statusSuccess {
					stdErr = fmt.Sprintf("StdErr failed with message: '%s'", subStatus.FormattedMessage.Message)
				} else {
					stdErr = subStatus.FormattedMessage.Message
				}
				continue
			}
		}

		log.Printf("StdOut: '%s'\n", stdOut)

		if len(stdOutBuff) == 0 {
			stdOutBuff = stdOut
		} else {
			stdOutBuff = utils.Clue(stdOutBuff, stdOut)
		}

		if len(stdErrBuff) == 0 {
			stdErrBuff = stdErr
		} else {
			stdErrBuff = utils.Clue(stdErrBuff, stdErr)
		}

		if extensionSettingStatus.Status == statusSuccess {
			break
		}

		c.sleepSec(40)
	}

	if updateCount == 0 {
		err = fmt.Errorf("extensionSettingStatus.Status in not 'Success' after %d attempts", attemptLimit)
		return
	}

	return
}

func (c *comm) pollCustomScriptIsUninstalled() error {
	reqManager := c.config.AzureServiceRequestManager
	log.Println("Polling CustomScript is uninstalled. It may take some time...")

	serviceName := c.config.ServiceName
	vmName := c.config.VmName

	requestData := reqManager.GetDeployment(serviceName, vmName)
	const attemptLimit uint = 30;
	repeatCount := attemptLimit
	for ; repeatCount > 0; repeatCount-- {
		resp, err := reqManager.Execute(requestData)

		if err != nil {
			return err
		}

		deployment, err := response.ParseDeployment(resp.Body)

		if err != nil {
			return err
		}

		if deployment.RoleInstanceList[0].InstanceStatus == "ReadyRole" {
			if len(deployment.RoleInstanceList[0].ResourceExtensionStatusList) == 0 {
				break
			}
		}

		c.sleepSec(45)
	}

	if repeatCount == 0 {
		err := fmt.Errorf("InstanceStatus is not 'ReadyRole' or ResourceExtensionStatusList is not empty after %d attempts", attemptLimit)
		return err
	}

	return nil
}

func (c *comm)Upload(string, io.Reader, *os.FileInfo) error {
	return fmt.Errorf("Upload is not supported for azureVmCustomScriptExtension")
}

func (c *comm) UploadDir(skipped string, src string, excl []string) error {

	src = filepath.FromSlash(src)
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	containerName := c.config.ContainerName

	if info.IsDir() {
		log.Println(fmt.Sprintf("Uploading files (only!) in the folder to Azure storage container '%s' => '%s'...",  src, containerName))
		err := c.uploadFolder("", src)
		if err != nil {
			return err
		}
	} else {
		err := c.uploadFile("", src)
		if err != nil {
			return err
		}
	}

	return err
}

func (c *comm) Download(string, io.Writer) error {
	return fmt.Errorf("Download is not supported for azureVmCustomScriptExtension")
}

// region private helpers

func (c *comm) uploadFile(dscPath string, srcPath string) error {

	srcPath = filepath.FromSlash(srcPath)

	_, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("Check file path is correct: %s", srcPath)
	}

	ui := c.config.Ui
	sa := c.config.StorageServiceDriver

	storageAccountName, _ := c.config.StorageServiceDriver.GetProps()
	containerName := c.config.ContainerName

	fileName := filepath.Base(srcPath)
	uri := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", storageAccountName, containerName, fileName)

	if len(c.uris) == 0 {
		c.uris = fmt.Sprintf("\"%s\"", uri)
	} else {
		c.uris += fmt.Sprintf(", \"%s\"", uri)
	}

	log.Println("uris: '" + c.uris + "'")

	ui.Message(fmt.Sprintf("Uploading file to to Azure storage container '%s' => '%s'...", srcPath, containerName))

	_, err = sa.PutBlob(containerName, srcPath)

	return err
}

func (c *comm) uploadFolder(dscPath string, srcPath string ) error {

	srcPath = filepath.FromSlash(srcPath)

	files, err := ioutil.ReadDir(srcPath)
	if err != nil {
		return err
	}

	for _, f := range files {
		if (f.IsDir()) {
			continue
		}

		err := c.uploadFile("", filepath.Join(srcPath,f.Name()))
		if err != nil {
			return err
		}
	}

	return err
}

