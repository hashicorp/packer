// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package driver_restapi

import (
	"encoding/xml"
	"os"
	"fmt"
	"io/ioutil"
	"encoding/base64"
	"path/filepath"
	"runtime"
	"os/exec"
	"bytes"
	"strings"
	"crypto/md5"
	"log"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/constants"
	"os/user"
)

type SubscriptionInfo struct {
	Id string
	CertData []byte
}

type publishData struct {
	PublishProfile publishProfile `xml:"PublishProfile"`
}

type publishProfile struct {
	SchemaVersion string `xml:",attr"`
	PublishMethod string `xml:",attr"`
	Url string `xml:",attr"`
	ManagementCertificate string `xml:",attr"`
	Subscriptions []subscription `xml:"Subscription"`
}

type subscription struct {
	ServiceManagementUrl string `xml:",attr"`
	Id string `xml:",attr"`
	Name string `xml:",attr"`
	ManagementCertificate string `xml:",attr"`
}

func ParsePublishSettings(path string, subscriptionName string) (*SubscriptionInfo, error){

	var err error
	if _, err = os.Stat(path); err != nil {
		err = fmt.Errorf("ParsePublishSettings: '%v' check the path is correct.", path)
		return nil, err
	}

	if len(subscriptionName)==0 {
		err = fmt.Errorf("ParsePublishSettings: '%v' subscriptionName is empty.", subscriptionName)
		return nil, err
	}

	log.Println(fmt.Sprintf("Reading file %s", path))

	var fileData []byte
	fileData, err = ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	log.Println(fmt.Sprintf("calculating md5..."))
	fileSumMd5Digest := fmt.Sprintf("%x", md5.Sum(fileData))

	publishData := publishData{}

	log.Println(fmt.Sprintf("parsing public settings..."))
	err = xml.Unmarshal(fileData, &publishData)
	if err != nil {
		return nil, err
	}

	if len(publishData.PublishProfile.Subscriptions) == 0 {
		err = fmt.Errorf("ParsePublishSettings: Subscriptions section is empty.")
		return nil, err
	}

	id := "none"
	certBase64 := publishData.PublishProfile.ManagementCertificate

	log.Println(fmt.Sprintf("looking for subscription info..."))
	for _, s := range(publishData.PublishProfile.Subscriptions) {
		if s.Name == subscriptionName {
			if len(s.Id) > 0 {
				id = s.Id
			} else {
				err = fmt.Errorf("ParsePublishSettings: subscription id is empty.")
				return nil, err
			}
			if len(s.ManagementCertificate) > 0 {
				certBase64 = s.ManagementCertificate
			} else if len(certBase64) == 0 {
				err = fmt.Errorf("ParsePublishSettings: ManagementCertificate is empty.")
				return nil, err
			}

			break
		}
	}

	if id == "none" {
		err = fmt.Errorf("ParsePublishSettings: Can't find subscriptionName '%v' in the file '%v'.", subscriptionName, path)
		return nil, err
	}

	log.Println(fmt.Sprintf("checking certificate..."))

	packerSubscriptionStoreDirName := ".packer_azure"

	var usrHome string

	if runtime.GOOS == constants.Windows {
		usrHome = os.TempDir()
	} else {
		log.Println(fmt.Sprintf("getting user home dir..."))
		// on Windows this operation takes too long (3+ minutes)
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}

		usrHome = usr.HomeDir
	}

	log.Println( usrHome )

	packerSubscriptionStoreDirPath := filepath.Join(usrHome, packerSubscriptionStoreDirName)
	subscrPath := filepath.Join(packerSubscriptionStoreDirPath,id)
	tagFilePath := filepath.Join(subscrPath,".tag")

	var modeDir os.FileMode = 0700
	var modeFile os.FileMode = 0600

	if _, err = os.Stat(packerSubscriptionStoreDirPath); err != nil {
		// create storage dir
		log.Println(fmt.Sprintf("creating packer folder..."))
		err = os.Mkdir(packerSubscriptionStoreDirPath, modeDir)
		if err != nil {
			return nil, err
		}
	}

	if _, err = os.Stat(subscrPath); err != nil {
		// create subscr dir
		log.Println(fmt.Sprintf("creating subscription folder..."))
		err = os.Mkdir(subscrPath, modeDir)
		if err != nil {
			return nil, err
		}
	}

	renewCert := false

	if _, err = os.Stat(tagFilePath); err != nil {
		renewCert = true
	} else {
		// read tag file
		log.Println(fmt.Sprintf("reading tag file..."))
		tagFileData, err := ioutil.ReadFile(tagFilePath)
		if err != nil {
			return nil, err
		}

		if string(tagFileData) != fileSumMd5Digest {
			renewCert = true
		}
	}

	certPemFilename := "cert.pem"
	certPemPath := filepath.Join(subscrPath, certPemFilename)

	if renewCert {
		log.Println("creating pemfile...")

		// put tag file here
		err := ioutil.WriteFile(tagFilePath, []byte(fileSumMd5Digest), modeFile)
		if err != nil {
			return nil, err
		}

		certPfxFilename := "cert.pfx"
		certPfxPath := filepath.Join(subscrPath, certPfxFilename)

		decBytes, err := base64.StdEncoding.DecodeString(certBase64)
		if err != nil {
			return nil, err
		}

		// Save data as pfx file
		err = ioutil.WriteFile(certPfxPath, decBytes, modeFile)
		if err != nil {
			return nil, err
		}

		// Find openssl
		progName := "openssl"
		binary, err := exec.LookPath(progName)
		if err != nil {
			err := fmt.Errorf("Can't find '%s' programm: %s", progName,  err.Error())
			return nil, err
		}

		if runtime.GOOS == constants.Linux {

			log.Println("executing openssl")
			err = Exec(binary, "pkcs12", "-in", certPfxPath, "-out", certPemPath, "-nodes", "-passin", "pass:")
			if err != nil {
				return nil, err
			}

		} else if runtime.GOOS == constants.Windows {

			var blockBuffer bytes.Buffer
			blockBuffer.WriteString("Invoke-Command -scriptblock {")
			blockBuffer.WriteString("$binary = '" + binary + "';")
			blockBuffer.WriteString("$cert_pfx = '" + certPfxPath + "';")
			blockBuffer.WriteString("$cert_pem = '" + certPemPath + "';")
			blockBuffer.WriteString("$args = \"pkcs12 -in $cert_pfx -out $cert_pem -nodes -passin pass:\";")
			blockBuffer.WriteString("Start-Process $binary -NoNewWindow -Wait -Argument $args;")
			blockBuffer.WriteString("}")

			err = Exec("powershell", blockBuffer.String())
			if err != nil {
				return nil, err
			}
		}
	}

	log.Println("reading pemfile: " + certPemPath )
	var pemData []byte
	pemData, err = ioutil.ReadFile(certPemPath)
	if err != nil {
		return nil, err
	}

	si := &SubscriptionInfo{
		Id : id,
		CertData: pemData,
	}

	return si, nil
}

func Exec(name string, arg ...string) error {

	log.Printf("Executing: %#v\n", arg)

	var stdout, stderr bytes.Buffer

	script := exec.Command(name, arg...)
	script.Stdout = &stdout
	script.Stderr = &stderr

	err := script.Run()

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("Exec error: %s\n", err)
	}

	stderrString := strings.TrimSpace(stderr.String())
	stdoutString := strings.TrimSpace(stdout.String())

	log.Printf("Exec stdout: %s\n", stdoutString)
	log.Printf("Exec stderr: %s\n", stderrString)

	return err
}

