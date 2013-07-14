package virtualbox

import (
	"encoding/hex"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/common"
	"github.com/mitchellh/packer/packer"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

// This step downloads the ISO specified.
//
// Uses:
//   cache packer.Cache
//   config *config
//   ui     packer.Ui
//
// Produces:
//   iso_path string
type stepDownloadISO struct {
	isoCopyDir string
}

func (s *stepDownloadISO) Run(state map[string]interface{}) multistep.StepAction {
	cache := state["cache"].(packer.Cache)
	config := state["config"].(*config)
	ui := state["ui"].(packer.Ui)

	checksum, err := hex.DecodeString(config.ISOChecksum)
	if err != nil {
		state["error"] = fmt.Errorf("Error parsing checksum: %s", err)
		return multistep.ActionHalt
	}

	log.Printf("Acquiring lock to download the ISO.")
	cachePath := cache.Lock(config.ISOUrl)
	defer cache.Unlock(config.ISOUrl)

	downloadConfig := &common.DownloadConfig{
		Url:        config.ISOUrl,
		TargetPath: cachePath,
		CopyFile:   false,
		Hash:       common.HashForType(config.ISOChecksumType),
		Checksum:   checksum,
	}

	download := common.NewDownloadClient(downloadConfig)

	downloadCompleteCh := make(chan error, 1)
	go func() {
		ui.Say("Copying or downloading ISO. Progress will be reported periodically.")
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
				err := fmt.Errorf("Error downloading ISO: %s", err)
				state["error"] = err
				ui.Error(err.Error())
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

	// VirtualBox is really dumb and can't figure out that the file is an
	// ISO unless it has a ".iso" extension. We can't modify the cache
	// filenames so we just do a copy.
	tempdir, err := ioutil.TempDir("", "packer")
	if err != nil {
		state["error"] = fmt.Errorf("Error copying ISO: %s", err)
		return multistep.ActionHalt
	}
	s.isoCopyDir = tempdir

	f, err := os.Create(filepath.Join(tempdir, "image.iso"))
	if err != nil {
		state["error"] = fmt.Errorf("Error copying ISO: %s", err)
		return multistep.ActionHalt
	}
	defer f.Close()

	sourceF, err := os.Open(cachePath)
	if err != nil {
		state["error"] = fmt.Errorf("Error copying ISO: %s", err)
		return multistep.ActionHalt
	}
	defer sourceF.Close()

	log.Printf("Copying ISO to temp location: %s", tempdir)
	if _, err := io.Copy(f, sourceF); err != nil {
		state["error"] = fmt.Errorf("Error copying ISO: %s", err)
		return multistep.ActionHalt
	}

	log.Printf("Path to ISO on disk: %s", cachePath)
	state["iso_path"] = f.Name()

	return multistep.ActionContinue
}

func (s *stepDownloadISO) Cleanup(map[string]interface{}) {
	if s.isoCopyDir != "" {
		os.RemoveAll(s.isoCopyDir)
	}
}
