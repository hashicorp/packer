// The amazonebs package contains a packer.Builder implementation that
// builds AMIs for Amazon EC2.
//
// In general, there are two types of AMIs that can be created: ebs-backed or
// instance-store. This builder _only_ builds ebs-backed images.
package amazonebs

import (
	"bufio"
	"cgl.tideland.biz/identifier"
	gossh "code.google.com/p/go.crypto/ssh"
	"encoding/hex"
	"fmt"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/communicator/ssh"
	"github.com/mitchellh/packer/packer"
	"log"
	"net"
	"time"
)

type config struct {
	// Access information
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`

	// Information for the source instance
	Region       string
	SourceAmi    string `mapstructure:"source_ami"`
	InstanceType string `mapstructure:"instance_type"`

	// Configuration of the resulting AMI
	AMIName string `mapstructure:"ami_name"`
}

type Builder struct {
	config config
}

func (b *Builder) Prepare(raw interface{}) (err error) {
	err = mapstructure.Decode(raw, &b.config)
	if err != nil {
		return
	}

	log.Printf("Config: %+v\n", b.config)

	// TODO: Validate the configuration
	return
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook) {
	auth := aws.Auth{b.config.AccessKey, b.config.SecretKey}
	region := aws.Regions[b.config.Region]
	ec2conn := ec2.New(auth, region)

	// Create a new keypair that we'll use to access the instance.
	keyName := fmt.Sprintf("packer %s", hex.EncodeToString(identifier.NewUUID().Raw()))
	ui.Say("Creating temporary keypair for this instance...\n")
	log.Printf("temporary keypair name: %s\n", keyName)
	keyResp, err := ec2conn.CreateKeyPair(keyName)
	if err != nil {
		ui.Error("%s\n", err.Error())
		return
	}

	// Make sure the keypair is properly deleted when we exit
	defer func() {
		ui.Say("Deleting temporary keypair...\n")
		_, err := ec2conn.DeleteKeyPair(keyName)
		if err != nil {
			ui.Error(
				"Error cleaning up keypair. Please delete the key manually: %s", keyName)
		}
	}()

	runOpts := &ec2.RunInstances{
		KeyName:      keyName,
		ImageId:      b.config.SourceAmi,
		InstanceType: b.config.InstanceType,
		MinCount:     0,
		MaxCount:     0,
	}

	ui.Say("Launching a source AWS instance...\n")
	runResp, err := ec2conn.RunInstances(runOpts)
	if err != nil {
		ui.Error("%s\n", err.Error())
		return
	}

	instance := &runResp.Instances[0]
	log.Printf("instance id: %s\n", instance.InstanceId)

	// Make sure we clean up the instance by terminating it, no matter what
	defer func() {
		// TODO: error handling
		ui.Say("Terminating the source AWS instance...\n")
		ec2conn.TerminateInstances([]string{instance.InstanceId})
	}()

	ui.Say("Waiting for instance to become ready...\n")
	instance, err = waitForState(ec2conn, instance, "running")
	if err != nil {
		ui.Error("%s\n", err.Error())
		return
	}

	// Build the SSH configuration
	keyring := &ssh.SimpleKeychain{}
	err = keyring.AddPEMKey(keyResp.KeyMaterial)
	if err != nil {
		ui.Say("Error setting up SSH config: %s\n", err.Error())
		return
	}

	sshConfig := &gossh.ClientConfig{
		User: "ubuntu",
		Auth: []gossh.ClientAuth{
			gossh.ClientAuthKeyring(keyring),
		},
	}

	// Try to connect for SSH a few times
	var conn net.Conn
	for i := 0; i < 5; i++ {
		time.Sleep(time.Duration(i) * time.Second)

		log.Printf(
			"Opening TCP conn for SSH to %s:22 (attempt %d)",
			instance.DNSName, i+1)
		conn, err = net.Dial("tcp", fmt.Sprintf("%s:22", instance.DNSName))
		if err != nil {
			continue
		}
		defer conn.Close()
	}

	var comm packer.Communicator
	if err == nil {
		comm, err = ssh.New(conn, sshConfig)
	}

	if err != nil {
		ui.Error("Error connecting to SSH: %s\n", err.Error())
		return
	}

	// XXX: TEST
	remote, err := comm.Start("echo foo")
	if err != nil {
		ui.Error("Error: %s", err.Error())
		return
	}

	remote.Wait()

	bufr := bufio.NewReader(remote.Stdout)
	line, _ := bufr.ReadString('\n')
	ui.Say("%s\n", line)

	// Stop the instance so we can create an AMI from it
	_, err = ec2conn.StopInstances(instance.InstanceId)
	if err != nil {
		ui.Error("%s\n", err.Error())
		return
	}

	// Wait for the instance to actual stop
	// TODO: Handle diff source states, i.e. this force state sucks
	ui.Say("Waiting for the instance to stop...\n")
	instance.State.Name = "stopping"
	instance, err = waitForState(ec2conn, instance, "stopped")
	if err != nil {
		ui.Error("%s\n", err.Error())
		return
	}

	// Create the image
	ui.Say("Creating the AMI...\n")
	createOpts := &ec2.CreateImage{
		InstanceId: instance.InstanceId,
		Name:       b.config.AMIName,
	}

	createResp, err := ec2conn.CreateImage(createOpts)
	if err != nil {
		ui.Error("%s\n", err.Error())
		return
	}

	ui.Say("AMI: %s\n", createResp.ImageId)

	// Wait for the image to become ready
	ui.Say("Waiting for AMI to become ready...\n")
	for {
		imageResp, err := ec2conn.Images([]string{createResp.ImageId}, ec2.NewFilter())
		if err != nil {
			ui.Error("%s\n", err.Error())
			return
		}

		if imageResp.Images[0].State == "available" {
			break
		}
	}
}

func waitForState(ec2conn *ec2.EC2, originalInstance *ec2.Instance, target string) (i *ec2.Instance, err error) {
	log.Printf("Waiting for instance state to become: %s\n", target)

	i = originalInstance
	original := i.State.Name
	for i.State.Name == original {
		var resp *ec2.InstancesResp
		resp, err = ec2conn.Instances([]string{i.InstanceId}, ec2.NewFilter())
		if err != nil {
			return
		}

		i = &resp.Reservations[0].Instances[0]

		time.Sleep(2 * time.Second)
	}

	if i.State.Name != target {
		err = fmt.Errorf("unexpected target state '%s', wanted '%s'", i.State.Name, target)
		return
	}

	return
}
