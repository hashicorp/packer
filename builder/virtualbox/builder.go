package virtualbox

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

const BuilderId = "mitchellh.virtualbox"

type Builder struct {
	config config
	driver Driver
	runner multistep.Runner
}

type config struct {
	BootCommand    []string      `mapstructure:"boot_command"`
	BootWait       time.Duration ``
	GuestOSType    string        `mapstructure:"guest_os_type"`
	HTTPDir        string        `mapstructure:"http_directory"`
	HTTPPortMin    uint          `mapstructure:"http_port_min"`
	HTTPPortMax    uint          `mapstructure:"http_port_max"`
	ISOMD5         string        `mapstructure:"iso_md5"`
	ISOUrl         string        `mapstructure:"iso_url"`
	OutputDir      string        `mapstructure:"output_directory"`
	SSHHostPortMin uint          `mapstructure:"ssh_host_port_min"`
	SSHHostPortMax uint          `mapstructure:"ssh_host_port_max"`
	SSHPort        uint          `mapstructure:"ssh_port"`
	VMName         string        `mapstructure:"vm_name"`

	RawBootWait string `mapstructure:"boot_wait"`
}

func (b *Builder) Prepare(raw interface{}) error {
	var err error
	if err := mapstructure.Decode(raw, &b.config); err != nil {
		return err
	}

	if b.config.GuestOSType == "" {
		b.config.GuestOSType = "Other"
	}

	if b.config.HTTPPortMin == 0 {
		b.config.HTTPPortMin = 8000
	}

	if b.config.HTTPPortMax == 0 {
		b.config.HTTPPortMax = 9000
	}

	if b.config.OutputDir == "" {
		b.config.OutputDir = "virtualbox"
	}

	if b.config.SSHHostPortMin == 0 {
		b.config.SSHHostPortMin = 2222
	}

	if b.config.SSHHostPortMax == 0 {
		b.config.SSHHostPortMax = 4444
	}

	if b.config.SSHPort == 0 {
		b.config.SSHPort = 22
	}

	if b.config.VMName == "" {
		b.config.VMName = "packer"
	}

	errs := make([]error, 0)

	if b.config.HTTPPortMin > b.config.HTTPPortMax {
		errs = append(errs, errors.New("http_port_min must be less than http_port_max"))
	}

	if b.config.ISOMD5 == "" {
		errs = append(errs, errors.New("Due to large file sizes, an iso_md5 is required"))
	} else {
		b.config.ISOMD5 = strings.ToLower(b.config.ISOMD5)
	}

	if b.config.ISOUrl == "" {
		errs = append(errs, errors.New("An iso_url must be specified."))
	} else {
		url, err := url.Parse(b.config.ISOUrl)
		if err != nil {
			errs = append(errs, fmt.Errorf("iso_url is not a valid URL: %s", err))
		} else {
			if url.Scheme == "" {
				url.Scheme = "file"
			}

			if url.Scheme == "file" {
				if _, err := os.Stat(b.config.ISOUrl); err != nil {
					errs = append(errs, fmt.Errorf("iso_url points to bad file: %s", err))
				}
			} else {
				supportedSchemes := []string{"file", "http", "https"}
				scheme := strings.ToLower(url.Scheme)

				found := false
				for _, supported := range supportedSchemes {
					if scheme == supported {
						found = true
						break
					}
				}

				if !found {
					errs = append(errs, fmt.Errorf("Unsupported URL scheme in iso_url: %s", scheme))
				}
			}
		}

		if len(errs) == 0 {
			// Put the URL back together since we may have modified it
			b.config.ISOUrl = url.String()
		}
	}

	if b.config.RawBootWait != "" {
		b.config.BootWait, err = time.ParseDuration(b.config.RawBootWait)
		if err != nil {
			errs = append(errs, fmt.Errorf("Failed parsing boot_wait: %s", err))
		}
	}

	if b.config.SSHHostPortMin > b.config.SSHHostPortMax {
		errs = append(errs, errors.New("ssh_host_port_min must be less than ssh_host_port_max"))
	}

	b.driver, err = b.newDriver()
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed creating VirtualBox driver: %s", err))
	}

	if len(errs) > 0 {
		return &packer.MultiError{errs}
	}

	return nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) packer.Artifact {
	steps := []multistep.Step{
		new(stepDownloadISO),
		new(stepPrepareOutputDir),
		new(stepHTTPServer),
		new(stepSuppressMessages),
		new(stepCreateVM),
		new(stepCreateDisk),
		new(stepAttachISO),
		new(stepForwardSSH),
		new(stepRun),
		new(stepTypeBootCommand),
	}

	// Setup the state bag
	state := make(map[string]interface{})
	state["cache"] = cache
	state["config"] = &b.config
	state["driver"] = b.driver
	state["hook"] = hook
	state["ui"] = ui

	// Run
	b.runner = &multistep.BasicRunner{Steps: steps}
	b.runner.Run(state)

	return nil
}

func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}

func (b *Builder) newDriver() (Driver, error) {
	vboxmanagePath, err := exec.LookPath("VBoxManage")
	if err != nil {
		return nil, err
	}

	log.Printf("VBoxManage path: %s", vboxmanagePath)
	driver := &VBox42Driver{vboxmanagePath}
	if err := driver.Verify(); err != nil {
		return nil, err
	}

	return driver, nil
}
