package ebs

import (
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"log"
	"time"
)

func waitForState(ec2conn *ec2.EC2, originalInstance *ec2.Instance, pending []string, target string) (i *ec2.Instance, err error) {
	log.Printf("Waiting for instance state to become: %s", target)

	i = originalInstance
	for i.State.Name != target {
		found := false
		for _, allowed := range pending {
			if i.State.Name == allowed {
				found = true
				break
			}
		}

		if !found {
			fmt.Errorf("unexpected state '%s', wanted target '%s'", i.State.Name, target)
			return
		}

		var resp *ec2.InstancesResp
		resp, err = ec2conn.Instances([]string{i.InstanceId}, ec2.NewFilter())
		if err != nil {
			return
		}

		i = &resp.Reservations[0].Instances[0]
		time.Sleep(2 * time.Second)
	}

	return
}
