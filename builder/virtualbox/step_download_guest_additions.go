package virtualbox

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/common"
	"github.com/mitchellh/packer/packer"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

var additionsVersionMap = map[string]string{
	"4.2.1":  "4.2.0",
	"4.1.23": "4.1.22",
}

// This step uploads a file containing the VirtualBox version, which
// can be useful for various provisioning reasons.
//
// Produces:
//   guest_additions_path string - Path to the guest additions.
type stepDownloadGuestAdditions struct{}

func (s *stepDownloadGuestAdditions) Run(state map[string]interface{}) multistep.StepAction {
	var action multistep.StepAction
	cache := state["cache"].(packer.Cache)
	driver := state["driver"].(Driver)
	ui := state["ui"].(packer.Ui)
	config := state["config"].(*config)

	// Get VBox version
	version, err := driver.Version()
	if err != nil {
		state["error"] = fmt.Errorf("Error reading version for guest additions download: %s", err)
		return multistep.ActionHalt
	}

	if newVersion, ok := additionsVersionMap[version]; ok {
		log.Printf("Rewriting guest additions version: %s to %s", version, newVersion)
		version = newVersion
	}

	additionsName := fmt.Sprintf("VBoxGuestAdditions_%s.iso", version)

	// Use provided version or get it from virtualbox.org
	var checksum string

	if config.GuestAdditionsSHA256 != "" {
		checksum = config.GuestAdditionsSHA256
	} else {
		checksum, action = s.downloadAdditionsSHA256(state, version, additionsName)
		if action != multistep.ActionContinue {
			return action
		}
	}

	checksumBytes, err := hex.DecodeString(checksum)
	if err != nil {
		state["error"] = fmt.Errorf("Couldn't decode checksum into bytes: %s", checksum)
		return multistep.ActionHalt
	}

	// Use the provided source (URL or file path) or generate it
	url := config.GuestAdditionsURL
	if url == "" {
		url = fmt.Sprintf(
			"http://download.virtualbox.org/virtualbox/%s/%s",
			version,
			additionsName)
	}

	log.Printf("Guest additions URL: %s", url)

	log.Printf("Acquiring lock to download the guest additions ISO.")
	cachePath := cache.Lock(url)
	defer cache.Unlock(url)

	downloadConfig := &common.DownloadConfig{
		Url:        url,
		TargetPath: cachePath,
		Hash:       sha256.New(),
		Checksum:   checksumBytes,
	}

	download := common.NewDownloadClient(downloadConfig)
	ui.Say("Downloading VirtualBox guest additions. Progress will be shown periodically.")
	state["guest_additions_path"], action = s.progressDownload(download, state)
	return action
}

func (s *stepDownloadGuestAdditions) Cleanup(state map[string]interface{}) {}

func (s *stepDownloadGuestAdditions) progressDownload(c *common.DownloadClient, state map[string]interface{}) (string, multistep.StepAction) {
	ui := state["ui"].(packer.Ui)

	var result string
	downloadCompleteCh := make(chan error, 1)

	// Start a goroutine to actually do the download...
	go func() {
		var err error
		result, err = c.Get()
		downloadCompleteCh <- err
	}()

	progressTicker := time.NewTicker(5 * time.Second)
	defer progressTicker.Stop()

	// A loop that handles showing progress as well as timing out and handling
	// interrupts and all that.
DownloadWaitLoop:
	for {
		select {
		case err := <-downloadCompleteCh:
			if err != nil {
				state["error"] = fmt.Errorf("Error downloading: %s", err)
				return "", multistep.ActionHalt
			}

			break DownloadWaitLoop
		case <-progressTicker.C:
			ui.Message(fmt.Sprintf("Download progress: %d%%", c.PercentProgress()))
		case <-time.After(1 * time.Second):
			if _, ok := state[multistep.StateCancelled]; ok {
				ui.Say("Interrupt received. Cancelling download...")
				return "", multistep.ActionHalt
			}
		}
	}

	return result, multistep.ActionContinue
}

func (s *stepDownloadGuestAdditions) downloadAdditionsSHA256(state map[string]interface{}, additionsVersion string, additionsName string) (string, multistep.StepAction) {
	// First things first, we get the list of checksums for the files available
	// for this version.
	checksumsUrl := fmt.Sprintf("http://download.virtualbox.org/virtualbox/%s/SHA256SUMS", additionsVersion)

	checksumsFile, err := ioutil.TempFile("", "packer")
	if err != nil {
		state["error"] = fmt.Errorf(
			"Failed creating temporary file to store guest addition checksums: %s",
			err)
		return "", multistep.ActionHalt
	}
	defer os.Remove(checksumsFile.Name())

	checksumsFile.Close()

	downloadConfig := &common.DownloadConfig{
		Url:        checksumsUrl,
		TargetPath: checksumsFile.Name(),
		Hash:       nil,
	}

	log.Printf("Downloading guest addition checksums: %s", checksumsUrl)
	download := common.NewDownloadClient(downloadConfig)
	checksumsPath, action := s.progressDownload(download, state)
	if action != multistep.ActionContinue {
		return "", action
	}

	// Next, we find the checksum for the file we're looking to download.
	// It is an error if the checksum cannot be found.
	checksumsF, err := os.Open(checksumsPath)
	if err != nil {
		state["error"] = fmt.Errorf("Error opening guest addition checksums: %s", err)
		return "", multistep.ActionHalt
	}
	defer checksumsF.Close()

	// We copy the contents of the file into memory. In general this file
	// is quite small so that is okay. In the future, we probably want to
	// use bufio and iterate line by line.
	var contents bytes.Buffer
	io.Copy(&contents, checksumsF)

	checksum := ""
	for _, line := range strings.Split(contents.String(), "\n") {
		parts := strings.Fields(line)
		log.Printf("Checksum file parts: %#v", parts)
		if len(parts) != 2 {
			// Bogus line
			continue
		}

		if strings.HasSuffix(parts[1], additionsName) {
			checksum = parts[0]
			log.Printf("Guest additions checksum: %s", checksum)
			break
		}
	}

	if checksum == "" {
		state["error"] = fmt.Errorf("The checksum for the file '%s' could not be found.", additionsName)
		return "", multistep.ActionHalt
	}

	return checksum, multistep.ActionContinue

}
