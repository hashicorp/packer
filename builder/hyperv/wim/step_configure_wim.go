package wim

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
	"github.com/hashicorp/packer/builder/hyperv/common/powershell"
	"github.com/mitchellh/mapstructure"
)

// This step configures a WIM offline.
//
type StepConfigureWIM struct {
	ImageIndex       uint32
	ImageName        string
	LogPath          string
	MountDir         string
	ScratchDir       string
	SkipOperation    bool
	WIMPathKey       string
	WindowsConfigUrl string

	mounted bool
}

type winConfig struct {
	ContentVersion string        `mapstructure:"contentVersion" required:"true"`
	AppXPackages   []appXPackage `mapstructure:"appXPackages, squash" required:"false"`
	Capabilities   []capability  `mapstructure:"capabilities, squash" required:"false"`
	Drivers        []driver      `mapstructure:"drivers, squash" required:"false"`
	Packages       []winPackage  `mapstructure:"packages, squash" required:"false"`
	Features       []feature     `mapstructure:"features, squash" required:"false"`
	ProductKey     string        `mapstructure:"productKey, omitempty" required:"false"`
	Unattend       string        `mapstructure:"unattend, omitempty" required:"false"`
}

type appXPackage struct {
	Action         string `mapstructure:"action" required:"true"`
	Path           string `mapstructure:"path" required:"false"`
	Name           string `mapstructure:"name" required:"false"`
	DependencyPath string `mapstructure:"dependencyPath" required:"false"`
	LicensePath    string `mapstructure:"licensePath" required:"false"`
}

type capability struct {
	Action string `mapstructure:"action" required:"true"`
	Name   string `mapstructure:"name" required:"true"`
	Path   string `mapstructure:"path" required:"false"`
}

type driver struct {
	Action string `mapstructure:"action" required:"true"`
	Path   string `mapstructure:"path" required:"true"`
}

type winPackage struct {
	Action string `mapstructure:"action" required:"true"`
	Path   string `mapstructure:"path" required:"true"`
}

type feature struct {
	Action string `mapstructure:"action" required:"true"`
	Name   string `mapstructure:"name" required:"true"`
}

func (s *StepConfigureWIM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if s.SkipOperation {
		return multistep.ActionContinue
	}

	ui := state.Get("ui").(packersdk.Ui)
	debug := state.Get("debug").(bool)
	wimPath := state.Get(s.WIMPathKey).(string)

	// If no WindowsConfigUrl is specified, return
	if s.WindowsConfigUrl == "" {
		return multistep.ActionContinue
	}

	var err error

	if s.MountDir == "" {
		if s.MountDir, err = tmp.Dir("mount"); err != nil {
			err = fmt.Errorf("Error creating mount directory: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	_, err = os.Stat(s.MountDir)
	if os.IsNotExist(err) {
		if err = os.Mkdir(s.MountDir, 0777); err != nil {
			err = fmt.Errorf("Error creating mount dir: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Open Windows configuration file
	jsonFile, err := os.Open(s.WindowsConfigUrl)
	if err != nil {
		err = fmt.Errorf("Error opening Windows configuration file: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	defer jsonFile.Close()

	// Read Windows configuration file into a byte array
	jsonBytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		err = fmt.Errorf("Error reading Windows configuration file: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Unmarshal the byte array
	var result map[string]interface{}
	if err = json.Unmarshal([]byte(jsonBytes), &result); err != nil {
		err = fmt.Errorf("Error unmarshaling Windows configuration: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Decode Windows configuration
	var config winConfig
	if err = mapstructure.Decode(result, &config); err != nil {
		err = fmt.Errorf("Error decoding Windows configuration: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if debug {
		ui.Say(fmt.Sprintf("%#v", config))
	}

	ui.Say("Mounting WIM...")

	log.Printf("Mount directory: %s", s.MountDir)

	// Mount WIM
	if err = s.mountWindowsImage(wimPath, true); err != nil {
		err = fmt.Errorf("Error mounting WIM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Remove AppX packages
	for _, appX := range config.AppXPackages {
		if appX.Action == "remove" {
			ui.Say(fmt.Sprintf("Removing AppX package: %s", filepath.ToSlash(appX.Name)))

			if err = s.removeAppxPackage(appX.Name); err != nil {
				err = fmt.Errorf("Error removing AppX package %s: %s", filepath.ToSlash(appX.Path), err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	// Remove capabilites
	for _, capability := range config.Capabilities {
		if capability.Action == "remove" {
			ui.Say(fmt.Sprintf("Removing capability: %s", capability.Name))

			if err = s.removeWindowsCapability(capability.Name); err != nil {
				err = fmt.Errorf("Error removing capability %s: %s", capability.Name, err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	// Remove drivers
	for _, driver := range config.Drivers {
		if driver.Action == "remove" {
			ui.Say(fmt.Sprintf("Removing driver: %s", filepath.ToSlash(driver.Path)))

			if err = s.removeWindowsDriver(driver.Path); err != nil {
				err = fmt.Errorf("Error removing driver %s: %s", filepath.ToSlash(driver.Path), err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	// Remove packages
	for _, winPackage := range config.Packages {
		if winPackage.Action == "remove" {
			ui.Say(fmt.Sprintf("Removing package: %s", filepath.ToSlash(winPackage.Path)))

			if err = s.removeWindowsPackageByPath(winPackage.Path); err != nil {
				err = fmt.Errorf("Error removing package %s: %s", filepath.ToSlash(winPackage.Path), err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	// Remove features
	for _, feature := range config.Features {
		if feature.Action == "disable" {
			ui.Say(fmt.Sprintf("Disabling feature: %s", feature.Name))

			if err = s.disableWindowsOptionalFeature(feature.Name); err != nil {
				err = fmt.Errorf("Error disabling feature %s: %s", feature.Name, err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	// Add AppX packages
	for _, appX := range config.AppXPackages {
		if appX.Action == "add" {
			ui.Say(fmt.Sprintf("Adding AppX package: %s", filepath.ToSlash(appX.Path)))

			if err = s.addAppxProvisionedPackage(appX.Path, appX.DependencyPath, appX.LicensePath); err != nil {
				err = fmt.Errorf("Error adding AppX package %s: %s", filepath.ToSlash(appX.Path), err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	// Add capabilities
	for _, capability := range config.Capabilities {
		if capability.Action == "add" {
			ui.Say(fmt.Sprintf("Adding capability: %s", capability.Name))

			if err = s.addWindowsCapability(capability.Name, capability.Path); err != nil {
				err = fmt.Errorf("Error adding capability %s: %s", capability.Name, err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	// Add drivers
	for _, driver := range config.Drivers {
		if driver.Action == "add" {
			ui.Say(fmt.Sprintf("Adding driver: %s", filepath.ToSlash(driver.Path)))

			if err = s.addWindowsDriver(driver.Path, true); err != nil {
				err = fmt.Errorf("Error adding driver %s: %s", filepath.ToSlash(driver.Path), err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	// Add packages
	for _, winPackage := range config.Packages {
		if winPackage.Action == "add" {
			ui.Say(fmt.Sprintf("Adding package: %s", filepath.ToSlash(winPackage.Path)))

			if err = s.addWindowsPackage(winPackage.Path); err != nil {
				err = fmt.Errorf("Error adding package %s: %s", filepath.ToSlash(winPackage.Path), err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	// Enable features
	for _, feature := range config.Features {
		if feature.Action == "enable" {
			ui.Say(fmt.Sprintf("Enabling feature: %s", feature.Name))

			if err = s.enableWindowsOptionalFeature(feature.Name, true); err != nil {
				err = fmt.Errorf("Error enabling feature %s: %s", feature.Name, err)
				state.Put("error", err)
				ui.Error(err.Error())
				return multistep.ActionHalt
			}
		}
	}

	// Set product key
	if config.ProductKey != "" {
		ui.Say(fmt.Sprintf("Setting product key..."))

		if err = s.setWindowsProductKey(config.ProductKey); err != nil {
			err = fmt.Errorf("Error setting product key: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Use Windows unattend
	if config.Unattend != "" {
		ui.Say(fmt.Sprintf("Using unattend..."))

		if err = s.useWindowsUnattend(config.Unattend); err != nil {
			err = fmt.Errorf("Error using unattend: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	ui.Say("Unmounting WIM...")

	// Unmount WIM
	if err = s.dismountWindowsImage(false); err != nil {
		err = fmt.Errorf("Error dismounting WIM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepConfigureWIM) Cleanup(state multistep.StateBag) {
	if s.SkipOperation {
		return
	}

	ui := state.Get("ui").(packersdk.Ui)

	if err := s.dismountWindowsImage(true); err != nil {
		err = fmt.Errorf("Error dismounting WIM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
	}

	if err := os.Remove(s.MountDir); err != nil {
		err = fmt.Errorf("Error removing mount dir: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
	}

	ui.Say(fmt.Sprintf("Removed mount dir in %s", s.MountDir))
}

// addAppxProvisionedPackage adds an app package (.appx) that will install for each new user to a Windows image
func (s *StepConfigureWIM) addAppxProvisionedPackage(packagePath string, dependencyPackagePath string, licensePath string) error {
	cmd := fmt.Sprintf("Add-AppxProvisionedPackage -Path \"%s\" -PackagePath \"%s\"", filepath.FromSlash(s.MountDir), filepath.FromSlash(packagePath))
	if dependencyPackagePath != "" {
		cmd = fmt.Sprintf("%s -DependencyPackagePath \"%s\"", cmd, filepath.FromSlash(dependencyPackagePath))
	}
	if licensePath != "" {
		cmd = fmt.Sprintf("%s -LicensePath \"%s\"", cmd, filepath.FromSlash(licensePath))
	}
	return s.execPSCmd(cmd)
}

// addWindowsCapability installs a Windows capability package on the specified operating system image
func (s *StepConfigureWIM) addWindowsCapability(name string, source string) error {
	cmd := fmt.Sprintf("Add-WindowsCapability -Path \"%s\" -Name \"%s\"", filepath.FromSlash(s.MountDir), name)
	if source != "" {
		cmd = fmt.Sprintf("%s -Source \"%s\"", cmd, filepath.FromSlash(source))
	}
	return s.execPSCmd(cmd)
}

// addWindowsDriver adds a driver to an offline Windows image
func (s *StepConfigureWIM) addWindowsDriver(driver string, recurse bool) error {
	cmd := fmt.Sprintf("Add-WindowsDriver -Path \"%s\" -Driver \"%s\"", filepath.FromSlash(s.MountDir), driver)
	if recurse {
		cmd = cmd + " -Recurse"
	}
	return s.execPSCmd(cmd)
}

// addWindowsPackage adds a single .cab or .msu file to a Windows image
func (s *StepConfigureWIM) addWindowsPackage(packagePath string) error {
	cmd := fmt.Sprintf("Add-WindowsPackage -Path \"%s\" -PackagePath \"%s\"", filepath.FromSlash(s.MountDir), filepath.FromSlash(packagePath))
	return s.execPSCmd(cmd)
}

// disableWindowsOptionalFeature disables a feature in a Windows image
func (s *StepConfigureWIM) disableWindowsOptionalFeature(name string) error {
	cmd := fmt.Sprintf("Disable-WindowsOptionalFeature -Path \"%s\" -FeatureName \"%s\"", filepath.FromSlash(s.MountDir), name)
	return s.execPSCmd(cmd)
}

// dismountWindowsImage dismounts a Windows image from the directory it is mapped to
func (s *StepConfigureWIM) dismountWindowsImage(discard bool) error {
	if s.mounted {
		cmd := fmt.Sprintf("Dismount-WindowsImage -Path \"%s\"", filepath.FromSlash(s.MountDir))
		if discard {
			cmd = cmd + " -Discard"
		} else {
			cmd = cmd + " -Save"
		}

		err := s.execPSCmd(cmd)
		if err == nil {
			s.mounted = false
		}
		return err
	}
	return nil
}

// enableWindowsOptionalFeature enables a feature in a Windows image
func (s *StepConfigureWIM) enableWindowsOptionalFeature(featureName string, all bool) error {
	cmd := fmt.Sprintf("Enable-WindowsOptionalFeature -Path \"%s\" -FeatureName \"%s\"", filepath.FromSlash(s.MountDir), featureName)
	if all {
		cmd = cmd + " -All"
	}
	return s.execPSCmd(cmd)
}

func (s *StepConfigureWIM) execPSCmd(cmd string) error {
	if s.LogPath != "" {
		cmd = fmt.Sprintf("%s -LogPath \"%s\"", cmd, filepath.FromSlash(s.LogPath))
	}
	if s.ScratchDir != "" {
		cmd = fmt.Sprintf("%s -ScratchDirectory \"%s\"", cmd, filepath.FromSlash(s.ScratchDir))
	}

	var ps powershell.PowerShellCmd
	return ps.Run(cmd)
}

// mountWindowsImage mounts a Windows image in a WIM to a directory on the local computer
func (s *StepConfigureWIM) mountWindowsImage(wimPath string, optimize bool) error {
	var cmd string
	if s.ImageIndex > 0 {
		cmd = fmt.Sprintf("Mount-WindowsImage -Path \"%s\" -ImagePath \"%s\" -Index %d", filepath.FromSlash(s.MountDir), filepath.FromSlash(wimPath), s.ImageIndex)
	} else {
		cmd = fmt.Sprintf("Mount-WindowsImage -Path \"%s\" -ImagePath \"%s\" -Name \"%s\"", filepath.FromSlash(s.MountDir), filepath.FromSlash(wimPath), s.ImageName)
	}
	if optimize {
		cmd = cmd + " -Optimize"
	}

	err := s.execPSCmd(cmd)
	if err == nil {
		s.mounted = true
	}
	return err
}

// removeAppxPackage removes an app package from one or more user accounts
func (s *StepConfigureWIM) removeAppxPackage(pkg string) error {
	cmd := fmt.Sprintf("Remove-AppxPackage -Path \"%s\" -Package \"%s\"", filepath.FromSlash(s.MountDir), pkg)
	return s.execPSCmd(cmd)
}

// removeWindowsCapability uninstalls a Windows capability package from an image
func (s *StepConfigureWIM) removeWindowsCapability(name string) error {
	cmd := fmt.Sprintf("Remove-WindowsCapability -Path \"%s\" -Name \"%s\"", filepath.FromSlash(s.MountDir), name)
	return s.execPSCmd(cmd)
}

// removeWindowsDriver removes a driver from an offline Windows image
func (s *StepConfigureWIM) removeWindowsDriver(driver string) error {
	cmd := fmt.Sprintf("Remove-WindowsDriver -Path \"%s\" -Driver \"%s\"", filepath.FromSlash(s.MountDir), driver)
	return s.execPSCmd(cmd)
}

// removeWindowsPackageByPath removes a package from a Windows image by package path
func (s *StepConfigureWIM) removeWindowsPackageByPath(path string) error {
	cmd := fmt.Sprintf("Remove-WindowsPackage -Path \"%s\" -PackagePath \"%s\"", filepath.FromSlash(s.MountDir), filepath.FromSlash(path))
	return s.execPSCmd(cmd)
}

// setWindowsProductKey sets the product key for the Windows image
func (s *StepConfigureWIM) setWindowsProductKey(productKey string) error {
	cmd := fmt.Sprintf("Set-WindowsProductKey -Path \"%s\" -ProductKey \"%s\"", filepath.FromSlash(s.MountDir), productKey)
	return s.execPSCmd(cmd)
}

// useWindowsUnattend applies an unattended answer file to a Windows image
func (s *StepConfigureWIM) useWindowsUnattend(unattendPath string) error {
	cmd := fmt.Sprintf("Use-WindowsUnattend -Path \"%s\" -UnattendPath \"%s\"", filepath.FromSlash(s.MountDir), filepath.FromSlash(unattendPath))
	return s.execPSCmd(cmd)
}
