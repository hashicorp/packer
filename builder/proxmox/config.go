//go:generate mapstructure-to-hcl2 -type Config,nicConfig,diskConfig,vgaConfig,storageConfig

package proxmox

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
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
	common.ISOConfig       `mapstructure:",squash"`
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
	VGA            vgaConfig    `mapstructure:"vga"`
	NICs           []nicConfig  `mapstructure:"network_adapters"`
	Disks          []diskConfig `mapstructure:"disks"`
	ISOFile        string       `mapstructure:"iso_file"`
	ISOStoragePool string       `mapstructure:"iso_storage_pool"`
	Agent          bool         `mapstructure:"qemu_agent"`
	SCSIController string       `mapstructure:"scsi_controller"`
	Onboot         bool         `mapstructure:"onboot"`
	DisableKVM     bool         `mapstructure:"disable_kvm"`

	TemplateName        string `mapstructure:"template_name"`
	TemplateDescription string `mapstructure:"template_description"`
	UnmountISO          bool   `mapstructure:"unmount_iso"`

	CloudInit            bool   `mapstructure:"cloud_init"`
	CloudInitStoragePool string `mapstructure:"cloud_init_storage_pool"`

	shouldUploadISO bool

	AdditionalISOFiles []storageConfig `mapstructure:"additional_iso_files"`

	ctx interpolate.Context
}

type nicConfig struct {
	Model        string `mapstructure:"model"`
	PacketQueues int    `mapstructure:"packet_queues"`
	MACAddress   string `mapstructure:"mac_address"`
	Bridge       string `mapstructure:"bridge"`
	VLANTag      string `mapstructure:"vlan_tag"`
	Firewall     bool   `mapstructure:"firewall"`
}
type diskConfig struct {
	Type            string `mapstructure:"type"`
	StoragePool     string `mapstructure:"storage_pool"`
	StoragePoolType string `mapstructure:"storage_pool_type"`
	Size            string `mapstructure:"disk_size"`
	CacheMode       string `mapstructure:"cache_mode"`
	DiskFormat      string `mapstructure:"format"`
}
type vgaConfig struct {
	Type   string `mapstructure:"type"`
	Memory int    `mapstructure:"memory"`
}
type storageConfig struct {
	Device    string `mapstructure:"device"`
	BusNumber int    `mapstructure:"bus_number"`
	Filename  string `mapstructure:"filename"`
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	// Agent defaults to true
	c.Agent = true
	// Do not add a cloud-init cdrom by default
	c.CloudInit = false

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
	warnings := make([]string, 0)

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
		// For any storage pool types which aren't in rxStorageTypes in proxmox-api/proxmox/config_qemu.go:890
		// (currently zfspool|lvm|rbd|cephfs), the format parameter is mandatory. Make sure this is still up to date
		// when updating the vendored code!
		if !contains([]string{"zfspool", "lvm", "rbd", "cephfs"}, c.Disks[idx].StoragePoolType) && c.Disks[idx].DiskFormat == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("disk format must be specified for pool type %q", c.Disks[idx].StoragePoolType))
		}
	}
	for idx := range c.AdditionalISOFiles {
		if c.AdditionalISOFiles[idx].Device == "" {
			log.Printf("AdditionalISOFile %d Device not set, using default 'ide'", idx)
			c.AdditionalISOFiles[idx].Device = "ide"
		}
		if !contains([]string{"ide", "sata", "scsi"}, c.AdditionalISOFiles[idx].Device) {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("%q is not a valid AdditionalISOFile Device", c.AdditionalISOFiles[idx]))
		}
		if c.AdditionalISOFiles[idx].BusNumber == 0 {
			log.Printf("AdditionalISOFile %d number not set, using default: '3'", idx)
			c.AdditionalISOFiles[idx].BusNumber = 3
		}
		if c.AdditionalISOFiles[idx].Device == "ide" && c.AdditionalISOFiles[idx].BusNumber == 2 {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("IDE bus 2 is used by boot ISO"))
		}
		if c.AdditionalISOFiles[idx].Device == "ide" && c.AdditionalISOFiles[idx].BusNumber > 3 {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("IDE bus number can't be higher than 3"))
		}
		if c.AdditionalISOFiles[idx].Device == "sata" && c.AdditionalISOFiles[idx].BusNumber > 5 {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("SATA bus number can't be higher than 5"))
		}
		if c.AdditionalISOFiles[idx].Device == "scsi" && c.AdditionalISOFiles[idx].BusNumber > 30 {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("SCSI bus number can't be higher than 30"))
		}
	}
	if c.SCSIController == "" {
		log.Printf("SCSI controller not set, using default 'lsi'")
		c.SCSIController = "lsi"
	}

	errs = packer.MultiErrorAppend(errs, c.Comm.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.BootConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.HTTPConfig.Prepare(&c.ctx)...)

	// Check ISO config
	// Either a pre-uploaded ISO should be referenced in iso_file, OR a URL
	// (possibly to a local file) to an ISO file that will be downloaded and
	// then uploaded to Proxmox.
	if c.ISOFile != "" {
		c.shouldUploadISO = false
	} else {
		isoWarnings, isoErrors := c.ISOConfig.Prepare(&c.ctx)
		errs = packer.MultiErrorAppend(errs, isoErrors...)
		warnings = append(warnings, isoWarnings...)
		c.shouldUploadISO = true
	}

	if (c.ISOFile == "" && len(c.ISOConfig.ISOUrls) == 0) || (c.ISOFile != "" && len(c.ISOConfig.ISOUrls) != 0) {
		errs = packer.MultiErrorAppend(errs, errors.New("either iso_file or iso_url, but not both, must be specified"))
	}
	if len(c.ISOConfig.ISOUrls) != 0 && c.ISOStoragePool == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("when specifying iso_url, iso_storage_pool must also be specified"))
	}

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
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Could not parse proxmox_url: %s", err))
	}
	if c.Node == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("node must be specified"))
	}
	if strings.ContainsAny(c.TemplateName, " ") {
		errs = packer.MultiErrorAppend(errs, errors.New("template_name must not contain spaces"))
	}
	for idx := range c.NICs {
		if c.NICs[idx].Bridge == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("network_adapters[%d].bridge must be specified", idx))
		}
		if c.NICs[idx].Model != "virtio" && c.NICs[idx].PacketQueues > 0 {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("network_adapters[%d].packet_queues can only be set for 'virtio' driver", idx))
		}
	}
	for idx := range c.Disks {
		if c.Disks[idx].StoragePool == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("disks[%d].storage_pool must be specified", idx))
		}
		if c.Disks[idx].StoragePoolType == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("disks[%d].storage_pool_type must be specified", idx))
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
