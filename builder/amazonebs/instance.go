package amazonebs

import (
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"log"
	"time"
)

func waitForState(ec2conn *ec2.EC2, originalInstance *ec2.Instance, target string) (i *ec2.Instance, err error) {
	log.Printf("Waiting for instance state to become: %s", target)

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
