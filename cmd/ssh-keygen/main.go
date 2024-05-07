// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package main

import (
	"flag"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/hashicorp/packer-plugin-sdk/communicator/sshkey"
)

type options struct {
	Type     string
	Bits     int
	Filename string
}

func (o *options) AddFlagSets(fs *flag.FlagSet) {
	fs.StringVar(&o.Type, "type", "rsa", `dsa | ecdsa | ed25519 | rsa
Specifies the type of key to create. The possible values are 'dsa', 'ecdsa',
'ed25519', or 'rsa'.
`)
	fs.IntVar(&o.Bits, "bits", 0, `Specifies the number of bits in the key to create. By default maximum
number will be picked. For RSA keys, the minimum size is 1024 bits and the
default is 3072 bits. Generally, 3072 bits is considered sufficient. DSA
keys must be exactly 1024 bits as specified by FIPS 186-2. For ECDSA keys,
the bits flag determines the key length by selecting from one of three
elliptic curve sizes: 256, 384 or 521 bits. Attempting to use bit lengths
other than these three values for ECDSA keys will fail. Ed25519 keys have a
fixed length and the bits flag will be ignored.
`)

	defaultPath := ""
	user, err := user.Current()
	if err == nil {
		defaultPath = filepath.Join(user.HomeDir, ".ssh", "tests")
	}

	fs.StringVar(&o.Filename, "filename", defaultPath, `Specifies the filename of the key file.
`)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("ssh-keygen: ")
	fs := flag.NewFlagSet("ssh-keygen", flag.ContinueOnError)
	cla := options{}
	cla.AddFlagSets(fs)
	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	algo, err := sshkey.AlgorithmString(cla.Type)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Generating public/private %s key pair.", algo)

	keypair, err := sshkey.GeneratePair(algo, nil, cla.Bits)
	if err != nil {
		log.Fatal(err)
	}

	if isDir(cla.Filename) {
		cla.Filename = filepath.Join(cla.Filename, "id_"+algo.String())
	}
	if fileExists(cla.Filename) {
		log.Fatalf("%s already exists.", cla.Filename)
	}
	log.Printf("Saving private key to %s", cla.Filename)
	if err := os.WriteFile(cla.Filename, keypair.Private, 0600); err != nil {
		log.Fatal(err)
	}
	publicFilename := cla.Filename + ".pub"
	log.Printf("Saving public key to %s", publicFilename)
	if err := os.WriteFile(publicFilename, keypair.Public, 0644); err != nil {
		log.Fatal(err)
	}
}

func isDir(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
