package common

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/helper/multistep"
)

const (
	privateIP  = "10.0.0.1"
	publicIP   = "192.168.1.1"
	privateDNS = "private.dns.test"
	publicDNS  = "public.dns.test"
)

func TestSSHHost(t *testing.T) {
	origSshHostSleepDuration := sshHostSleepDuration
	defer func() { sshHostSleepDuration = origSshHostSleepDuration }()
	sshHostSleepDuration = 0

	var cases = []struct {
		allowTries   int
		vpcId        string
		sshInterface string

		ok       bool
		wantHost string
	}{
		{1, "", "", true, publicDNS},
		{1, "", "private_ip", true, privateIP},
		{1, "vpc-id", "", true, publicIP},
		{1, "vpc-id", "private_ip", true, privateIP},
		{1, "vpc-id", "private_dns", true, privateDNS},
		{1, "vpc-id", "public_dns", true, publicDNS},
		{1, "vpc-id", "public_ip", true, publicIP},
		{2, "", "", true, publicDNS},
		{2, "", "private_ip", true, privateIP},
		{2, "vpc-id", "", true, publicIP},
		{2, "vpc-id", "private_ip", true, privateIP},
		{2, "vpc-id", "private_dns", true, privateDNS},
		{2, "vpc-id", "public_dns", true, publicDNS},
		{2, "vpc-id", "public_ip", true, publicIP},
		{3, "", "", false, ""},
		{3, "", "private_ip", false, ""},
		{3, "vpc-id", "", false, ""},
		{3, "vpc-id", "private_ip", false, ""},
		{3, "vpc-id", "private_dns", false, ""},
		{3, "vpc-id", "public_dns", false, ""},
		{3, "vpc-id", "public_ip", false, ""},
	}

	for _, c := range cases {
		testSSHHost(t, c.allowTries, c.vpcId, c.sshInterface, c.ok, c.wantHost)
	}
}

func testSSHHost(t *testing.T, allowTries int, vpcId string, sshInterface string, ok bool, wantHost string) {
	t.Logf("allowTries=%d vpcId=%s sshInterface=%s ok=%t wantHost=%q", allowTries, vpcId, sshInterface, ok, wantHost)

	e := &fakeEC2Describer{
		allowTries: allowTries,
		vpcId:      vpcId,
		privateIP:  privateIP,
		publicIP:   publicIP,
		privateDNS: privateDNS,
		publicDNS:  publicDNS,
	}

	f := SSHHost(e, sshInterface)
	st := &multistep.BasicStateBag{}
	st.Put("instance", &ec2.Instance{
		InstanceId: aws.String("instance-id"),
	})

	host, err := f(st)

	if e.tries > allowTries {
		t.Fatalf("got %d ec2 DescribeInstances tries, want %d", e.tries, allowTries)
	}

	switch {
	case ok && err != nil:
		t.Fatalf("expected no error, got %+v", err)
	case !ok && err == nil:
		t.Fatalf("expected error, got none and host %s", host)
	}

	if host != wantHost {
		t.Fatalf("got host %s, want %s", host, wantHost)
	}
}

type fakeEC2Describer struct {
	allowTries int
	tries      int

	vpcId                                      string
	privateIP, publicIP, privateDNS, publicDNS string
}

func (d *fakeEC2Describer) DescribeInstances(in *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	d.tries++

	instance := &ec2.Instance{
		InstanceId: aws.String("instance-id"),
	}

	if d.vpcId != "" {
		instance.VpcId = aws.String(d.vpcId)
	}

	if d.tries >= d.allowTries {
		instance.PublicIpAddress = aws.String(d.publicIP)
		instance.PrivateIpAddress = aws.String(d.privateIP)
		instance.PublicDnsName = aws.String(d.publicDNS)
		instance.PrivateDnsName = aws.String(d.privateDNS)
	}

	out := &ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{
			{
				Instances: []*ec2.Instance{instance},
			},
		},
	}

	return out, nil
}
