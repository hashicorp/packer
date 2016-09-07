package common

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
)

const (
	privateIP = "10.0.0.1"
	publicIP  = "192.168.1.1"
	publicDNS = "public.dns.test"
)

func TestSSHHost(t *testing.T) {
	origSshHostSleepDuration := sshHostSleepDuration
	defer func() { sshHostSleepDuration = origSshHostSleepDuration }()
	sshHostSleepDuration = 0

	var cases = []struct {
		allowTries int
		vpcId      string
		private    bool

		ok       bool
		wantHost string
	}{
		{1, "", false, true, publicDNS},
		{1, "", true, true, publicDNS},
		{1, "vpc-id", false, true, publicIP},
		{1, "vpc-id", true, true, privateIP},
		{2, "", false, true, publicDNS},
		{2, "", true, true, publicDNS},
		{2, "vpc-id", false, true, publicIP},
		{2, "vpc-id", true, true, privateIP},
		{3, "", false, false, ""},
		{3, "", true, false, ""},
		{3, "vpc-id", false, false, ""},
		{3, "vpc-id", true, false, ""},
	}

	for _, c := range cases {
		testSSHHost(t, c.allowTries, c.vpcId, c.private, c.ok, c.wantHost)
	}
}

func testSSHHost(t *testing.T, allowTries int, vpcId string, private, ok bool, wantHost string) {
	t.Logf("allowTries=%d vpcId=%s private=%t ok=%t wantHost=%q", allowTries, vpcId, private, ok, wantHost)

	e := &fakeEC2Describer{
		allowTries: allowTries,
		vpcId:      vpcId,
		privateIP:  privateIP,
		publicIP:   publicIP,
		publicDNS:  publicDNS,
	}

	f := SSHHost(e, private)
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

	vpcId                          string
	privateIP, publicIP, publicDNS string
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
