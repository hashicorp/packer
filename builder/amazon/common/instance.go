package common

import (
	"errors"
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"log"
	"time"
)

type StateChangeConf struct {
	Conn      *ec2.EC2
	Instance  *ec2.Instance
	Pending   []string
	StepState map[string]interface{}
	Target    string
}

func WaitForState(conf *StateChangeConf) (i *ec2.Instance, err error) {
	log.Printf("Waiting for instance state to become: %s", conf.Target)

	i = conf.Instance
	for i.State.Name != conf.Target {
		if conf.StepState != nil {
			if _, ok := conf.StepState[multistep.StateCancelled]; ok {
				return nil, errors.New("interrupted")
			}
		}

		found := false
		for _, allowed := range conf.Pending {
			if i.State.Name == allowed {
				found = true
				break
			}
		}

		if !found {
			fmt.Errorf("unexpected state '%s', wanted target '%s'", i.State.Name, conf.Target)
			return
		}

		var resp *ec2.InstancesResp
		resp, err = conf.Conn.Instances([]string{i.InstanceId}, ec2.NewFilter())
		if err != nil {
			return
		}

		i = &resp.Reservations[0].Instances[0]
		time.Sleep(2 * time.Second)
	}

	return
}
