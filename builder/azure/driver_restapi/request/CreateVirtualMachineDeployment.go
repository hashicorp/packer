// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.

package request

import (
	"fmt"
	"bytes"
)

func (m *Manager) CreateVirtualMachineDeploymentLin(isOSImage bool, serviceName, vmName, vmSize, certThumbprint, userName, osImageName, mediaLoc string) (*Data) {

	uri := fmt.Sprintf("https://management.core.windows.net/%s/services/hostedservices/%s/deployments", m.SubscrId, serviceName)

	headers := map[string]string{
		"Content-Type":  "application/xml",
		"x-ms-version":  "2014-02-01",
	}

	var buff bytes.Buffer
	buff.WriteString("<Deployment xmlns:i='http://www.w3.org/2001/XMLSchema-instance' xmlns='http://schemas.microsoft.com/windowsazure'>")
	buff.WriteString("<Name>"+ vmName +"</Name>")
	buff.WriteString("<DeploymentSlot>Production</DeploymentSlot>")
	buff.WriteString("<Label>"+ vmName +"</Label>")
	buff.WriteString("<RoleList>")
	buff.WriteString("<Role i:type='PersistentVMRole'>")
		buff.WriteString("<RoleName>"+ vmName +"</RoleName>")
		buff.WriteString("<RoleType>PersistentVMRole</RoleType>")
		buff.WriteString("<ConfigurationSets>")
			buff.WriteString("<ConfigurationSet i:type='LinuxProvisioningConfigurationSet'>")
				buff.WriteString("<ConfigurationSetType>LinuxProvisioningConfiguration</ConfigurationSetType>")
				buff.WriteString("<HostName>"+ vmName +"</HostName>")
				buff.WriteString("<UserName>"+ userName +"</UserName>")
				buff.WriteString("<DisableSshPasswordAuthentication>true</DisableSshPasswordAuthentication>")
				buff.WriteString("<SSH>")
					buff.WriteString("<PublicKeys>")
						buff.WriteString("<PublicKey>")
						buff.WriteString("<Fingerprint>"+ certThumbprint +"</Fingerprint>")
						buff.WriteString("<Path>/home/"+ userName +"/.ssh/authorized_keys</Path>")
						buff.WriteString("</PublicKey>")
					buff.WriteString("</PublicKeys>")
//					buff.WriteString("<KeyPairs>")
//						buff.WriteString("<KeyPair>")
//						buff.WriteString("<Fingerprint>"+ certThumbprint +"</Fingerprint>")
//						buff.WriteString("<Path>/home/"+ userName +"/.ssh/id_rsa</Path>")
//						buff.WriteString("</KeyPair>")
//					buff.WriteString("</KeyPairs>")
				buff.WriteString("</SSH>")
			buff.WriteString("</ConfigurationSet>")
			buff.WriteString("<ConfigurationSet i:type='NetworkConfigurationSet'>")
				buff.WriteString("<ConfigurationSetType>NetworkConfiguration</ConfigurationSetType>")
				buff.WriteString("<InputEndpoints>")
					buff.WriteString("<InputEndpoint>")
					buff.WriteString("<LocalPort>22</LocalPort>")
					buff.WriteString("<Name>SSH</Name>")
					buff.WriteString("<Protocol>tcp</Protocol>")
					buff.WriteString("</InputEndpoint>")
				buff.WriteString("</InputEndpoints>")
			buff.WriteString("</ConfigurationSet>")
		buff.WriteString("</ConfigurationSets>")
	if !isOSImage {
		buff.WriteString("<VMImageName>"+ osImageName +"</VMImageName>")
	} else {
		buff.WriteString("<OSVirtualHardDisk>")
		buff.WriteString("<MediaLink>"+ mediaLoc +"</MediaLink>")
		buff.WriteString("<SourceImageName>"+ osImageName +"</SourceImageName>")
		buff.WriteString("</OSVirtualHardDisk>")
	}
	buff.WriteString("<RoleSize>"+ vmSize +"</RoleSize>")
	buff.WriteString("<ProvisionGuestAgent>true</ProvisionGuestAgent>")
	buff.WriteString("</Role>")
	buff.WriteString("</RoleList>")
	buff.WriteString("</Deployment>")

	data := &Data{
		Verb: "POST",
		Uri : uri,
		Headers: headers,
		Body : &buff,
	}

	return data
}

func (m *Manager) CreateVirtualMachineDeploymentWin(isOSImage bool, serviceName, vmName, vmSize, userName, userPassword, osImageName, mediaLoc string) (*Data) {

	uri := fmt.Sprintf("https://management.core.windows.net/%s/services/hostedservices/%s/deployments", m.SubscrId, serviceName)

	headers := map[string]string{
		"Content-Type":  "application/xml",
		"x-ms-version":  "2014-05-01",
	}

	var buff bytes.Buffer
	buff.WriteString("<Deployment xmlns:i='http://www.w3.org/2001/XMLSchema-instance' xmlns='http://schemas.microsoft.com/windowsazure'>")
	buff.WriteString("<Name>"+ vmName +"</Name>")
	buff.WriteString("<DeploymentSlot>Production</DeploymentSlot>")
	buff.WriteString("<Label>"+ vmName +"</Label>")
	buff.WriteString("<RoleList>")
		buff.WriteString("<Role i:type='PersistentVMRole'>")
			buff.WriteString("<RoleName>"+ vmName +"</RoleName>")
			buff.WriteString("<RoleType>PersistentVMRole</RoleType>")
			buff.WriteString("<ConfigurationSets>")
				buff.WriteString("<ConfigurationSet i:type='WindowsProvisioningConfigurationSet'>")
					buff.WriteString("<ConfigurationSetType>WindowsProvisioningConfiguration</ConfigurationSetType>")
					buff.WriteString("<ComputerName>"+ vmName +"</ComputerName>")
					buff.WriteString("<AdminPassword>"+ userPassword +"</AdminPassword>")
					buff.WriteString("<EnableAutomaticUpdates>true</EnableAutomaticUpdates>")
					buff.WriteString("<AdminUsername>"+ userName +"</AdminUsername>")
				buff.WriteString("</ConfigurationSet>")
			buff.WriteString("</ConfigurationSets>")
	if !isOSImage {
			buff.WriteString("<VMImageName>"+ osImageName +"</VMImageName>")
	} else {
			buff.WriteString("<OSVirtualHardDisk>")
				buff.WriteString("<MediaLink>"+ mediaLoc +"</MediaLink>")
				buff.WriteString("<SourceImageName>"+ osImageName +"</SourceImageName>")
			buff.WriteString("</OSVirtualHardDisk>")
	}
			buff.WriteString("<Label>"+ vmName +"</Label>")
			buff.WriteString("<RoleSize>"+ vmSize +"</RoleSize>")
			buff.WriteString("<ProvisionGuestAgent>true</ProvisionGuestAgent>")
		buff.WriteString("</Role>")
	buff.WriteString("</RoleList>")
	buff.WriteString("</Deployment>")

	data := &Data{
		Verb: "POST",
		Uri : uri,
		Headers: headers,
		Body : &buff,
	}

	return data
}
