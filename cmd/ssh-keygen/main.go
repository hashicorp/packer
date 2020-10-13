package main

import (
	"flag"
	"log"
	"os"

	"github.com/hashicorp/packer/helper/communicator/sshkey"
)

type options struct {
	Type string
	Bits int
}

func (o *options) AddFlagSets(fs *flag.FlagSet) {
	fs.StringVar(&o.Type, "type", "rsa", `dsa | ecdsa | ed25519 | rsa

	Specifies the type of key to create.  The possible values are 'dsa', 'ecdsa',
	'ed25519', or 'rsa' ( the default ).
`)
	fs.IntVar(&o.Bits, "bits", 0, `bits

	Specifies the number of bits in the key to create.  For RSA keys, the min-
	imum size is 1024 bits and the default is 3072 bits.  Generally, 3072 bits
	is considered sufficient.  DSA keys must be exactly 1024 bits as specified
	by FIPS 186-2.  For ECDSA keys, the bits flag determines the key length by
	selecting from one of three elliptic curve sizes: 256, 384 or 521 bits.
	Attempting to use bit lengths other than these three values for ECDSA keys
	will fail.  Ed25519 keys have a fixed length and the bits flag will be
	ignored.	
`)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("ssh-keygen: ")
	fs := flag.NewFlagSet("ssh-keygen", flag.ContinueOnError)
	cla := options{}
	cla.AddFlagSets(fs)
	if err := fs.Parse(os.Args[1:]); err != nil {
		fs.Usage()
		os.Exit(1)
	}

	algo, err := sshkey.AlgorithmString(cla.Type)
	if err != nil {
		log.Fatal(err)
	}

	keypair, err := sshkey.GeneratePair(algo, nil, cla.Bits)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("keypair.Private:")
	log.Printf("%s", keypair.Private)
	log.Printf("keypair.Public:")
	log.Printf("%s", keypair.Public)
}
