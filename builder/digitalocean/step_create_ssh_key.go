package digitalocean

import (
	"cgl.tideland.biz/identifier"
	"encoding/hex"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type stepCreateSSHKey struct {
	keyId uint
}

func (s *stepCreateSSHKey) Run(state map[string]interface{}) multistep.StepAction {
	client := state["client"].(*DigitalOceanClient)
	ui := state["ui"].(packer.Ui)

	ui.Say("Creating temporary ssh key for droplet...")
	// priv, err := rsa.GenerateKey(rand.Reader, 2014)
	// if err != nil {
	// 	ui.Error(err.Error())
	// 	return multistep.ActionHalt
	// }

	// priv_der := x509.MarshalPKCS1PrivateKey(priv)
	// priv_blk := pem.Block{
	// 	Type:    "RSA PRIVATE KEY",
	// 	Headers: nil,
	// 	Bytes:   priv_der,
	// }

	// Set the pem formatted private key on the state for later
	// state["privateKey"] = string(pem.EncodeToMemory(&priv_blk))
	// log.Printf("PRIVATE KEY:\n\n%v\n\n", state["privateKey"])

	// Create the public key for uploading to DO
	// pub := priv.PublicKey

	// pub_bytes, err := x509.MarshalPKIXPublicKey(&pub)

	// pub_blk := pem.Block{
	// 	Type:    "RSA PUBLIC KEY",
	// 	Headers: nil,
	// 	Bytes:   pub_bytes,
	// }

	// if err != nil {
	// 	ui.Error(err.Error())
	// 	return multistep.ActionHalt
	// }

	// // Encode the public key to base64
	// pub_str := base64.StdEncoding.EncodeToString(pub_bytes)
	// pub_str = "ssh-rsa " + pub_str

	// log.Printf("PUBLIC KEY:\n\n%v\n\n", string(pem.EncodeToMemory(&pub_blk)))
	// log.Printf("PUBLIC KEY BASE64:\n\n%v\n\n", pub_str)

	pub_str := `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD29LZNMe0f7nOmdOIXDrF6eAmLZEk1yrnnsPI+xjLsnKxggMjdD3HvkBPXMdhakOj3pEF6DNtXbK43A7Pilezvu7y2awz+dxCavgUNtwaJkiTJw3C2qleNDDgrq7ZYLJ/wKmfhgPO4jZBej/8ONA0VjxemCNBPTTBeZ8FaeOpeUqopdhk78KGeGmUJ8Bvl8ACuYNdtJ5Y0BQCZkJT+g1ntTwHvuq/Vy/E2uCwJ2xV3vCDkLlqXVyksuVIcLJxTPtd5LdasD4WMQwoOPNdNMBLBG6ZBhXC/6kCVbMgzy5poSZ7r6BK0EA6b2EdAanaojYs3i52j6JeCIIrYtu9Ub173 jack@jose.local`
	state["privateKey"] = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA9vS2TTHtH+5zpnTiFw6xengJi2RJNcq557DyPsYy7JysYIDI
3Q9x75AT1zHYWpDo96RBegzbV2yuNwOz4pXs77u8tmsM/ncQmr4FDbcGiZIkycNw
tqpXjQw4K6u2WCyf8Cpn4YDzuI2QXo//DjQNFY8XpgjQT00wXmfBWnjqXlKqKXYZ
O/ChnhplCfAb5fAArmDXbSeWNAUAmZCU/oNZ7U8B77qv1cvxNrgsCdsVd7wg5C5a
l1cpLLlSHCycUz7XeS3WrA+FjEMKDjzXTTASwRumQYVwv+pAlWzIM8uaaEme6+gS
tBAOm9hHQGp2qI2LN4udo+iXgiCK2LbvVG9e9wIDAQABAoIBABuBB6izTciHoyO/
0spknYmZQt7ebXTrPic6wtAQ/OzzShN5ZGWSacsXjc4ixAjaKMgj6BLyyZ8EAKcp
52ft8LSGgS8D3y+cDSJe1WtAnh7GQwihlrURZazU1pCukCFj3vA9mNI5rWs5gQG3
Id3wGCD1jdm1E5Yxb5ikD5nG67tTW5Pn4+tidsavTNsDLsks/pW/0EcPcKAS+TJ8
Zy15MsGGfHVVkxf+ldULIxxidAeplQhWuED6wkbuD3LQi6Kt4yElHS+UCATca8Fe
CvXNcQWrEHiYUvpyrvU3ybw7WEUUWFa/dctSZwmHvkvRD/bwJPf5M8sIIl8zlyuy
3YCIlSkCgYEA/ZqGOnYIK/bA/QVuyFkFkP3aJjOKJtH0RV9V5XVKSBlU1/Lm3DUZ
XVmp7JuWZHVhPxZa8tswj4x15dX+TwTvGdoUuqPC7K/UMOt6Qzk11o0+o2VRYU97
GzYyEDxGEnRqoZsc1922I6nBv8YqsW4WkMRhkFN4JNzLJBVXMTXcDCMCgYEA+Uob
VQfVF+7BfCOCNdSu9dqZoYRCyBm5JNEp5bqF1kiEbGw4FhJYp95Ix5ogD3Ug4aqe
8ylwUK86U2BhfkKmGQ5yf+6VNoTx3EPFaGrODIi82BUraYPyYEN10ZrR8Czy5X9g
1WC+WuboRgvTZs+grwnDVJwqQIOqIB2L0p+SdR0CgYEAokHavc7E/bP72CdAsSjb
+d+hUq3JJ3tPiY8suwnnQ+gJM72y3ZOPrf1vTfZiK9Y6KQ4ZlKaPFFkvGaVn95DV
ljnE54FddugsoDwZVqdk/egS+qIZhmQ/BLMRJvgZcTdQ/iLrOmYdYgX788JLkIg6
Ide0AI6XISavRl/tEIxARPcCgYEAlgh+6K8dFhlRA7iPPnyxjDAzdF0YoDuzDTCB
icy3jh747BQ5sTb7epSyssbU8tiooIjCv1A6U6UScmm4Y3gTZVMnoE1kKnra4Zk8
LzrQpgSJu3cKOKf78OnI+Ay4u1ciHPOLwQBHsIf2VWn6oo7lg1NZ5wtR9qAHfOqr
Y2k8iRUCgYBKQCtY4SNDuFb6+r5YSEFVfelCn6DJzNgTxO2mkUzzM7RcgejHbd+i
oqgnYXsFLJgm+NpN1eFpbs2RgAe8Zd4pKQNwJFJf0EbEP57sW3kujgFFEsPYJPOp
n8wFU32yrKgrVCftmCk1iI+WPfr1r9LKgKhb0sRX1+DsdWqfN6J7Sw==
-----END RSA PRIVATE KEY-----`

	// The name of the public key on DO
	name := fmt.Sprintf("packer-%s", hex.EncodeToString(identifier.NewUUID().Raw()))

	// Create the key!
	keyId, err := client.CreateKey(name, pub_str)

	if err != nil {
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// We use this to check cleanup
	s.keyId = keyId

	log.Printf("temporary ssh key name: %s", name)

	// Remember some state for the future
	state["ssh_key_id"] = keyId

	return multistep.ActionContinue
}

func (s *stepCreateSSHKey) Cleanup(state map[string]interface{}) {
	// If no key name is set, then we never created it, so just return
	if s.keyId == 0 {
		return
	}

	client := state["client"].(*DigitalOceanClient)
	ui := state["ui"].(packer.Ui)

	ui.Say("Deleting temporary ssh key...")
	err := client.DestroyKey(s.keyId)
	if err != nil {
		log.Printf("Error cleaning up ssh key: %v", err.Error())
		ui.Error(fmt.Sprintf(
			"Error cleaning up ssh key. Please delete the key manually: %v", s.keyId))
	}
}
