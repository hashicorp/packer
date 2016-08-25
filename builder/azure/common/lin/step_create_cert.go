// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package lin

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/mitchellh/packer/builder/azure/common/constants"

	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type StepCreateCert struct {
	TmpServiceName string
}

func (s *StepCreateCert) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating temporary certificate...")

	err := s.createCert(state)
	if err != nil {
		err = fmt.Errorf("Error creating temporary certificate: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepCreateCert) Cleanup(state multistep.StateBag) {}

func (s *StepCreateCert) createCert(state multistep.StateBag) error {

	log.Println("createCert: Generating RSA key pair...")

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		err = fmt.Errorf("Failed to Generate Private Key: %s", err)
		return err
	}

	// ASN.1 DER encoded form
	privkey := string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	}))

	// Set the private key in the state bag for later
	state.Put(constants.PrivateKey, privkey)
	log.Printf("createCert: Private key:\n%s", privkey)

	log.Println("createCert: Creating certificate...")

	host := fmt.Sprintf("%s.cloudapp.net", s.TmpServiceName)
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		err = fmt.Errorf("Failed to Generate Serial Number: %v", err)
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

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		err = fmt.Errorf("Failed to Create Certificate: %s", err)
		return err
	}

	cert := string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derBytes,
	}))
	state.Put(constants.Certificate, cert)
	log.Printf("createCert: Certificate:\n%s", cert)

	h := sha1.New()
	h.Write(derBytes)
	thumbprint := fmt.Sprintf("%X", h.Sum(nil))
	state.Put(constants.Thumbprint, thumbprint)
	log.Printf("createCert: Thumbprint:\n%s", thumbprint)

	return nil
}
