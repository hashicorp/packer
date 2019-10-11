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

func waitUntilForVmRunning(conn *oapi.Client, vmID string) error {
	errCh := make(chan error, 1)
	go waitForState(errCh, "running", waitUntilVmStateFunc(conn, vmID))
	err := <-errCh
	return err
}

func waitUntilVmDeleted(conn *oapi.Client, vmID string) error {
	errCh := make(chan error, 1)
	go waitForState(errCh, "terminated", waitUntilVmStateFunc(conn, vmID))
	return <-errCh
}

func waitUntilVmStopped(conn *oapi.Client, vmID string) error {
	errCh := make(chan error, 1)
	go waitForState(errCh, "stopped", waitUntilVmStateFunc(conn, vmID))
	return <-errCh
}

func WaitUntilSnapshotCompleted(conn *oapi.Client, id string) error {
	errCh := make(chan error, 1)
	go waitForState(errCh, "completed", waitUntilSnapshotStateFunc(conn, id))
	return <-errCh
}

func WaitUntilImageAvailable(conn *oapi.Client, imageID string) error {
	errCh := make(chan error, 1)
	go waitForState(errCh, "available", waitUntilImageStateFunc(conn, imageID))
	return <-errCh
}

func WaitUntilVolumeAvailable(conn *oapi.Client, volumeID string) error {
	errCh := make(chan error, 1)
	go waitForState(errCh, "available", volumeWaitFunc(conn, volumeID))
	return <-errCh
}

func WaitUntilVolumeIsLinked(conn *oapi.Client, volumeID string) error {
	errCh := make(chan error, 1)
	go waitForState(errCh, "attached", waitUntilVolumeLinkedStateFunc(conn, volumeID))
	return <-errCh
}

func WaitUntilVolumeIsUnlinked(conn *oapi.Client, volumeID string) error {
	errCh := make(chan error, 1)
	go waitForState(errCh, "dettached", waitUntilVolumeUnLinkedStateFunc(conn, volumeID))
	return <-errCh
}

func WaitUntilSnapshotDone(conn *oapi.Client, snapshotID string) error {
	errCh := make(chan error, 1)
	go waitForState(errCh, "completed", waitUntilSnapshotDoneStateFunc(conn, snapshotID))
	return <-errCh
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

func waitUntilVmStateFunc(conn *oapi.Client, id string) stateRefreshFunc {
	return func() (string, error) {
		log.Printf("[Debug] Check if SG with id %s exists", id)
		resp, err := conn.POST_ReadVms(oapi.ReadVmsRequest{
			Filters: oapi.FiltersVm{
				VmIds: []string{id},
			},
		})

		log.Printf("[Debug] Read Response %+v", resp.OK)

		if err != nil {
			return "", err
		}

		if resp.OK == nil {
			return "", fmt.Errorf("Vm with ID %s. Not Found", id)
		}

		if len(resp.OK.Vms) == 0 {
			return "pending", nil
		}

		return resp.OK.Vms[0].State, nil
	}
}

func waitUntilVolumeLinkedStateFunc(conn *oapi.Client, id string) stateRefreshFunc {
	return func() (string, error) {
		log.Printf("[Debug] Check if volume with id %s exists", id)
		resp, err := conn.POST_ReadVolumes(oapi.ReadVolumesRequest{
			Filters: oapi.FiltersVolume{
				VolumeIds: []string{id},
			},
		})

		if err != nil {
			return "", err
		}

		if resp.OK == nil {
			return "", fmt.Errorf("Vm with ID %s. Not Found", id)
		}

		log.Printf("[Debug] Read Response %+v", resp.OK)

		if len(resp.OK.Volumes) == 0 {
			return "pending", nil
		}

		if len(resp.OK.Volumes[0].LinkedVolumes) == 0 {
			return "pending", nil
		}

		return resp.OK.Volumes[0].LinkedVolumes[0].State, nil
	}
}

func waitUntilVolumeUnLinkedStateFunc(conn *oapi.Client, id string) stateRefreshFunc {
	return func() (string, error) {
		log.Printf("[Debug] Check if volume with id %s exists", id)
		resp, err := conn.POST_ReadVolumes(oapi.ReadVolumesRequest{
			Filters: oapi.FiltersVolume{
				VolumeIds: []string{id},
			},
		})

		if err != nil {
			return "", err
		}

		if resp.OK == nil {
			return "", fmt.Errorf("Vm with ID %s. Not Found", id)
		}

		log.Printf("[Debug] Read Response %+v", resp.OK)

		if len(resp.OK.Volumes) == 0 {
			return "pending", nil
		}

		if len(resp.OK.Volumes[0].LinkedVolumes) == 0 {
			return "dettached", nil
		}

		return "failed", nil
	}
}

func waitUntilSnapshotStateFunc(conn *oapi.Client, id string) stateRefreshFunc {
	return func() (string, error) {
		log.Printf("[Debug] Check if Snapshot with id %s exists", id)
		resp, err := conn.POST_ReadSnapshots(oapi.ReadSnapshotsRequest{
			Filters: oapi.FiltersSnapshot{
				SnapshotIds: []string{id},
			},
		})

		log.Printf("[Debug] Read Response %+v", resp.OK)

		if err != nil {
			return "", err
		}

		if resp.OK == nil {
			return "", fmt.Errorf("Vm with ID %s. Not Found", id)
		}

		if len(resp.OK.Snapshots) == 0 {
			return "pending", nil
		}

		return resp.OK.Snapshots[0].State, nil
	}
}

func waitUntilImageStateFunc(conn *oapi.Client, id string) stateRefreshFunc {
	return func() (string, error) {
		log.Printf("[Debug] Check if Image with id %s exists", id)
		resp, err := conn.POST_ReadImages(oapi.ReadImagesRequest{
			Filters: oapi.FiltersImage{
				ImageIds: []string{id},
			},
		})

		log.Printf("[Debug] Read Response %+v", resp.OK)

		if err != nil {
			return "", err
		}

		if resp.OK == nil {
			return "", fmt.Errorf("Vm with ID %s. Not Found", id)
		}

		if len(resp.OK.Images) == 0 {
			return "pending", nil
		}

		if resp.OK.Images[0].State == "failed" {
			return resp.OK.Images[0].State, fmt.Errorf("Image (%s) creation is failed", id)
		}

		return resp.OK.Images[0].State, nil
	}
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

func waitUntilSnapshotDoneStateFunc(conn *oapi.Client, id string) stateRefreshFunc {
	return func() (string, error) {
		log.Printf("[Debug] Check if Snapshot with id %s exists", id)
		resp, err := conn.POST_ReadSnapshots(oapi.ReadSnapshotsRequest{
			Filters: oapi.FiltersSnapshot{
				SnapshotIds: []string{id},
			},
		})

		log.Printf("[Debug] Read Response %+v", resp.OK)

		if err != nil {
			return "", err
		}

		if resp.OK == nil {
			return "", fmt.Errorf("Snapshot with ID %s. Not Found", id)
		}

		if len(resp.OK.Snapshots) == 0 {
			return "", fmt.Errorf("Snapshot with ID %s. Not Found", id)
		}

		if resp.OK.Snapshots[0].State == "error" {
			return resp.OK.Snapshots[0].State, fmt.Errorf("Snapshot (%s) creation is failed", id)
		}

		return resp.OK.Snapshots[0].State, nil
	}
}

func volumeWaitFunc(conn *oapi.Client, id string) stateRefreshFunc {
	return func() (string, error) {
		log.Printf("[Debug] Check if SvolumeG with id %s exists", id)
		resp, err := conn.POST_ReadVolumes(oapi.ReadVolumesRequest{
			Filters: oapi.FiltersVolume{
				VolumeIds: []string{id},
			},
		})

		log.Printf("[Debug] Read Response %+v", resp.OK)

		if err != nil {
			return "", err
		}

		if resp.OK == nil {
			return "", fmt.Errorf("Volume with ID %s. Not Found", id)
		}

		if len(resp.OK.Volumes) == 0 {
			return "waiting", nil
		}

		if resp.OK.Volumes[0].State == "error" {
			return resp.OK.Volumes[0].State, fmt.Errorf("Volume (%s) creation is failed", id)
		}

		return resp.OK.Volumes[0].State, nil
	}
}
