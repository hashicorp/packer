package iso

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	neturl "net/url"

	vmwcommon "github.com/hashicorp/packer/builder/vmware/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/packer"
	"github.com/mitchellh/multistep"
	"runtime"
)

type stepDownload struct {
	step *common.StepDownload
}

func (s *stepDownload) Run(state multistep.StateBag) multistep.StepAction {
	cache := state.Get("cache").(packer.Cache)
	driver := state.Get("driver").(vmwcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	if esx5, ok := driver.(*ESX5Driver); ok {
		ui.Say("Verifying remote cache")

		targetPath := ""
		for _, url := range s.step.Url {
			targetPath = s.step.TargetPath

			if targetPath == "" {
				if u, err := neturl.Parse(url); err == nil {

					if u.Scheme == "file" {

						if u.Path != "" {
							targetPath = u.Path
						} else if u.Opaque != "" {
							targetPath = u.Opaque
						}

						if runtime.GOOS == "windows" && len(targetPath) > 0 && targetPath[0] == '/' {
							targetPath = targetPath[1:]
						}
					}
				}

				if targetPath == "" {
					hash := sha1.Sum([]byte(url))
					cacheKey := fmt.Sprintf("%s.%s", hex.EncodeToString(hash[:]), s.step.Extension)
					targetPath = cache.Lock(cacheKey)
					cache.Unlock(cacheKey)
				}
			}

			remotePath := esx5.cachePath(targetPath)
			ui.Message(remotePath)
			if esx5.verifyChecksum(s.step.ChecksumType, s.step.Checksum, remotePath) {
				state.Put(s.step.ResultKey, "skip_upload:"+remotePath)
				ui.Message("Remote cache verified, skipping download step")
				return multistep.ActionContinue
			}

			ui.Message("Remote cache couldn't be verified")
		}
	}

	return s.step.Run(state)
}

func (s *stepDownload) Cleanup(multistep.StateBag) {}
