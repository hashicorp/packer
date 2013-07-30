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
	Pending   []string
	Refresh func() (interface{}, string, error)
	StepState map[string]interface{}
	Target    string
}

func InstanceStateRefreshFunc(conn *ec2.EC2, i *ec2.Instance) func() (interface{}, string, error) {
	return func() (interface{}, string, error) {
		resp, err := conn.Instances([]string{i.InstanceId}, ec2.NewFilter())
		if err != nil {
			return nil, "", err
		}

		i = &resp.Reservations[0].Instances[0]
		return i, i.State.Name, nil
	}
}

func WaitForState(conf *StateChangeConf) (i interface{}, err error) {
	log.Printf("Waiting for instance state to become: %s", conf.Target)

	for {
		var currentState string
		i, currentState, err = conf.Refresh()
		if err != nil {
			return
		}

		if currentState == conf.Target {
			return
		}

		if conf.StepState != nil {
			if _, ok := conf.StepState[multistep.StateCancelled]; ok {
				return nil, errors.New("interrupted")
			}
		}

		found := false
		for _, allowed := range conf.Pending {
			if currentState == allowed {
				found = true
				break
			}
		}

		if !found {
			fmt.Errorf("unexpected state '%s', wanted target '%s'", currentState, conf.Target)
			return
		}

		time.Sleep(2 * time.Second)
	}

	return
}
