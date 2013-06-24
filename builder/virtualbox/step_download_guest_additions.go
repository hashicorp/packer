package virtualbox

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/common"
	"github.com/mitchellh/packer/packer"
	"log"
	"time"
)

// This step uploads a file containing the VirtualBox version, which
// can be useful for various provisioning reasons.
//
// Produces:
//   guest_additions_path string - Path to the guest additions.
type stepDownloadGuestAdditions struct{}

func (s *stepDownloadGuestAdditions) Run(state map[string]interface{}) multistep.StepAction {
	cache := state["cache"].(packer.Cache)
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)

	version, err := driver.Version()
	if err != nil {
		state["error"] = fmt.Errorf("Error reading version for guest additions download: %s", err)
		return multistep.ActionHalt
	}

	url := fmt.Sprintf(
		"http://download.virtualbox.org/virtualbox/%s/VBoxGuestAdditions_%s.iso",
		version, version)
	log.Printf("Guest additions URL: %s", url)

	log.Printf("Acquiring lock to download the guest additions ISO.")
	cachePath := cache.Lock(url)
	defer cache.Unlock(url)

	downloadConfig := &common.DownloadConfig{
		Url:        url,
		TargetPath: cachePath,
		Hash:       nil,
	}

	download := common.NewDownloadClient(downloadConfig)

	downloadCompleteCh := make(chan error, 1)
	go func() {
		ui.Say("Downloading VirtualBox guest additions. Progress will be shown periodically.")
		cachePath, err = download.Get()
		downloadCompleteCh <- err
	}()

	progressTicker := time.NewTicker(5 * time.Second)
	defer progressTicker.Stop()

DownloadWaitLoop:
	for {
		select {
		case err := <-downloadCompleteCh:
			if err != nil {
				state["error"] = fmt.Errorf("Error downloading guest additions: %s", err)
				return multistep.ActionHalt
			}

			break DownloadWaitLoop
		case <-progressTicker.C:
			ui.Message(fmt.Sprintf("Download progress: %d%%", download.PercentProgress()))
		case <-time.After(1 * time.Second):
			if _, ok := state[multistep.StateCancelled]; ok {
				ui.Say("Interrupt received. Cancelling download...")
				return multistep.ActionHalt
			}
		}
	}

	state["guest_additions_path"] = cachePath

	return multistep.ActionContinue
}

func (s *stepDownloadGuestAdditions) Cleanup(state map[string]interface{}) {}
