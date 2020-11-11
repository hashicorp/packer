package ssh_test

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/user"
	"testing"
	"time"

	helperssh "github.com/hashicorp/packer/helper/ssh"
	"golang.org/x/crypto/ssh"
)

func getIdentityCertFile() (certSigner ssh.Signer, err error) {
	usr, _ := user.Current()
	privateKeyFile := usr.HomeDir + "/.ssh/id_ed25519"
	certificateFile := usr.HomeDir + "/.ssh/id_ed25519-cert.pub"

	return helperssh.FileSignerWithCert(privateKeyFile, certificateFile)
}

func TestConnectFunc(t *testing.T) {
	{
		if os.Getenv("PACKER_ACC") == "" {
			t.Skip("This test is only run with PACKER_ACC=1")
		}

		const host = "mybastionhost.com:2222"

		certSigner, err := getIdentityCertFile()
		if err != nil {
			panic(fmt.Errorf("we have an error %v", err))
		}

		publicKeys := ssh.PublicKeys(certSigner)
		usr, _ := user.Current()

		config := &ssh.ClientConfig{
			User: usr.Username,
			Auth: []ssh.AuthMethod{
				publicKeys,
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         30 * time.Second,
		}

		println("Dialing", config.User)
		connection, err := ssh.Dial("tcp", host, config)

		if err != nil {
			log.Fatal("Failed to dial ", err)
			return
		}

		session, err := connection.NewSession()
		if err != nil {
			log.Fatal("Failed to create session: ", err)
			return
		}
		defer session.Close()

		var stdoutBuf bytes.Buffer
		session.Stdout = &stdoutBuf

		err = session.Run("ls")
		if err != nil {
			log.Fatal("Failed to ls")
		}
		fmt.Println(stdoutBuf.String())
	}
}
