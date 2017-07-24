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

type stepVerifyCache struct {
	download     *common.StepDownload
	remoteUpload *stepRemoteUpload
}

func (s *stepVerifyCache) Run(state multistep.StateBag) multistep.StepAction {
	cache := state.Get("cache").(packer.Cache)
	driver := state.Get("driver").(vmwcommon.Driver)
	ui := state.Get("ui").(packer.Ui)

	if esx5, ok := driver.(*ESX5Driver); ok {
		ui.Say("Verifying remote cache")

		for _, url := range s.download.Url {
			targetPath := s.download.TargetPath

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
				cacheKey := fmt.Sprintf("%s.%s", hex.EncodeToString(hash[:]), s.download.Extension)
				targetPath = cache.Lock(cacheKey)
				cache.Unlock(cacheKey)
			}

			remotePath := esx5.cachePath(targetPath)
			ui.Message(remotePath)

			if esx5.verifyChecksum(s.download.ChecksumType, s.download.Checksum, remotePath) {
				ui.Message("Remote cache verified, skipping download/upload step")

				s.remoteUpload.Skip = true
				state.Put(s.download.ResultKey, remotePath)
				return multistep.ActionContinue
			}

			ui.Message("Remote cache couldn't be verified")
		}
	}

	return s.download.Run(state)
}

func (s *stepVerifyCache) Cleanup(multistep.StateBag) {}
