package virtualbox

import (
	"errors"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/builder/common"
	"github.com/mitchellh/packer/packer"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
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
	BootCommand          []string      `mapstructure:"boot_command"`
	BootWait             time.Duration ``
	DiskSize             uint          `mapstructure:"disk_size"`
	FloppyFiles          []string      `mapstructure:"floppy_files"`
	GuestAdditionsPath   string        `mapstructure:"guest_additions_path"`
	GuestAdditionsURL    string        `mapstructure:"guest_additions_url"`
	GuestAdditionsSHA256 string        `mapstructure:"guest_additions_sha256"`
	GuestOSType          string        `mapstructure:"guest_os_type"`
	Headless             bool          `mapstructure:"headless"`
	HTTPDir              string        `mapstructure:"http_directory"`
	HTTPPortMin          uint          `mapstructure:"http_port_min"`
	HTTPPortMax          uint          `mapstructure:"http_port_max"`
	ISOMD5               string        `mapstructure:"iso_md5"`
	ISOUrl               string        `mapstructure:"iso_url"`
	OutputDir            string        `mapstructure:"output_directory"`
	ShutdownCommand      string        `mapstructure:"shutdown_command"`
	ShutdownTimeout      time.Duration ``
	SSHHostPortMin       uint          `mapstructure:"ssh_host_port_min"`
	SSHHostPortMax       uint          `mapstructure:"ssh_host_port_max"`
	SSHPassword          string        `mapstructure:"ssh_password"`
	SSHPort              uint          `mapstructure:"ssh_port"`
	SSHUser              string        `mapstructure:"ssh_username"`
	SSHWaitTimeout       time.Duration ``
	VBoxVersionFile      string        `mapstructure:"virtualbox_version_file"`
	VBoxManage           [][]string    `mapstructure:"vboxmanage"`
	VMName               string        `mapstructure:"vm_name"`

	PackerBuildName string `mapstructure:"packer_build_name"`
	PackerDebug     bool   `mapstructure:"packer_debug"`

	RawBootWait        string `mapstructure:"boot_wait"`
	RawShutdownTimeout string `mapstructure:"shutdown_timeout"`
	RawSSHWaitTimeout  string `mapstructure:"ssh_wait_timeout"`
}

func (b *Builder) Prepare(raws ...interface{}) error {
	var err error

	for _, raw := range raws {
		err := mapstructure.Decode(raw, &b.config)
		if err != nil {
			return err
		}
	}

	if b.config.DiskSize == 0 {
		b.config.DiskSize = 40000
	}

	if b.config.FloppyFiles == nil {
		b.config.FloppyFiles = make([]string, 0)
	}

	if b.config.GuestAdditionsPath == "" {
		b.config.GuestAdditionsPath = "VBoxGuestAdditions.iso"
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
		b.config.OutputDir = fmt.Sprintf("output-%s", b.config.PackerBuildName)
	}

	if b.config.RawBootWait == "" {
		b.config.RawBootWait = "10s"
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

	if b.config.VBoxManage == nil {
		b.config.VBoxManage = make([][]string, 0)
	}

	if b.config.VBoxVersionFile == "" {
		b.config.VBoxVersionFile = ".vbox_version"
	}

	if b.config.VMName == "" {
		b.config.VMName = fmt.Sprintf("packer-%s", b.config.PackerBuildName)
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
				if _, err := os.Stat(url.Path); err != nil {
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

	if b.config.GuestAdditionsSHA256 != "" {
		b.config.GuestAdditionsSHA256 = strings.ToLower(b.config.GuestAdditionsSHA256)
	}

	if b.config.GuestAdditionsURL != "" {
		url, err := url.Parse(b.config.GuestAdditionsURL)
		if err != nil {
			errs = append(errs, fmt.Errorf("guest_additions_url is not a valid URL: %s", err))
		} else {
			if url.Scheme == "" {
				url.Scheme = "file"
			}

			if url.Scheme == "file" {
				if _, err := os.Stat(url.Path); err != nil {
					errs = append(errs, fmt.Errorf("guest_additions_url points to bad file: %s", err))
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
					errs = append(errs, fmt.Errorf("Unsupported URL scheme in guest_additions_url: %s", scheme))
				}
			}
		}

		if len(errs) == 0 {
			// Put the URL back together since we may have modified it
			b.config.GuestAdditionsURL = url.String()
		}
	}

	if _, err := os.Stat(b.config.OutputDir); err == nil {
		errs = append(errs, errors.New("Output directory already exists. It must not exist."))
	}

	b.config.BootWait, err = time.ParseDuration(b.config.RawBootWait)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing boot_wait: %s", err))
	}

	if b.config.RawShutdownTimeout == "" {
		b.config.RawShutdownTimeout = "5m"
	}

	if b.config.RawSSHWaitTimeout == "" {
		b.config.RawSSHWaitTimeout = "20m"
	}

	b.config.ShutdownTimeout, err = time.ParseDuration(b.config.RawShutdownTimeout)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing shutdown_timeout: %s", err))
	}

	if b.config.SSHHostPortMin > b.config.SSHHostPortMax {
		errs = append(errs, errors.New("ssh_host_port_min must be less than ssh_host_port_max"))
	}

	if b.config.SSHUser == "" {
		errs = append(errs, errors.New("An ssh_username must be specified."))
	}

	b.config.SSHWaitTimeout, err = time.ParseDuration(b.config.RawSSHWaitTimeout)
	if err != nil {
		errs = append(errs, fmt.Errorf("Failed parsing ssh_wait_timeout: %s", err))
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

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	steps := []multistep.Step{
		new(stepDownloadGuestAdditions),
		new(stepDownloadISO),
		new(stepPrepareOutputDir),
		&common.StepCreateFloppy{
			Files: b.config.FloppyFiles,
		},
		new(stepHTTPServer),
		new(stepSuppressMessages),
		new(stepCreateVM),
		new(stepCreateDisk),
		new(stepAttachISO),
		new(stepAttachFloppy),
		new(stepForwardSSH),
		new(stepVBoxManage),
		new(stepRun),
		new(stepTypeBootCommand),
		new(stepWaitForSSH),
		new(stepUploadVersion),
		new(stepUploadGuestAdditions),
		new(stepProvision),
		new(stepShutdown),
		new(stepExport),
	}

	// Setup the state bag
	state := make(map[string]interface{})
	state["cache"] = cache
	state["config"] = &b.config
	state["driver"] = b.driver
	state["hook"] = hook
	state["ui"] = ui

	// Run
	if b.config.PackerDebug {
		b.runner = &multistep.DebugRunner{
			Steps:   steps,
			PauseFn: common.MultistepDebugFn(ui),
		}
	} else {
		b.runner = &multistep.BasicRunner{Steps: steps}
	}

	b.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state["error"]; ok {
		return nil, rawErr.(error)
	}

	// If we were interrupted or cancelled, then just exit.
	if _, ok := state[multistep.StateCancelled]; ok {
		return nil, errors.New("Build was cancelled.")
	}

	if _, ok := state[multistep.StateHalted]; ok {
		return nil, errors.New("Build was halted.")
	}

	// Compile the artifact list
	files := make([]string, 0, 5)
	visit := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}

		return err
	}

	if err := filepath.Walk(b.config.OutputDir, visit); err != nil {
		return nil, err
	}

	artifact := &Artifact{
		dir: b.config.OutputDir,
		f:   files,
	}

	return artifact, nil
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
