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
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi/nfc"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

var sha = map[int]func() hash.Hash{
	1:   sha1.New,
	256: sha256.New,
	512: sha512.New,
}

type ExportConfig struct {
	// name of the ovf. defaults to the name of the VM
	Name string `mapstructure:"name"`
	// overwrites existing ovf
	Force bool `mapstructure:"force"`
	// Include image files (*.{iso,img})
	Images bool `mapstructure:"images"`
	// Generate manifest using SHA 1, 256, 512 or 0 to skip
	Sha       int          `mapstructure:"sha"`
	OutputDir OutputConfig `mapstructure:",squash"`
}

type StepExport struct {
	Name      string
	Force     bool
	Images    bool
	Sha       int
	OutputDir string
	dest      string
	mf        bytes.Buffer
}

func (s *StepExport) Cleanup(multistep.StateBag) {
}

func (s *StepExport) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	vm := state.Get("vm").(*driver.VirtualMachine)

	if s.Sha != 0 {
		if _, ok := sha[s.Sha]; !ok {
			state.Put("error", fmt.Errorf("unknown hash: sha%d", s.Sha))
			return multistep.ActionHalt
		}
	}

	if s.Name == "" {
		info, err := vm.Info("name")
		if err != nil {
			state.Put("error", fmt.Errorf("unable to look up vm name"))
			return multistep.ActionHalt
		}
		s.Name = info.Name
	}
	s.dest = filepath.Join(s.OutputDir, s.Name)

	target := filepath.Join(s.dest, s.Name+".ovf")

	if !s.Force {
		if _, err := os.Stat(target); err == nil {
			state.Put("error", fmt.Errorf("file already exists: %s", target))
			return multistep.ActionHalt
		}
	}

	if err := os.MkdirAll(s.dest, 0750); err != nil {
		state.Put("error", errors.Wrap(err, "unable to make directory"))
		return multistep.ActionHalt
	}

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

	for _, i := range info.Items {
		if !s.include(&i) {
			continue
		}

		if !strings.HasPrefix(i.Path, s.Name) {
			i.Path = s.Name + "-" + i.Path
		}

		err = s.Download(ctx, lease, i)
		if err != nil {
			state.Put("error", err)
			return multistep.ActionHalt
		}

		ui.Message("exporting file: " + i.File().Path)
		cdp.OvfFiles = append(cdp.OvfFiles, i.File())
	}

	if err = lease.Complete(ctx); err != nil {
		state.Put("error", errors.Wrap(err, "unable to complete lease"))
		return multistep.ActionHalt
	}

	desc, err := vm.CreateDescriptor(cdp)
	if err != nil {
		state.Put("error", errors.Wrap(err, "unable to create descriptor"))
		return multistep.ActionHalt
	}

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

	_, err = io.WriteString(w, desc.OvfDescriptor)
	if err != nil {
		state.Put("error", errors.Wrap(err, "unable to write descriptor"))
		return multistep.ActionHalt
	}

	if err = file.Close(); err != nil {
		state.Put("error", errors.Wrap(err, "unable to close descriptor"))
		return multistep.ActionHalt
	}

	if s.Sha == 0 {
		return multistep.ActionContinue
	}

	s.addHash(filepath.Base(target), h)

	file, err = os.Create(filepath.Join(s.dest, s.Name+".mf"))
	if err != nil {
		state.Put("error", err)
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

	return multistep.ActionContinue
}

func (s *StepExport) include(item *nfc.FileItem) bool {
	if s.Images {
		return true
	}

	return filepath.Ext(item.Path) == ".vmdk"
}

func (s *StepExport) newHash() (hash.Hash, bool) {
	if h, ok := sha[s.Sha]; ok {
		return h(), true
	}

	return nil, false
}

func (s *StepExport) addHash(p string, h hash.Hash) {
	_, _ = fmt.Fprintf(&s.mf, "SHA%d(%s)= %x\n", s.Sha, p, h.Sum(nil))
}

func (s *StepExport) Download(ctx context.Context, lease *nfc.Lease, item nfc.FileItem) error {
	path := filepath.Join(s.dest, item.Path)

	opts := soap.Download{}

	if h, ok := s.newHash(); ok {
		opts.Writer = h

		defer s.addHash(item.Path, h)
	}

	return lease.DownloadFile(ctx, path, item, opts)
}
