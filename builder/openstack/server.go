package openstack

import (
	"errors"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/rackspace/gophercloud"
	"log"
	"time"
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
	Target    string
}

// ServerStateRefreshFunc returns a StateRefreshFunc that is used to watch
// an openstacn server.
func ServerStateRefreshFunc(csp gophercloud.CloudServersProvider, s *gophercloud.Server) StateRefreshFunc {
	return func() (interface{}, string, int, error) {
		resp, err := csp.ServerById(s.Id)
		if err != nil {
			log.Printf("Error on ServerStateRefresh: %s", err)
			return nil, "", 0, err
		}

		return resp, resp.Status, resp.Progress, nil
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

		if currentState == conf.Target {
			return
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

	return
}
