//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type ExportConfig

package common

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi/nfc"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

// You may optionally export an ovf from VSphere to the instance running Packer.
//
// Example usage:
//
// In JSON:
// ```json
// ...
//   "vm_name": "example-ubuntu",
// ...
//   "export": {
//     "force": true,
//     "output_directory": "./output_vsphere"
//   },
// ```
// In HCL2:
// ```hcl
//   # ...
//   vm_name = "example-ubuntu"
//   # ...
//   export {
//     force = true
//     output_directory = "./output_vsphere"
//   }
// ```
// The above configuration would create the following files:
//
// ```text
// ./output_vsphere/example-ubuntu-disk-0.vmdk
// ./output_vsphere/example-ubuntu.mf
// ./output_vsphere/example-ubuntu.ovf
// ```
type ExportConfig struct {
	// name of the ovf. defaults to the name of the VM
	Name string `mapstructure:"name"`
	// overwrite ovf if it exists
	Force bool `mapstructure:"force"`
	// include iso and img image files that are attached to the VM
	Images bool `mapstructure:"images"`
	// generate manifest using sha1, sha256, sha512. Defaults to 'sha256'. Use 'none' for no manifest.
	Manifest string `mapstructure:"manifest"`
	// Directory on the computer running Packer to export files to
	OutputDir OutputConfig `mapstructure:",squash"`
	// Advanced ovf export options. Options can include:
	// * mac - MAC address is exported for all ethernet devices
	// * uuid - UUID is exported for all virtual machines
	// * extraconfig - all extra configuration options are exported for a virtual machine
	// * nodevicesubtypes - resource subtypes for CD/DVD drives, floppy drives, and serial and parallel ports are not exported
	//
	// For example, adding the following export config option would output the mac addresses for all Ethernet devices in the ovf file:
	//
	// In JSON:
	// ```json
	// ...
	//   "export": {
	//     "options": ["mac"]
	//   },
	// ```
	// In HCL2:
	// ```hcl
	// ...
	//   export {
	//     options = ["mac"]
	//   }
	// ```
	Options []string `mapstructure:"options"`
}

var sha = map[string]func() hash.Hash{
	"none":   nil,
	"sha1":   sha1.New,
	"sha256": sha256.New,
	"sha512": sha512.New,
}

func (c *ExportConfig) Prepare(ctx *interpolate.Context, lc *LocationConfig, pc *common.PackerConfig) []error {
	var errs *packersdk.MultiError

	errs = packersdk.MultiErrorAppend(errs, c.OutputDir.Prepare(ctx, pc)...)

	// manifest should default to sha256
	if c.Manifest == "" {
		c.Manifest = "sha256"
	}
	if _, ok := sha[c.Manifest]; !ok {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("unknown hash: %s. available options include available options being 'none', 'sha1', 'sha256', 'sha512'", c.Manifest))
	}

	if c.Name == "" {
		c.Name = lc.VMName
	}
	target := getTarget(c.OutputDir.OutputDir, c.Name)
	if !c.Force {
		if _, err := os.Stat(target); err == nil {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("file already exists: %s", target))
		}
	}

	if err := os.MkdirAll(c.OutputDir.OutputDir, c.OutputDir.DirPerm); err != nil {
		errs = packersdk.MultiErrorAppend(errs, errors.Wrap(err, "unable to make directory for export"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs.Errors
	}

	return nil
}

func getTarget(dir string, name string) string {
	return filepath.Join(dir, name+".ovf")
}

type StepExport struct {
	Name      string
	Force     bool
	Images    bool
	Manifest  string
	OutputDir string
	Options   []string
	mf        bytes.Buffer
}

func (s *StepExport) Cleanup(multistep.StateBag) {
}

func (s *StepExport) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(*driver.VirtualMachineDriver)

	ui.Message("Starting export...")
	lease, err := vm.Export()
	if err != nil {
		state.Put("error", errors.Wrap(err, "error exporting vm"))
		return multistep.ActionHalt
	}

	info, err := lease.Wait(ctx, nil)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	u := lease.StartUpdater(ctx, info)
	defer u.Done()

	cdp := types.OvfCreateDescriptorParams{
		Name: s.Name,
	}

	m := vm.NewOvfManager()
	if len(s.Options) > 0 {
		exportOptions, err := vm.GetOvfExportOptions(m)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}
		var unknown []string
		for _, option := range s.Options {
			found := false
			for _, exportOpt := range exportOptions {
				if exportOpt.Option == option {
					found = true
					break
				}
			}
			if !found {
				unknown = append(unknown, option)
			}
			cdp.ExportOption = append(cdp.ExportOption, option)
		}

		// only printing error message because the unknown options are just ignored by vcenter
		if len(unknown) > 0 {
			ui.Error(fmt.Sprintf("unknown export options %s", strings.Join(unknown, ",")))
		}
	}

	for _, i := range info.Items {
		if !s.include(&i) {
			continue
		}

		if !strings.HasPrefix(i.Path, s.Name) {
			i.Path = s.Name + "-" + i.Path
		}

		file := i.File()

		ui.Message("Downloading: " + file.Path)
		size, err := s.Download(ctx, lease, i)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}

		// Fix file size descriptor
		file.Size = size

		ui.Message("Exporting file: " + file.Path)
		cdp.OvfFiles = append(cdp.OvfFiles, file)
	}

	if err = lease.Complete(ctx); err != nil {
		state.Put("error", errors.Wrap(err, "unable to complete lease"))
		return multistep.ActionHalt
	}

	desc, err := vm.CreateDescriptor(m, cdp)
	if err != nil {
		state.Put("error", errors.Wrap(err, "unable to create descriptor"))
		return multistep.ActionHalt
	}

	target := getTarget(s.OutputDir, s.Name)
	file, err := os.Create(target)
	if err != nil {
		state.Put("error", errors.Wrap(err, "unable to create file: "+target))
		return multistep.ActionHalt
	}

	var w io.Writer = file
	h, ok := s.newHash()
	if ok {
		w = io.MultiWriter(file, h)
	}

	ui.Message("Writing ovf...")
	_, err = io.WriteString(w, desc.OvfDescriptor)
	if err != nil {
		state.Put("error", errors.Wrap(err, "unable to write descriptor"))
		return multistep.ActionHalt
	}

	if err = file.Close(); err != nil {
		state.Put("error", errors.Wrap(err, "unable to close descriptor"))
		return multistep.ActionHalt
	}

	if s.Manifest == "none" {
		// manifest does not need to be created, return
		return multistep.ActionContinue
	}

	ui.Message("Creating manifest...")
	s.addHash(filepath.Base(target), h)

	file, err = os.Create(filepath.Join(s.OutputDir, s.Name+".mf"))
	if err != nil {
		state.Put("error", errors.Wrap(err, "unable to create manifest"))
		return multistep.ActionHalt
	}

	_, err = io.Copy(file, &s.mf)
	if err != nil {
		state.Put("error", errors.Wrap(err, "unable to write manifest"))
		return multistep.ActionHalt
	}

	err = file.Close()
	if err != nil {
		state.Put("error", errors.Wrap(err, "unable to close file"))
		return multistep.ActionHalt
	}

	ui.Message("Finished exporting...")
	return multistep.ActionContinue
}

func (s *StepExport) include(item *nfc.FileItem) bool {
	if s.Images {
		return true
	}

	return filepath.Ext(item.Path) == ".vmdk"
}

func (s *StepExport) newHash() (hash.Hash, bool) {
	// check if function is nil to handle the 'none' case
	if h, ok := sha[s.Manifest]; ok && h != nil {
		return h(), true
	}

	return nil, false
}

func (s *StepExport) addHash(p string, h hash.Hash) {
	_, _ = fmt.Fprintf(&s.mf, "%s(%s)= %x\n", strings.ToUpper(s.Manifest), p, h.Sum(nil))
}

func (s *StepExport) Download(ctx context.Context, lease *nfc.Lease, item nfc.FileItem) (int64, error) {
	path := filepath.Join(s.OutputDir, item.Path)
	opts := soap.Download{}

	if h, ok := s.newHash(); ok {
		opts.Writer = h
		defer s.addHash(item.Path, h)
	}

	err := lease.DownloadFile(ctx, path, item, opts)
	if err != nil {
		return 0, err
	}

	f, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return f.Size(), err
}
