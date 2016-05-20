/*
Copyright 2013 Brice Figureau

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"crypto/x509/pkix"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/masterzen/winrm/winrm"
)

func main() {
	var (
		hostname string
		user     string
		pass     string
		cmd      string
		port     int
		https    bool
		insecure bool
		cacert   string
		gencert  bool
		certsize string
		timeout  string
	)

	flag.StringVar(&hostname, "hostname", "localhost", "winrm host")
	flag.StringVar(&user, "username", "vagrant", "winrm admin username")
	flag.StringVar(&pass, "password", "vagrant", "winrm admin password")
	flag.IntVar(&port, "port", 5985, "winrm port")
	flag.BoolVar(&https, "https", false, "use https")
	flag.BoolVar(&insecure, "insecure", false, "skip SSL validation")
	flag.StringVar(&cacert, "cacert", "", "CA certificate to use")
	flag.BoolVar(&gencert, "gencert", false, "Generate x509 client certificate to use with secure connections")
	flag.StringVar(&certsize, "certsize", "", "Priv RSA key between 512, 1024, 2048, 4096. Default :2048")
	flag.StringVar(&timeout, "timeout", "0s", "connection timeout")

	flag.Parse()

	if gencert {
		cersize := pickSizeCert(certsize)
		config := winrm.CertConfig{
			Subject: pkix.Name{
				CommonName: "winrm client cert",
			},
			ValidFrom: time.Now(),
			ValidFor:  365 * 24 * time.Hour,
			SizeT:     cersize,
			Method:    winrm.RSA,
		}

		certPem, privPem, err := winrm.NewCert(config)
		check(err)
		err = ioutil.WriteFile("cert.cer", []byte(certPem), 0644)
		check(err)
		err = ioutil.WriteFile("priv.pem", []byte(privPem), 0644)
		check(err)
	} else {

		var (
			certBytes      []byte
			err            error
			connectTimeout time.Duration
		)

		if cacert != "" {
			certBytes, err = ioutil.ReadFile(cacert)
			check(err)
		} else {
			certBytes = nil
		}

		cmd = flag.Arg(0)

		connectTimeout, err = time.ParseDuration(timeout)
		check(err)

		endpoint := winrm.NewEndpointWithTimeout(hostname, port, https, insecure, &certBytes, connectTimeout)
		client, err := winrm.NewClient(endpoint, user, pass)
		check(err)

		exitCode, err := client.RunWithInput(cmd, os.Stdout, os.Stderr, os.Stdin)
		check(err)

		os.Exit(exitCode)
	}
}

func pickSizeCert(size string) int {
	switch size {
	case "512":
		return 512
	case "1024":
		return 1024
	case "2048":
		return 2048
	case "4096":
		return 4096
	default:
		return 2048
	}
}

// generic check error func
func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
