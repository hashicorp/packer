//go:generate mapstructure-to-hcl2 -type Config,nicConfig,diskConfig

package proxmox

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/mitchellh/mapstructure"
)

type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	common.HTTPConfig      `mapstructure:",squash"`
	bootcommand.BootConfig `mapstructure:",squash"`
	BootKeyInterval        time.Duration       `mapstructure:"boot_key_interval"`
	Comm                   communicator.Config `mapstructure:",squash"`

	ProxmoxURLRaw      string `mapstructure:"proxmox_url"`
	proxmoxURL         *url.URL
	SkipCertValidation bool   `mapstructure:"insecure_skip_tls_verify"`
	Username           string `mapstructure:"username"`
	Password           string `mapstructure:"password"`
	Node               string `mapstructure:"node"`
	Pool               string `mapstructure:"pool"`

	VMName string `mapstructure:"vm_name"`
	VMID   int    `mapstructure:"vm_id"`

	Memory         int          `mapstructure:"memory"`
	Cores          int          `mapstructure:"cores"`
	CPUType        string       `mapstructure:"cpu_type"`
	Sockets        int          `mapstructure:"sockets"`
	OS             string       `mapstructure:"os"`
	NICs           []nicConfig  `mapstructure:"network_adapters"`
	Disks          []diskConfig `mapstructure:"disks"`
	ISOFile        string       `mapstructure:"iso_file"`
	Agent          bool         `mapstructure:"qemu_agent"`
	SCSIController string       `mapstructure:"scsi_controller"`

	TemplateName        string `mapstructure:"template_name"`
	TemplateDescription string `mapstructure:"template_description"`
	UnmountISO          bool   `mapstructure:"unmount_iso"`

	ctx interpolate.Context
}

type nicConfig struct {
	Model      string `mapstructure:"model"`
	MACAddress string `mapstructure:"mac_address"`
	Bridge     string `mapstructure:"bridge"`
	VLANTag    string `mapstructure:"vlan_tag"`
}
type diskConfig struct {
	Type            string `mapstructure:"type"`
	StoragePool     string `mapstructure:"storage_pool"`
	StoragePoolType string `mapstructure:"storage_pool_type"`
	Size            string `mapstructure:"disk_size"`
	CacheMode       string `mapstructure:"cache_mode"`
	DiskFormat      string `mapstructure:"format"`
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	// Agent defaults to true
	c.Agent = true

	var md mapstructure.Metadata
	err := config.Decode(c, &config.DecodeOpts{
		Metadata:           &md,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"boot_command",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	var errs *packer.MultiError

	// Defaults
	if c.ProxmoxURLRaw == "" {
		c.ProxmoxURLRaw = os.Getenv("PROXMOX_URL")
	}
	if c.Username == "" {
		c.Username = os.Getenv("PROXMOX_USERNAME")
	}
	if c.Password == "" {
		c.Password = os.Getenv("PROXMOX_PASSWORD")
	}
	if c.BootKeyInterval == 0 && os.Getenv(common.PackerKeyEnv) != "" {
		var err error
		c.BootKeyInterval, err = time.ParseDuration(os.Getenv(common.PackerKeyEnv))
		if err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}
	if c.BootKeyInterval == 0 {
		c.BootKeyInterval = 5 * time.Millisecond
	}

	if c.VMName == "" {
		// Default to packer-[time-ordered-uuid]
		c.VMName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}
	if c.Memory < 16 {
		log.Printf("Memory %d is too small, using default: 512", c.Memory)
		c.Memory = 512
	}
	if c.Cores < 1 {
		log.Printf("Number of cores %d is too small, using default: 1", c.Cores)
		c.Cores = 1
	}
	if c.Sockets < 1 {
		log.Printf("Number of sockets %d is too small, using default: 1", c.Sockets)
		c.Sockets = 1
	}
	if c.CPUType == "" {
		log.Printf("CPU type not set, using default 'kvm64'")
		c.CPUType = "kvm64"
	}
	if c.OS == "" {
		log.Printf("OS not set, using default 'other'")
		c.OS = "other"
	}
	for idx := range c.NICs {
		if c.NICs[idx].Model == "" {
			log.Printf("NIC %d model not set, using default 'e1000'", idx)
			c.NICs[idx].Model = "e1000"
		}
	}
	for idx := range c.Disks {
		if c.Disks[idx].Type == "" {
			log.Printf("Disk %d type not set, using default 'scsi'", idx)
			c.Disks[idx].Type = "scsi"
		}
		if c.Disks[idx].Size == "" {
			log.Printf("Disk %d size not set, using default '20G'", idx)
			c.Disks[idx].Size = "20G"
		}
		if c.Disks[idx].CacheMode == "" {
			log.Printf("Disk %d cache mode not set, using default 'none'", idx)
			c.Disks[idx].CacheMode = "none"
		}
		// For any storage pool types which aren't in rxStorageTypes in proxmox-api/proxmox/config_qemu.go:651
		// (currently zfspool and lvm), the format parameter is mandatory. Make sure this is still up to date
		// when updating the vendored code!
		if !contains([]string{"zfspool", "lvm"}, c.Disks[idx].StoragePoolType) && c.Disks[idx].DiskFormat == "" {
			errs = packer.MultiErrorAppend(errs, errors.New(fmt.Sprintf("disk format must be specified for pool type %q", c.Disks[idx].StoragePoolType)))
		}
	}
	if c.SCSIController == "" {
		log.Printf("SCSI controller not set, using default 'lsi'")
		c.SCSIController = "lsi"
	}

	errs = packer.MultiErrorAppend(errs, c.Comm.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.BootConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.HTTPConfig.Prepare(&c.ctx)...)

	// Required configurations that will display errors if not set
	if c.Username == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("username must be specified"))
	}
	if c.Password == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("password must be specified"))
	}
	if c.ProxmoxURLRaw == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("proxmox_url must be specified"))
	}
	if c.proxmoxURL, err = url.Parse(c.ProxmoxURLRaw); err != nil {
		errs = packer.MultiErrorAppend(errs, errors.New(fmt.Sprintf("Could not parse proxmox_url: %s", err)))
	}
	if c.ISOFile == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("iso_file must be specified"))
	}
	if c.Node == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("node must be specified"))
	}
	for idx := range c.NICs {
		if c.NICs[idx].Bridge == "" {
			errs = packer.MultiErrorAppend(errs, errors.New(fmt.Sprintf("network_adapters[%d].bridge must be specified", idx)))
		}
	}
	for idx := range c.Disks {
		if c.Disks[idx].StoragePool == "" {
			errs = packer.MultiErrorAppend(errs, errors.New(fmt.Sprintf("disks[%d].storage_pool must be specified", idx)))
		}
		if c.Disks[idx].StoragePoolType == "" {
			errs = packer.MultiErrorAppend(errs, errors.New(fmt.Sprintf("disks[%d].storage_pool_type must be specified", idx)))
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	packer.LogSecretFilter.Set(c.Password)
	return nil, nil
}

func contains(haystack []string, needle string) bool {
	for _, candidate := range haystack {
		if candidate == needle {
			return true
		}
	}
	return false
}
