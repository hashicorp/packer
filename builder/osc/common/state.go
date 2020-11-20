package common

import (
	"context"
	"fmt"
	"log"

	"github.com/antihax/optional"
	"github.com/hashicorp/packer/builder/osc/common/retry"
	"github.com/outscale/osc-sdk-go/osc"
)

type stateRefreshFunc func() (string, error)

func waitUntilForOscVmRunning(conn *osc.APIClient, vmID string) error {
	errCh := make(chan error, 1)
	go waitForState(errCh, "running", waitUntilOscVmStateFunc(conn, vmID))
	err := <-errCh
	return err
}

func waitUntilOscVmDeleted(conn *osc.APIClient, vmID string) error {
	errCh := make(chan error, 1)
	go waitForState(errCh, "terminated", waitUntilOscVmStateFunc(conn, vmID))
	return <-errCh
}

func waitUntilOscVmStopped(conn *osc.APIClient, vmID string) error {
	errCh := make(chan error, 1)
	go waitForState(errCh, "stopped", waitUntilOscVmStateFunc(conn, vmID))
	return <-errCh
}

func WaitUntilOscSnapshotCompleted(conn *osc.APIClient, id string) error {
	errCh := make(chan error, 1)
	go waitForState(errCh, "completed", waitUntilOscSnapshotStateFunc(conn, id))
	return <-errCh
}

func WaitUntilOscImageAvailable(conn *osc.APIClient, imageID string) error {
	errCh := make(chan error, 1)
	go waitForState(errCh, "available", waitUntilOscImageStateFunc(conn, imageID))
	return <-errCh
}

func WaitUntilOscVolumeAvailable(conn *osc.APIClient, volumeID string) error {
	errCh := make(chan error, 1)
	go waitForState(errCh, "available", volumeOscWaitFunc(conn, volumeID))
	return <-errCh
}

func WaitUntilOscVolumeIsLinked(conn *osc.APIClient, volumeID string) error {
	errCh := make(chan error, 1)
	go waitForState(errCh, "attached", waitUntilOscVolumeLinkedStateFunc(conn, volumeID))
	return <-errCh
}

func WaitUntilOscVolumeIsUnlinked(conn *osc.APIClient, volumeID string) error {
	errCh := make(chan error, 1)
	go waitForState(errCh, "dettached", waitUntilOscVolumeUnLinkedStateFunc(conn, volumeID))
	return <-errCh
}

func WaitUntilOscSnapshotDone(conn *osc.APIClient, snapshotID string) error {
	errCh := make(chan error, 1)
	go waitForState(errCh, "completed", waitUntilOscSnapshotDoneStateFunc(conn, snapshotID))
	return <-errCh
}

func waitForState(errCh chan<- error, target string, refresh stateRefreshFunc) {
	err := retry.Run(2, 2, 0, func(_ uint) (bool, error) {
		state, err := refresh()
		if err != nil {
			return false, err
		} else if state == target {
			return true, nil
		}
		return false, nil
	})
	errCh <- err
}

func waitUntilOscVmStateFunc(conn *osc.APIClient, id string) stateRefreshFunc {
	return func() (string, error) {
		log.Printf("[Debug] Retrieving state for VM with id %s", id)
		resp, _, err := conn.VmApi.ReadVms(context.Background(), &osc.ReadVmsOpts{
			ReadVmsRequest: optional.NewInterface(osc.ReadVmsRequest{
				Filters: osc.FiltersVm{
					VmIds: []string{id},
				},
			}),
		})

		if err != nil {
			return "", err
		}

		//TODO: check if needed
		// if resp == nil {
		// 	return "", fmt.Errorf("Vm with ID %s not Found", id)
		// }

		if len(resp.Vms) == 0 {
			return "pending", nil
		}

		return resp.Vms[0].State, nil
	}
}

func waitUntilOscVolumeLinkedStateFunc(conn *osc.APIClient, id string) stateRefreshFunc {
	return func() (string, error) {
		log.Printf("[Debug] Check if volume with id %s exists", id)
		resp, _, err := conn.VolumeApi.ReadVolumes(context.Background(), &osc.ReadVolumesOpts{
			ReadVolumesRequest: optional.NewInterface(osc.ReadVolumesRequest{
				Filters: osc.FiltersVolume{
					VolumeIds: []string{id},
				},
			}),
		})

		if err != nil {
			return "", err
		}

		if len(resp.Volumes) == 0 {
			return "pending", nil
		}

		if len(resp.Volumes[0].LinkedVolumes) == 0 {
			return "pending", nil
		}

		return resp.Volumes[0].LinkedVolumes[0].State, nil
	}
}

func waitUntilOscVolumeUnLinkedStateFunc(conn *osc.APIClient, id string) stateRefreshFunc {
	return func() (string, error) {
		log.Printf("[Debug] Check if volume with id %s exists", id)
		resp, _, err := conn.VolumeApi.ReadVolumes(context.Background(), &osc.ReadVolumesOpts{
			ReadVolumesRequest: optional.NewInterface(osc.ReadVolumesRequest{
				Filters: osc.FiltersVolume{
					VolumeIds: []string{id},
				},
			}),
		})

		if err != nil {
			return "", err
		}

		if len(resp.Volumes) == 0 {
			return "pending", nil
		}

		if len(resp.Volumes[0].LinkedVolumes) == 0 {
			return "dettached", nil
		}

		return "failed", nil
	}
}

func waitUntilOscSnapshotStateFunc(conn *osc.APIClient, id string) stateRefreshFunc {
	return func() (string, error) {
		log.Printf("[Debug] Check if Snapshot with id %s exists", id)
		resp, _, err := conn.SnapshotApi.ReadSnapshots(context.Background(), &osc.ReadSnapshotsOpts{
			ReadSnapshotsRequest: optional.NewInterface(osc.ReadSnapshotsRequest{
				Filters: osc.FiltersSnapshot{
					SnapshotIds: []string{id},
				},
			}),
		})

		if err != nil {
			return "", err
		}

		if len(resp.Snapshots) == 0 {
			return "pending", nil
		}

		return resp.Snapshots[0].State, nil
	}
}

func waitUntilOscImageStateFunc(conn *osc.APIClient, id string) stateRefreshFunc {
	return func() (string, error) {
		log.Printf("[Debug] Check if Image with id %s exists", id)
		resp, _, err := conn.ImageApi.ReadImages(context.Background(), &osc.ReadImagesOpts{
			ReadImagesRequest: optional.NewInterface(osc.ReadImagesRequest{
				Filters: osc.FiltersImage{
					ImageIds: []string{id},
				},
			}),
		})

		if err != nil {
			return "", err
		}

		if len(resp.Images) == 0 {
			return "pending", nil
		}

		if resp.Images[0].State == "failed" {
			return resp.Images[0].State, fmt.Errorf("Image (%s) creation is failed", id)
		}

		return resp.Images[0].State, nil
	}
}

func waitUntilOscSnapshotDoneStateFunc(conn *osc.APIClient, id string) stateRefreshFunc {
	return func() (string, error) {
		log.Printf("[Debug] Check if Snapshot with id %s exists", id)
		resp, _, err := conn.SnapshotApi.ReadSnapshots(context.Background(), &osc.ReadSnapshotsOpts{
			ReadSnapshotsRequest: optional.NewInterface(osc.ReadSnapshotsRequest{
				Filters: osc.FiltersSnapshot{
					SnapshotIds: []string{id},
				},
			}),
		})

		if err != nil {
			return "", err
		}

		if len(resp.Snapshots) == 0 {
			return "", fmt.Errorf("Snapshot with ID %s. Not Found", id)
		}

		if resp.Snapshots[0].State == "error" {
			return resp.Snapshots[0].State, fmt.Errorf("Snapshot (%s) creation is failed", id)
		}

		return resp.Snapshots[0].State, nil
	}
}

func volumeOscWaitFunc(conn *osc.APIClient, id string) stateRefreshFunc {
	return func() (string, error) {
		log.Printf("[Debug] Check if SvolumeG with id %s exists", id)
		resp, _, err := conn.VolumeApi.ReadVolumes(context.Background(), &osc.ReadVolumesOpts{
			ReadVolumesRequest: optional.NewInterface(osc.ReadVolumesRequest{
				Filters: osc.FiltersVolume{
					VolumeIds: []string{id},
				},
			}),
		})

		if err != nil {
			return "", err
		}

		if len(resp.Volumes) == 0 {
			return "waiting", nil
		}

		if resp.Volumes[0].State == "error" {
			return resp.Volumes[0].State, fmt.Errorf("Volume (%s) creation is failed", id)
		}

		return resp.Volumes[0].State, nil
	}
}
