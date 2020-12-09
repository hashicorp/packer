package common

import (
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

func TestStepRemoteUpload_Cleanup(t *testing.T) {
	state := new(multistep.BasicStateBag)
	driver := new(RemoteDriverMock)
	state.Put("driver", driver)
	state.Put("path_key", "packer_cache")

	// Should clean up cache
	s := &StepRemoteUpload{
		Key:       "path_key",
		DoCleanup: true,
	}
	s.Cleanup(state)

	if !driver.CacheRemoved {
		t.Fatalf("bad: remote cache was not removed")
	}
	if driver.RemovedCachePath != "packer_cache" {
		t.Fatalf("bad: removed cache path was expected to be packer_cache but was %s", driver.RemovedCachePath)
	}

	// Should NOT clean up cache
	s = &StepRemoteUpload{
		Key: "path_key",
	}
	driver = new(RemoteDriverMock)
	state.Put("driver", driver)
	s.Cleanup(state)

	if driver.CacheRemoved {
		t.Fatalf("bad: remote cache was removed but was expected to not")
	}
}
