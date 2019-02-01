package common

import (
	"fmt"
	"log"

	"github.com/hashicorp/packer/common"
	"github.com/outscale/osc-go/oapi"
)

type stateRefreshFunc func() (string, error)

func waitForSecurityGroup(conn *oapi.Client, securityGroupID string) error {
	errCh := make(chan error, 1)
	go waitForState(errCh, "exists", securityGroupWaitFunc(conn, securityGroupID))
	err := <-errCh
	return err
}

func waitForState(errCh chan<- error, target string, refresh stateRefreshFunc) error {
	err := common.Retry(2, 2, 0, func(_ uint) (bool, error) {
		state, err := refresh()
		if err != nil {
			return false, err
		} else if state == target {
			return true, nil
		}
		return false, nil
	})
	errCh <- err
	return err
}

func securityGroupWaitFunc(conn *oapi.Client, id string) stateRefreshFunc {
	return func() (string, error) {
		log.Printf("[Debug] Check if SG with id %s exists", id)
		resp, err := conn.POST_ReadSecurityGroups(oapi.ReadSecurityGroupsRequest{
			Filters: oapi.FiltersSecurityGroup{
				SecurityGroupIds: []string{id},
			},
		})

		log.Printf("[Debug] Read Response %+v", resp.OK)

		if err != nil {
			return "", err
		}

		if resp.OK == nil {
			return "", fmt.Errorf("Security Group with ID %s. Not Found", id)
		}

		if len(resp.OK.SecurityGroups) == 0 {
			return "waiting", nil
		}

		return "exists", nil
	}
}
