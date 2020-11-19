//go:generate mapstructure-to-hcl2 -type Config,nicConfig,diskConfig,vgaConfig

package proxmoxiso

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	proxmox "github.com/hashicorp/packer/builder/proxmox/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep/commonsteps"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type Config struct {
	proxmox.Config `mapstructure:",squash"`

	commonsteps.ISOConfig `mapstructure:",squash"`
	ISOFile               string `mapstructure:"iso_file"`
	ISOStoragePool        string `mapstructure:"iso_storage_pool"`
	UnmountISO            bool   `mapstructure:"unmount_iso"`
	shouldUploadISO       bool
}

func (c *Config) Prepare(raws ...interface{}) ([]string, []string, error) {
	var errs *packersdk.MultiError
	_, warnings, merrs := c.Config.Prepare(c, raws...)
	if merrs != nil {
		errs = packersdk.MultiErrorAppend(errs, merrs)
	}

	// Check ISO config
	// Either a pre-uploaded ISO should be referenced in iso_file, OR a URL
	// (possibly to a local file) to an ISO file that will be downloaded and
	// then uploaded to Proxmox.
	if c.ISOFile != "" {
		c.shouldUploadISO = false
	} else {
		isoWarnings, isoErrors := c.ISOConfig.Prepare(&c.Ctx)
		errs = packersdk.MultiErrorAppend(errs, isoErrors...)
		warnings = append(warnings, isoWarnings...)
		c.shouldUploadISO = true
	}

	if (c.ISOFile == "" && len(c.ISOConfig.ISOUrls) == 0) || (c.ISOFile != "" && len(c.ISOConfig.ISOUrls) != 0) {
		errs = packersdk.MultiErrorAppend(errs, errors.New("either iso_file or iso_url, but not both, must be specified"))
	}
	if len(c.ISOConfig.ISOUrls) != 0 && c.ISOStoragePool == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("when specifying iso_url, iso_storage_pool must also be specified"))
	}

	for idx := range c.AdditionalISOFiles {
		// Check AdditionalISO config
		// Either a pre-uploaded ISO should be referenced in iso_file, OR a URL
		// (possibly to a local file) to an ISO file that will be downloaded and
		// then uploaded to Proxmox.
		if c.AdditionalISOFiles[idx].ISOFile != "" {
			c.AdditionalISOFiles[idx].ShouldUploadISO = false
		} else {
			c.AdditionalISOFiles[idx].DownloadPathKey = "downloaded_additional_iso_path_" + strconv.Itoa(idx)
			isoWarnings, isoErrors := c.AdditionalISOFiles[idx].ISOConfig.Prepare(&c.Ctx)
			errs = packersdk.MultiErrorAppend(errs, isoErrors...)
			warnings = append(warnings, isoWarnings...)
			c.AdditionalISOFiles[idx].ShouldUploadISO = true
		}
		if c.AdditionalISOFiles[idx].Device == "" {
			log.Printf("AdditionalISOFile %d Device not set, using default 'ide3'", idx)
			c.AdditionalISOFiles[idx].Device = "ide3"
		}
		if strings.HasPrefix(c.AdditionalISOFiles[idx].Device, "ide") {
			busnumber, err := strconv.Atoi(c.AdditionalISOFiles[idx].Device[3:])
			if err != nil {
				errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("%s is not a valid bus index", c.AdditionalISOFiles[idx].Device[3:]))
			}
			if busnumber == 2 {
				errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("IDE bus 2 is used by boot ISO"))
			}
			if busnumber > 3 {
				errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("IDE bus index can't be higher than 3"))
			}
		}
		if strings.HasPrefix(c.AdditionalISOFiles[idx].Device, "sata") {
			busnumber, err := strconv.Atoi(c.AdditionalISOFiles[idx].Device[4:])
			if err != nil {
				errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("%s is not a valid bus index", c.AdditionalISOFiles[idx].Device[4:]))
			}
			if busnumber > 5 {
				errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("SATA bus index can't be higher than 5"))
			}
		}
		if strings.HasPrefix(c.AdditionalISOFiles[idx].Device, "scsi") {
			busnumber, err := strconv.Atoi(c.AdditionalISOFiles[idx].Device[4:])
			if err != nil {
				errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("%s is not a valid bus index", c.AdditionalISOFiles[idx].Device[4:]))
			}
			if busnumber > 30 {
				errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("SCSI bus index can't be higher than 30"))
			}
		}
		if (c.AdditionalISOFiles[idx].ISOFile == "" && len(c.AdditionalISOFiles[idx].ISOConfig.ISOUrls) == 0) || (c.AdditionalISOFiles[idx].ISOFile != "" && len(c.AdditionalISOFiles[idx].ISOConfig.ISOUrls) != 0) {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("either iso_file or iso_url, but not both, must be specified for AdditionalISO file %s", c.AdditionalISOFiles[idx].Device))
		}
		if len(c.ISOConfig.ISOUrls) != 0 && c.ISOStoragePool == "" {
			errs = packersdk.MultiErrorAppend(errs, errors.New("when specifying iso_url, iso_storage_pool must also be specified"))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}
	return nil, warnings, nil
}
