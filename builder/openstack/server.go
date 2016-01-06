package openstack

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/mitchellh/multistep"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
)

// StateRefreshFunc is a function type used for StateChangeConf that is
// responsible for refreshing the item being watched for a state change.
//
// It returns three results. `result` is any object that will be returned
// as the final object after waiting for state change. This allows you to
// return the final updated object, for example an openstack instance after
// refreshing it.
//
// `state` is the latest state of that object. And `err` is any error that
// may have happened while refreshing the state.
type StateRefreshFunc func() (result interface{}, state string, progress int, err error)

// StateChangeConf is the configuration struct used for `WaitForState`.
type StateChangeConf struct {
	Pending   []string
	Refresh   StateRefreshFunc
	StepState multistep.StateBag
	Target    []string
}

// ServerStateRefreshFunc returns a StateRefreshFunc that is used to watch
// an openstack server.
func ServerStateRefreshFunc(
	client *gophercloud.ServiceClient, s *servers.Server) StateRefreshFunc {
	return func() (interface{}, string, int, error) {
		serverNew, err := servers.Get(client, s.ID).Extract()
		if err != nil {
			errCode, ok := err.(*gophercloud.UnexpectedResponseCodeError)
			if ok && errCode.Actual == 404 {
				log.Printf("[INFO] 404 on ServerStateRefresh, returning DELETED")
				return nil, "DELETED", 0, nil
			} else {
				log.Printf("[ERROR] Error on ServerStateRefresh: %s", err)
				return nil, "", 0, err
			}
		}

		return serverNew, serverNew.Status, serverNew.Progress, nil
	}
}

// WaitForState watches an object and waits for it to achieve a certain
// state.
func WaitForState(conf *StateChangeConf) (i interface{}, err error) {
	log.Printf("Waiting for state to become: %s", conf.Target)

	for {
		var currentProgress int
		var currentState string
		i, currentState, currentProgress, err = conf.Refresh()
		if err != nil {
			return
		}

		for _, t := range conf.Target {
			if currentState == t {
				return
			}
		}

		if conf.StepState != nil {
			if _, ok := conf.StepState.GetOk(multistep.StateCancelled); ok {
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
			return nil, fmt.Errorf("unexpected state '%s', wanted target '%s'", currentState, conf.Target)
		}

		log.Printf("Waiting for state to become: %s currently %s (%d%%)", conf.Target, currentState, currentProgress)
		time.Sleep(2 * time.Second)
	}
}
