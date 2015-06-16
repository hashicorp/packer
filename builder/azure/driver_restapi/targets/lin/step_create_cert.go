// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package lin

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"strings"
	"time"
	"io/ioutil"
	"path/filepath"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/cert"
	"github.com/mitchellh/packer/builder/azure/driver_restapi/constants"
)

type StepCreateCert struct {
	CertFileName string
	KeyFileName string
	TempDir string
	TmpServiceName string
}

func (s *StepCreateCert) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating temporary certificate...")

	if(len(s.TempDir) == 0){
		//Creating temporary directory
		ui.Message("Creating temporary directory...")
		tempDir := os.TempDir()
		packerTempDir, err := ioutil.TempDir(tempDir, "packer_cert")
		if err != nil {
			err := fmt.Errorf("Error creating temporary directory: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}

		s.TempDir = packerTempDir;
	}

	certPath := filepath.Join(s.TempDir, s.CertFileName)
	ui.Message("CertPath: " + certPath)

	err := s.createCert(state)
	if err != nil {
		err := fmt.Errorf("Error creating temporary certificate: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	thumbprint, err :=  cert.GetThumbprint(certPath)
	ui.Message("thumbprint: " + thumbprint)

	if err != nil {
		err = fmt.Errorf("Can't get certificate thumbprint '%s'", certPath)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put(constants.UserCertPath, certPath)
	state.Put(constants.UserCertThumbprint, thumbprint)
	state.Put(constants.CertCreated, 1)

	return multistep.ActionContinue
}

func (s *StepCreateCert) Cleanup(state multistep.StateBag) {

	if s.TempDir == "" {
		return
	}

	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting temporary certificate...")
	ui.Message("Deleting temporary directory...")

	err := os.RemoveAll(s.TempDir)

	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting temporary directory: %s", err))
	}
}

func (s *StepCreateCert)createCert(state multistep.StateBag) error {

	if(len(s.TempDir) == 0){
		return fmt.Errorf("StepCreateCert: s.TempDir is empty")
	}

	host  := fmt.Sprintf("%s.cloudapp.net", s.TmpServiceName)
	validFor  := 365*24*time.Hour
	isCA      := false
	rsaBits   := 2048

	priv, err := rsa.GenerateKey(rand.Reader, rsaBits)
	if err != nil {
		err := fmt.Errorf("Failed to generate private key: %s", err)
		return err
	}

	// ASN.1 DER encoded form
	priv_der := x509.MarshalPKCS1PrivateKey(priv)
	priv_blk := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   priv_der,
	}

	// Set the private key in the statebag for later
	state.Put(constants.PrivateKey, string(pem.EncodeToMemory(&priv_blk)))

	notBefore := time.Now()

	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		err := fmt.Errorf("Failed to generate serial number: %s: %s", err)
		return err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Issuer: pkix.Name{
			CommonName: host,
		},
		Subject: pkix.Name{
			CommonName: host,
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	hosts := strings.Split(host, ",")
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		err := fmt.Errorf("Failed to Create Certificate: %s", err)
		return err
	}

	certOut, err := os.Create(filepath.Join(s.TempDir, s.CertFileName))
	if err != nil {
		err := fmt.Errorf("Failed to open cert.pem for writing: %s: %s", err)
		return err
	}

	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	log.Printf("Written %s", s.CertFileName)

	keyOut, err := os.OpenFile(filepath.Join(s.TempDir, s.KeyFileName), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		err := fmt.Errorf("Failed to open key.pem for writing: %s", err)
		return err
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	keyOut.Close()
	log.Printf("Written %s", s.KeyFileName)

	return nil
}

