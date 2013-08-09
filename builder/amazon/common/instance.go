package common

import (
	"errors"
	"fmt"
	"github.com/mitchellh/goamz/ec2"
	"github.com/mitchellh/multistep"
	"log"
	"time"
)

// StateRefreshFunc is a function type used for StateChangeConf that is
// responsible for refreshing the item being watched for a state change.
//
// It returns three results. `result` is any object that will be returned
// as the final object after waiting for state change. This allows you to
// return the final updated object, for example an EC2 instance after refreshing
// it.
//
// `state` is the latest state of that object. And `err` is any error that
// may have happened while refreshing the state.
type StateRefreshFunc func() (result interface{}, state string, err error)

// StateChangeConf is the configuration struct used for `WaitForState`.
type StateChangeConf struct {
	Conn      *ec2.EC2
	Pending   []string
	Refresh   StateRefreshFunc
	StepState map[string]interface{}
	Target    string
}

// InstanceStateRefreshFunc returns a StateRefreshFunc that is used to watch
// an EC2 instance.
func InstanceStateRefreshFunc(conn *ec2.EC2, i *ec2.Instance) StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.Instances([]string{i.InstanceId}, ec2.NewFilter())
		if err != nil {
			log.Printf("Error on InstanceStateRefresh: %s", err)
			return nil, "", err
		}

		i = &resp.Reservations[0].Instances[0]
		return i, i.State.Name, nil
	}
}

// WaitForState watches an object and waits for it to achieve a certain
// state.
func WaitForState(conf *StateChangeConf) (i interface{}, err error) {
	log.Printf("Waiting for state to become: %s", conf.Target)

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
